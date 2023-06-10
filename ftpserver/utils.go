package ftpserver

import (
	"runtime"
	"strings"
)

func handleTYPECommand(conn *FTPConn, args []string) {
	if !IsAuthenticated(conn) {
		conn.Write([]byte("530 Not logged in\r\n"))
		return
	}

	if len(args) < 1 {
		conn.Write([]byte("501 Syntax error in parameters or arguments\r\n"))
		return
	}

	mode := strings.ToUpper(args[0])
	switch mode {
	case "A", "A N":
		// ASCII mode
		conn.TransferMode = "A"
		conn.Write([]byte("200 Switching to ASCII mode\r\n"))
	case "I", "L 8":
		// Binary mode
		conn.TransferMode = "I"
		conn.Write([]byte("200 Switching to binary mode\r\n"))
	default:
		conn.Write([]byte("504 Command not implemented for that parameter\r\n"))
	}
}

func handleSYSTCommand(conn *FTPConn) {
	if !IsAuthenticated(conn) {
		conn.Write([]byte("530 Not logged in\r\n"))
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
	conn.Write([]byte(response))
}
