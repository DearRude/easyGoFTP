package ftpserver

import (
	"crypto/tls"
	"net"

	"bufio"
	"database/sql"
	"log"
	"strings"
)

// FTPServer represents a connection to the FTP server
type FTPServer struct {
	net.Conn

	Logger    *log.Logger
	ErrLogger *log.Logger

	UseTLS  bool
	TLSConf *tls.Config

	DataConn     net.Conn
	DataListener net.Listener
	IsPassive    bool
	Username     string
	MainDir      string
	CurrDir      string
	TransferMode string

	DB       *sql.DB
	IsAuthed bool
}

// handleFTPCommands handles the FTP commands received on the connection
func HandleFTPCommands(conn *FTPServer) {
	reader := bufio.NewReader(conn)

	for {
		command, err := reader.ReadString('\n')
		if err != nil {
			_, _ = conn.Write([]byte("500 Error reading command\r\n"))
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
		case "CWD":
			handleCWDCommand(conn, args)
		case "MKD":
			handleMKDCommand(conn, args)
		case "LIST":
			handleLISTCommand(conn)
		case "PWD":
			handlePWDCommand(conn)
		case "RETR":
			handleRETRCommand(conn, args)
		case "STOR":
			handleSTORCommand(conn, args)
		case "EPSV":
			handleEPSVCommand(conn, args)
		case "PASV":
			handlePASVCommand(conn, args)
		case "LPRT":
			handleLPRTCommand(conn, args)
		case "EPRT":
			handleEPRTCommand(conn, args)
		case "SYST":
			handleSYSTCommand(conn)
		case "STRU":
			handleSTRUCommand(conn, args)
		case "MODE":
			handleMODECommand(conn, args)
		case "TYPE":
			handleTYPECommand(conn, args)
		case "QUIT":
			_, _ = conn.Write([]byte("221 Goodbye\r\n"))
			conn.Close()
			return
		default:
			_, _ = conn.Write([]byte("502 Command not implemented\r\n"))
		}
	}
}
