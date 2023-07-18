package ftpserver

import (
	"runtime"
	"strings"
)

func handleTYPECommand(conn *FTPServer, args []string) {
	if !conn.IsAuthed {
		_, _ = conn.Write([]byte("530 Not logged in\r\n"))
		return
	}

	if len(args) < 1 {
		_, _ = conn.Write([]byte("501 Syntax error in parameters or arguments\r\n"))
		return
	}

	mode := strings.ToUpper(args[0])
	switch mode {
	case "A", "A N":
		// ASCII mode
		conn.TransferMode = "A"
		_, _ = conn.Write([]byte("200 Switching to ASCII mode\r\n"))
	case "I", "L 8":
		// Binary mode
		conn.TransferMode = "I"
		_, _ = conn.Write([]byte("200 Switching to binary mode\r\n"))
	default:
		_, _ = conn.Write([]byte("504 Command not implemented for that parameter\r\n"))
	}
}

func handleSYSTCommand(conn *FTPServer) {
	if !conn.IsAuthed {
		_, _ = conn.Write([]byte("530 Not logged in\r\n"))
		return
	}

	// Determine the system type based on the operating system
	var systemType string
	switch runtime.GOOS {
	case "windows":
		systemType = "Windows"
	case "darwin":
		systemType = "UNIX"
	case "linux":
		systemType = "UNIX"
	default:
		systemType = "Unknown"
	}

	// Send the system information as the response
	response := "215 " + systemType + " Type: L8\r\n"
	_, _ = conn.Write([]byte(response))
}
