package utils

import (
	"net"
	"os"
)

func OpenVirtioPort(role string) (*os.File, error) {
	if role == "host" {
		// Host connects to the Unix socket
		conn, err := net.Dial("unix", LinuxSocketPath)
		if err != nil {
			return nil, err
		}
		// Convert net.Conn to *os.File
		file, err := conn.(*net.UnixConn).File()
		if err != nil {
			conn.Close()
			return nil, err
		}
		return file, nil
	} else {
		// Guest opens the virtio-serial port
		return os.OpenFile(LinuxVirtioPortPath, os.O_RDWR, 0600)
	}
}
