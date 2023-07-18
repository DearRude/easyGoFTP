package main

import (
	"net"

	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"

	fs "github.com/dearrude/easygoftp/ftpserver"
	ws "github.com/dearrude/easygoftp/webserver"
)

func main() {
	c := GenConfig()

	// Get current directory
	cur_dir, err := os.Getwd()
	if err != nil {
		c.StderrLogger.Println("Failed to get current directory:", err)
		return
	}

	db := initDb(cur_dir)
	defer db.Close()

	// If secure, handle TLS
	tlsConfig := &tls.Config{}
	if c.UseTLS {
		tlsConfig = fs.GetTLSConfig(c.Domain)
	}

	// Init web server
	go ws.Setup(
		c.WebPort,
		&c.StdoutLogger,
		&c.StderrLogger,
		filepath.Join(cur_dir, "files"),
		db,
	)

	// Init FTP server
	go func() {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", c.FTPPort))
		if err != nil {
			c.StderrLogger.Println("Failed to start the FTP server on designated port:", err)
			return
		}
		defer listener.Close()
		c.StdoutLogger.Printf("FTP server started on port %d\n", c.FTPPort)

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
	}()

	select {} // keep the main goroutine running
}
