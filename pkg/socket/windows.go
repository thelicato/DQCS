//go:build windows
// +build windows

package socket

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/thelicato/dqcs/pkg/utils"
	"golang.design/x/clipboard"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

type dqcsService struct{}

func (m *dqcsService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	s <- svc.Status{State: svc.StartPending}
	go guest()
	s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	for {
		c := <-r
		switch c.Cmd {
		case svc.Interrogate:
			s <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			s <- svc.Status{State: svc.StopPending}
			return false, 0
		}
	}
}

func installService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.CreateService(utils.WinServiceName, os.Args[0], mgr.Config{DisplayName: "DQCS"})
	if err != nil {
		return err
	}
	defer s.Close()

	err = eventlog.InstallAsEventCreate(utils.WinServiceName, eventlog.Info|eventlog.Error|eventlog.Warning)
	if err != nil {
		s.Delete()
		return fmt.Errorf("setup event log failed: %s", err)
	}
	return nil
}

func uninstallService() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(utils.WinServiceName)
	if err != nil {
		return err
	}
	defer s.Close()

	err = s.Delete()
	if err != nil {
		return err
	}

	err = eventlog.Remove(utils.WinServiceName)
	if err != nil {
		return fmt.Errorf("remove event log failed: %s", err)
	}
	return nil
}

func runAsHost(_ string) {
	fmt.Println("Windows is not supported as Host")
}

func openVirtioSerialPort(path string) (*os.File, error) {
	handle, err := windows.CreateFile(
		windows.StringToUTF16Ptr(path),
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		0,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_OVERLAPPED,
		0,
	)
	if err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(handle), path), nil
}

func runAsGuest() {
	err := svc.Run(utils.WinServiceName, &dqcsService{})
	if err != nil {
		fmt.Println("Running as interactive mode")
		guest()
	}
}

func guest() {
	file, err := openVirtioSerialPort(utils.WinVirtioPortName)
	if err != nil {
		fmt.Printf("Error opening virtio-serial port: %v\n", err)
		return
	}
	defer file.Close()

	handleConnection(file)
}

func handleConnection(conn *os.File) {
	defer conn.Close()

	err := clipboard.Init()
	if err != nil {
		fmt.Println("Error initializing clipboard:", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	quit := make(chan struct{})

	wg.Add(2)
	go readLoop(conn, quit, &wg)
	go writeLoop(conn, quit, &wg)

	wg.Wait()
}

func readLoop(conn *os.File, quit chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-quit:
			return
		default:
			message, err := readMessage(conn)
			if err != nil {
				if err == io.EOF {
					fmt.Println("Connection closed by remote.")
				} else {
					fmt.Printf("Error reading message: %v\n", err)
				}
				close(quit)
				return
			}
			currentClipboard := clipboard.Read(clipboard.FmtText)
			if !bytes.Equal(message, currentClipboard) {
				done := clipboard.Write(clipboard.FmtText, message)
				<-done // Wait for the write operation to complete
			}
		}
	}
}

func writeLoop(conn *os.File, quit chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clipboardCh := clipboard.Watch(ctx, clipboard.FmtText)
	var lastClipboard []byte

	for {
		select {
		case <-quit:
			return
		case data := <-clipboardCh:
			if bytes.Equal(data, lastClipboard) {
				continue // Skip if the data hasn't changed
			}

			err := writeMessage(conn, data)
			if err != nil {
				close(quit)
				return
			}
		}
	}
}

func asyncWrite(conn *os.File, data []byte) error {
	handle := windows.Handle(conn.Fd())
	overlapped := new(windows.Overlapped)

	// Create an event for the overlapped operation
	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(event)
	overlapped.HEvent = event

	var bytesWritten uint32
	err = windows.WriteFile(handle, data, &bytesWritten, overlapped)
	if err != nil && err != windows.ERROR_IO_PENDING {
		return err
	}

	// Wait for the write operation to complete
	_, err = windows.WaitForSingleObject(overlapped.HEvent, windows.INFINITE)
	if err != nil {
		return err
	}

	// Get the result of the write operation
	err = windows.GetOverlappedResult(handle, overlapped, &bytesWritten, false)
	if err != nil {
		return err
	}

	if bytesWritten != uint32(len(data)) {
		return fmt.Errorf("asyncWrite: incomplete write (%d/%d bytes)", bytesWritten, len(data))
	}

	return nil
}

func asyncRead(conn *os.File, buffer []byte) (int, error) {
	handle := windows.Handle(conn.Fd())
	overlapped := new(windows.Overlapped)

	// Create an event for the overlapped operation
	event, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(event)
	overlapped.HEvent = event

	var bytesRead uint32
	err = windows.ReadFile(handle, buffer, &bytesRead, overlapped)
	if err != nil && err != windows.ERROR_IO_PENDING {
		return 0, err
	}

	// Wait for the read operation to complete
	_, err = windows.WaitForSingleObject(overlapped.HEvent, windows.INFINITE)
	if err != nil {
		return 0, err
	}

	// Get the result of the read operation
	err = windows.GetOverlappedResult(handle, overlapped, &bytesRead, false)
	if err != nil {
		return 0, err
	}

	return int(bytesRead), nil
}

func readMessage(conn *os.File) ([]byte, error) {
	// Read the length prefix asynchronously
	lengthBuf := make([]byte, 4)
	_, err := asyncRead(conn, lengthBuf)
	if err != nil {
		return nil, err
	}

	length := binary.LittleEndian.Uint32(lengthBuf)
	if length == 0 {
		return nil, fmt.Errorf("Invalid message length: 0")
	}

	// Read the message data asynchronously
	data := make([]byte, length)
	totalRead := 0
	for totalRead < int(length) {
		n, err := asyncRead(conn, data[totalRead:])
		if err != nil {
			return nil, err
		}
		totalRead += n
	}

	return data, nil
}

func writeMessage(conn *os.File, data []byte) error {
	length := uint32(len(data))
	if length == 0 {
		return fmt.Errorf("Cannot send empty message")
	}

	// Prepare the length prefix
	lengthBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthBuf, length)

	// Write the length prefix asynchronously
	err := asyncWrite(conn, lengthBuf)
	if err != nil {
		return fmt.Errorf("Error writing length prefix: %v", err)
	}

	// Write the data asynchronously
	err = asyncWrite(conn, data)
	if err != nil {
		return fmt.Errorf("Error writing data: %v", err)
	}

	return nil
}
