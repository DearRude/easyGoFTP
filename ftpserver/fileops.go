package ftpserver

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func handleLISTCommand(conn *FTPConn) {
	if conn.Username == "" {
		conn.Write([]byte("530 Not logged in\r\n"))
		return
	}

	// Open the main directory
	dir, err := os.Open(conn.MainDir)
	if err != nil {
		conn.Write([]byte("550 Failed to open directory\r\n"))
		return
	}
	defer dir.Close()

	// Read the directory contents
	entries, err := dir.Readdir(-1)
	if err != nil {
		conn.Write([]byte("550 Failed to read directory\r\n"))
		return
	}

	conn.Write([]byte("150 Opening ASCII mode data connection for file list\r\n"))
	for _, file := range entries {
		line := fmt.Sprintf("%s\r\n", file.Name())
		conn.DataConn.Write([]byte(line))
	}
	conn.DataConn.Close()
	if conn.IsPassive {
		conn.DataListener.Close()
	}
	conn.Write([]byte("226 Transfer complete\r\n"))
}

func handleRETRCommand(conn *FTPConn, args []string) {
	if conn.Username == "" {
		conn.Write([]byte("530 Not logged in\r\n"))
		return
	}

	if len(args) < 1 {
		conn.Write([]byte("501 Syntax error in parameters\r\n"))
		return
	}

	filename := args[0]

	// Construct the absolute path of the file based on the main directory
	absFilePath := filepath.Join(conn.MainDir, filename)

	// Open the file
	file, err := os.Open(absFilePath)
	if err != nil {
		conn.Write([]byte("550 File not found\r\n"))
		return
	}
	defer file.Close()
	conn.Write([]byte("150 File status okay\r\n"))

	// Set the data transfer mode based on the TYPE command
	if conn.TransferMode == "A" {
		// ASCII mode
		reader := bufio.NewReader(file)
		writer := bufio.NewWriter(conn.DataConn)

		// Read and write data line by line in ASCII mode
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					conn.Write([]byte("451 Error reading file\r\n"))
				}
				break
			}

			_, err = writer.WriteString(line)
			if err != nil {
				conn.Write([]byte("451 Error writing data\r\n"))
				break
			}

			err = writer.Flush()
			if err != nil {
				conn.Write([]byte("451 Error flushing data\r\n"))
				break
			}
		}
	} else if conn.TransferMode == "I" {
		// Binary mode
		_, err := io.Copy(conn.DataConn, file)
		if err != nil {
			conn.Write([]byte("451 Error transferring data\r\n"))
			return
		}
	} else {
		// Unsupported mode
		conn.Write([]byte("504 Command not implemented for that parameter\r\n"))
		return
	}

	conn.DataConn.Close()
	if conn.IsPassive {
		conn.DataListener.Close()
	}
	conn.Write([]byte("226 Transfer complete\r\n"))
}

func handleSTORCommand(conn *FTPConn, args []string) {
	if conn.Username == "" {
		conn.Write([]byte("530 Not logged in\r\n"))
		return
	}

	if len(args) < 1 {
		conn.Write([]byte("501 Syntax error in parameters\r\n"))
		return
	}

	filename := args[0]

	// Construct the absolute path of the file based on the current directory
	absFilePath := filepath.Join(conn.CurrDir, filename)

	// Create or open the file for writing
	file, err := os.Create(absFilePath)
	if err != nil {
		conn.Write([]byte("550 Failed to create or open file\r\n"))
		return
	}
	defer file.Close()
	conn.Write([]byte("150 File status okay\r\n"))

	// Set the data transfer mode based on the TYPE command
	if conn.TransferMode == "A" {
		// ASCII mode
		reader := bufio.NewReader(conn.DataConn)
		writer := bufio.NewWriter(file)

		// Read and write data line by line in ASCII mode
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					conn.Write([]byte("451 Error receiving data\r\n"))
				}
				break
			}

			_, err = writer.WriteString(line)
			if err != nil {
				conn.Write([]byte("451 Error writing file\r\n"))
				break
			}

			err = writer.Flush()
			if err != nil {
				conn.Write([]byte("451 Error flushing data\r\n"))
				break
			}
		}
	} else if conn.TransferMode == "I" {
		// Binary mode
		_, err := io.Copy(file, conn.DataConn)
		if err != nil {
			conn.Write([]byte("451 Error receiving data\r\n"))
			return
		}
	} else {
		// Unsupported mode
		conn.Write([]byte("504 Command not implemented for that parameter\r\n"))
		return
	}

	conn.DataConn.Close()
	if conn.IsPassive {
		conn.DataListener.Close()
	}
	conn.Write([]byte("226 Transfer complete\r\n"))
}
