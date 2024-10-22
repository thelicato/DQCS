package utils

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"

	"golang.design/x/clipboard"
)

func HandleConnection(conn io.ReadWriteCloser) {
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

	wg.Add(2)
	go readFromConnection(conn, &wg)
	go writeToConnection(conn, clipboardCh, &wg)

	wg.Wait()
}

func readFromConnection(conn io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		// Read 4-byte length header
		header := make([]byte, 4)
		_, err := io.ReadFull(conn, header)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading header from connection:", err)
			}
			return
		}
		length := binary.BigEndian.Uint32(header)
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
		currentClipboard := clipboard.Read(clipboard.FmtText)
		if !equalBytes(data, currentClipboard) {
			clipboard.Write(clipboard.FmtText, data)
		}
	}
}

func writeToConnection(conn io.Writer, clipboardCh <-chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	for data := range clipboardCh {
		// Send data length as 4-byte big-endian
		length := uint32(len(data))
		header := make([]byte, 4)
		binary.BigEndian.PutUint32(header, length)
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

// Helper function to compare byte slices
func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
