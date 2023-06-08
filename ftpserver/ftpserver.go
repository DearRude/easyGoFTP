package ftpserver

import (
	"net"

	"bufio"
	"log"
	"strings"
)

// FTPConn represents a connection to the FTP server
type FTPConn struct {
	net.Conn
	Username     string
	MainDir      string
	PassiveMode  bool
	DataHost     string
	DataPort     int
	DataListener net.Listener
	DataConn     net.Conn
	TransferMode string
}

// handleFTPCommands handles the FTP commands received on the connection
func HandleFTPCommands(conn *FTPConn) {
	reader := bufio.NewReader(conn)

	for {
		command, err := reader.ReadString('\n')
		if err != nil {
			conn.Write([]byte("500 Error reading command\r\n"))
			return
		}

		command = strings.TrimSpace(command)
		log.Printf("Connection command: %s\r\n", command)

		parts := strings.Split(command, " ")
		cmd := strings.ToUpper(parts[0])
		args := parts[1:]

		switch cmd {
		case "USER":
			handleUSERCommand(conn, args)
		case "PASS":
			handlePASSCommand(conn, args)
		case "LIST":
			handleLISTCommand(conn)
		case "RETR":
			handleRETRCommand(conn, args)
		case "STOR":
			handleSTORCommand(conn, args)
		case "EPSV":
			handleEPSVCommand(conn)
		case "PASV":
			handlePASVCommand(conn)
		case "SYST":
			handleSYSTCommand(conn)
		case "TYPE":
			handleTYPECommand(conn, args)
		case "QUIT":
			conn.Write([]byte("221 Goodbye\r\n"))
			conn.Close()
			return
		default:
			conn.Write([]byte("502 Command not implemented\r\n"))
		}
	}
}
