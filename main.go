package main

import (
	"net"

	"fmt"
	"log"
	"os"
	"path/filepath"

	fs "github.com/dearrude/easygoftp/ftpserver"
)

func main() {
	port := 21211

	cur_dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get current directory:", err)
		return
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	log.Printf("FTP server started on port %d\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		log.Printf("Connection accepted from %s", conn.RemoteAddr())
		conn.Write([]byte("220 Service ready for new user\r\n"))

		ftpConn := fs.FTPConn{
			Conn:    conn,
			MainDir: filepath.Join(cur_dir, "files"),
		}

		go fs.HandleFTPCommands(&ftpConn)
	}
}
