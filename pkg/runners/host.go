package runners

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/thelicato/dqcs/pkg/logger"
	"github.com/thelicato/dqcs/pkg/utils"
	"golang.design/x/clipboard"
)

func RunHost() {
	logger.Info("Running Host component....")
	conn, err := utils.OpenVirtioPort("host")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening virtio port: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	handleConnection(conn)
}

func handleConnection(conn io.ReadWriteCloser) {
	var wg sync.WaitGroup
	clipboardCh := make(chan []byte)

	// Initialize the clipboard package
	err := clipboard.Init()
	if err != nil {
		fmt.Println("Error initializing clipboard:", err)
		os.Exit(1)
	}

	wg.Add(3)
	go monitorClipboard(clipboardCh, &wg)
	go readFromConnection(conn, &wg)
	go writeToConnection(conn, clipboardCh, &wg)

	wg.Wait()
}

func monitorClipboard(ch chan<- []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	var lastText []byte
	for {
		text := clipboard.Read(clipboard.FmtText)
		if text == nil {
			time.Sleep(time.Second)
			continue
		}
		if !bytes.Equal(text, lastText) {
			// Make a copy of text to avoid data race
			lastText = make([]byte, len(text))
			copy(lastText, text)
			ch <- text
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func readFromConnection(conn io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		data := scanner.Bytes()
		currentClipboard := clipboard.Read(clipboard.FmtText)
		if !bytes.Equal(data, currentClipboard) {
			clipboard.Write(clipboard.FmtText, data)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from connection:", err)
	}
}

func writeToConnection(conn io.Writer, ch <-chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	writer := bufio.NewWriter(conn)
	for text := range ch {
		_, err := writer.Write(text)
		if err != nil {
			fmt.Println("Error writing to connection:", err)
			return
		}
		_, err = writer.Write([]byte{'\n'})
		if err != nil {
			fmt.Println("Error writing newline to connection:", err)
			return
		}
		writer.Flush()
	}
}
