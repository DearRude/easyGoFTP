package main

import (
	"net"

	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"

	fs "github.com/dearrude/easygoftp/ftpserver"
)

func main() {
	c := GenConfig()

	cur_dir, err := os.Getwd()
	if err != nil {
		c.StderrLogger.Println("Failed to get current directory:", err)
		return
	}

	// If secure, handle TLS
	tlsConfig := &tls.Config{}
	if c.UseTLS {
		tlsConfig = fs.GetTLSConfig(c.Domain)
	}

	// Init FTP server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Port))
	if err != nil {
		c.StderrLogger.Println("Failed to start the FTP server on designated port:", err)
		return
	}
	defer listener.Close()
	c.StdoutLogger.Printf("FTP server started on port %d\n", c.Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			c.StderrLogger.Println("Connection could not be accepted", err)
			continue
		}
		c.StdoutLogger.Printf("Connection accepted from %s", conn.RemoteAddr())
		_, _ = conn.Write([]byte("220 Service ready for new user\r\n"))

		ftpConn := fs.FTPServer{
			Logger:    &c.StdoutLogger,
			ErrLogger: &c.StderrLogger,
			UseTLS:    c.UseTLS,
			TLSConf:   tlsConfig,
			Conn:      conn,
			MainDir:   filepath.Join(cur_dir, "files"),
		}

		// Handle the requests concurrently
		go fs.HandleFTPCommands(&ftpConn)
	}
}
