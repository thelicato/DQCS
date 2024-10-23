//go:build linux
// +build linux

package socket

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/thelicato/dqcs/pkg/utils"
	"golang.design/x/clipboard"
)

var connClosed = make(chan bool)

func runAsHost(socketPath string) {
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Printf("Error starting UNIX socket listener: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Waiting for guest to connect...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Guest connected.")

		go handleConnection(conn)
		// Wait for the connection to close before accepting another
		<-connClosed
		fmt.Println("Connection closed, cleaning up and closing")
		break
	}

	listener.Close()
	os.Remove(socketPath)
}

func openGuestConnection() (io.ReadWriteCloser, error) {
	file, err := os.OpenFile(utils.LinuxVirtioPortPath, os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("error opening virtio-serial port: %v", err)
	}
	return file, nil
}

func runAsGuest() {
	fmt.Println("Running Guest component....")

	conn, err := openGuestConnection()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening virtio port: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Guest connection opened")
	handleConnection(conn)
}

func handleConnection(conn io.ReadWriteCloser) {
	defer conn.Close()

	var wg sync.WaitGroup

	// Initialize the clipboard package
	err := clipboard.Init()
	if err != nil {
		fmt.Println("Error initializing clipboard:", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start clipboard monitoring
	clipboardCh := clipboard.Watch(ctx, clipboard.FmtText)

	// Shared variable to keep track of the last clipboard content received from the connection
	var lastReceivedClipboard []byte

	wg.Add(2)
	go writeToConnection(conn, clipboardCh, &lastReceivedClipboard, &wg)
	go readFromConnection(conn, &lastReceivedClipboard, &wg)
	wg.Wait()
}

func readFromConnection(conn io.Reader, lastReceivedClipboard *[]byte, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		// Read 4-byte length header
		header := make([]byte, 4)
		_, err := io.ReadFull(conn, header)
		if err != nil {
			fmt.Println("Some error")
			if err == io.EOF {
				fmt.Println("Client disconnected")
				connClosed <- true
			}
			if err != io.EOF {
				fmt.Println("Error reading header from connection:", err)
			}
			return
		}
		length := binary.LittleEndian.Uint32(header)
		if length == 0 {
			continue
		}
		// Read the data
		data := make([]byte, length)
		_, err = io.ReadFull(conn, data)
		if err != nil {
			fmt.Println("Error reading data from connection:", err)
			return
		}

		*lastReceivedClipboard = data
		clipboard.Write(clipboard.FmtText, data)
	}
}

func writeToConnection(conn io.Writer, clipboardCh <-chan []byte, lastReceivedClipboard *[]byte, wg *sync.WaitGroup) {
	defer wg.Done()
	for data := range clipboardCh {
		if bytes.Equal(data, *lastReceivedClipboard) {
			continue
		}

		// Send data length as 4-byte big-endian
		length := uint32(len(data))
		header := make([]byte, 4)
		binary.LittleEndian.PutUint32(header, length)

		// Send header
		_, err := conn.Write(header)
		if err != nil {
			fmt.Println("Error writing header to connection:", err)
			return
		}
		// Send data
		_, err = conn.Write(data)
		if err != nil {
			fmt.Println("Error writing data to connection:", err)
			return
		}
	}
}
