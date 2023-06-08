package main

import (
	"fmt"
	"net"

	"bufio"
	"log"
	"strconv"
	"strings"

	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

// Store hashed passwords for user accounts
var userPasswords = map[string]string{
	"user_1": "$2a$10$yolHY3AWMRzWrIXIhtAv6uMB/gvyqpD4I.SAmMvGZIMa3.hievNT6",
	"user_2": "$2a$10$NLbsJ2KpLyVsvSZ0yVfiY.UgwL9NNvi0VVtDOoz5s3352A3rU2LCS",
	"user_3": "$2a$10$4St/8SvmVkzrWkkau67pl.lRJ5E9iMxmOnY5LAIJabpVGANgXrhsa",
}

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

func handleUSERCommand(conn *FTPConn, args []string) {
	if len(args) < 1 {
		conn.Write([]byte("501 Syntax error in parameters\r\n"))
		return
	}

	username := args[0]

	// Check if the username exists in the password database
	_, ok := userPasswords[username]
	if !ok {
		conn.Write([]byte("530 Not logged in\r\n"))
		return
	}

	// Store the username in the FTPConn for later use
	conn.Username = username

	log.Printf("Connection requested with username %s", conn.Username)
	conn.Write([]byte("331 User name okay, need password\r\n"))
}

func handlePASSCommand(conn *FTPConn, args []string) {
	if len(args) < 1 {
		conn.Write([]byte("501 Syntax error in parameters\r\n"))
		return
	}

	password := args[0]

	// Retrieve the previously stored hashed password for the user
	hashedPassword, ok := userPasswords[conn.Username]
	if !ok {
		conn.Write([]byte("530 Not logged in\r\n"))
		return
	}

	// Compare the provided password with the stored hashed password
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		conn.Write([]byte("530 Not logged in\r\n"))
		return
	}

	log.Printf("Connection authenticaed with username %s", conn.Username)
	conn.Write([]byte("230 User logged in, proceed\r\n"))
}

func handleLISTCommand(conn *FTPConn) {
	if conn.Username == "" {
		conn.Write([]byte("530 Not logged in\r\n"))
		return
	}

	// Establish a data connection with the client
	dataConn, err := conn.connectDataConn()
	if err != nil {
		conn.Write([]byte("425 Can't open data connection\r\n"))
		log.Printf("Err: %s", err)
		return
	}
	defer dataConn.Close()

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
		dataConn.Write([]byte(line))
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

	// Establish a data connection with the client
	dataConn, err := conn.connectDataConn()
	if err != nil {
		conn.Write([]byte("425 Can't open data connection\r\n"))
		return
	}
	defer dataConn.Close()

	// Construct the absolute path of the file based on the main directory
	absFilePath := filepath.Join(conn.MainDir, filename)

	// Open the file
	file, err := os.Open(absFilePath)
	if err != nil {
		conn.Write([]byte("550 File not found\r\n"))
		return
	}
	defer file.Close()

	// Set the data transfer mode based on the TYPE command
	if conn.TransferMode == "A" {
		// ASCII mode
		reader := bufio.NewReader(file)
		writer := bufio.NewWriter(dataConn)

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
		_, err := io.Copy(dataConn, file)
		if err != nil {
			conn.Write([]byte("451 Error transferring data\r\n"))
			return
		}
	} else {
		// Unsupported mode
		conn.Write([]byte("504 Command not implemented for that parameter\r\n"))
		return
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

	// Establish a data connection with the client
	dataConn, err := conn.connectDataConn()
	if err != nil {
		conn.Write([]byte("425 Can't open data connection\r\n"))
		return
	}
	defer dataConn.Close()

	filename := args[0]

	// Construct the absolute path of the file based on the main directory
	absFilePath := filepath.Join(conn.MainDir, filename)

	// Create or open the file for writing
	file, err := os.Create(absFilePath)
	if err != nil {
		conn.Write([]byte("550 Failed to create or open file\r\n"))
		return
	}
	defer file.Close()

	// Set the data transfer mode based on the TYPE command
	if conn.TransferMode == "A" {
		// ASCII mode
		reader := bufio.NewReader(dataConn)
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
		_, err := io.Copy(file, dataConn)
		if err != nil {
			conn.Write([]byte("451 Error receiving data\r\n"))
			return
		}
	} else {
		// Unsupported mode
		conn.Write([]byte("504 Command not implemented for that parameter\r\n"))
		return
	}

	conn.Write([]byte("226 Transfer complete\r\n"))
}

// Establishes a data connection with the client
func (conn *FTPConn) connectDataConn() (net.Conn, error) {
	var dataConn net.Conn
	var err error

	// Check the preferred mode for data connection (PASV or EPSV)
	if conn.PassiveMode {
		// PASV mode: Accept the client's connection
		dataConn, err = conn.acceptDataConn()
	} else {
		// EPSV mode: Dial a connection to the client
		dataConn, err = conn.dialDataConn()
	}

	return dataConn, err
}

// Accepts the client's data connection (PASV mode)
func (conn *FTPConn) acceptDataConn() (net.Conn, error) {
	dataConn, err := conn.DataListener.Accept()
	if err != nil {
		return nil, err
	}

	return dataConn, nil
}

// Dials a connection to the client's data connection (EPSV mode)
func (conn *FTPConn) dialDataConn() (net.Conn, error) {
	epsvResponse := fmt.Sprintf("229 Entering Extended Passive Mode (|||%d|)\r\n", conn.DataPort)
	conn.Write([]byte(epsvResponse))

	dataConn, err := net.Dial("tcp", conn.DataHost+":"+strconv.Itoa(conn.DataPort))
	if err != nil {
		return nil, err
	}

	return dataConn, nil
}

func handleEPSVCommand(conn *FTPConn) {
	if !isAuthenticated(conn) {
		conn.Write([]byte("530 Not logged in\r\n"))
		return
	}

	// Prepare to listen for incoming data connections
	dataListener, err := net.Listen("tcp", ":0")
	if err != nil {
		conn.Write([]byte("500 Failed to set up data connection\r\n"))
		return
	}

	// Get the port number on which the data connection is listening
	_, dataPortStr, err := net.SplitHostPort(dataListener.Addr().String())
	if err != nil {
		conn.Write([]byte("500 Failed to get data connection port\r\n"))
		return
	}

	// Convert the port number to an integer
	dataPort, err := strconv.Atoi(dataPortStr)
	if err != nil {
		conn.Write([]byte("500 Failed to parse data connection port\r\n"))
		return
	}

	// Build the response string with the port number
	response := fmt.Sprintf("229 Entering Extended Passive Mode (|||%d|)\r\n", dataPort)

	// Send the response to the client
	conn.Write([]byte(response))
}

func handlePASVCommand(conn *FTPConn) {
	if !isAuthenticated(conn) {
		conn.Write([]byte("530 Not logged in\r\n"))
		return
	}

	// Prepare to listen for incoming data connections
	dataListener, err := net.Listen("tcp", ":0")
	if err != nil {
		conn.Write([]byte("500 Failed to set up data connection\r\n"))
		return
	}

	// Get the host and port information for the data connection
	host, portStr, err := net.SplitHostPort(dataListener.Addr().String())
	if err != nil {
		conn.Write([]byte("500 Failed to get data connection details\r\n"))
		return
	}

	// Parse the port number from the string
	port, err := strconv.Atoi(portStr)
	if err != nil {
		conn.Write([]byte("500 Failed to parse data connection port\r\n"))
		return
	}

	// Build the response string with the host and port information
	response := fmt.Sprintf("227 Entering Passive Mode (%s,%d,%d)\r\n",
		strings.ReplaceAll(host, ".", ","), port/256, port%256)

	// Send the response to the client
	conn.Write([]byte(response))
}

func handleTYPECommand(conn *FTPConn, args []string) {
	if !isAuthenticated(conn) {
		conn.Write([]byte("530 Not logged in\r\n"))
		return
	}

	if len(args) < 1 {
		conn.Write([]byte("501 Syntax error in parameters or arguments\r\n"))
		return
	}

	mode := strings.ToUpper(args[0])
	switch mode {
	case "A":
		// ASCII mode
		conn.Write([]byte("200 Switching to ASCII mode\r\n"))
	case "I":
		// Binary mode
		conn.Write([]byte("200 Switching to binary mode\r\n"))
	default:
		conn.Write([]byte("504 Command not implemented for that parameter\r\n"))
	}
}

func handleSYSTCommand(conn *FTPConn) {
	if !isAuthenticated(conn) {
		conn.Write([]byte("530 Not logged in\r\n"))
		return
	}

	// Send the system information as the response
	response := "215 UNIX Type: L8\r\n"
	conn.Write([]byte(response))
}

func isAuthenticated(conn *FTPConn) bool {
	// Check if the username exists in the password database
	_, ok := userPasswords[conn.Username]
	return ok
}

func handleFTPCommands(conn *FTPConn) {
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

func main() {
	port := 21211
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

		ftpConn := FTPConn{Conn: conn, MainDir: "/home/ebrahim/temp-ftp-dir"}

		go handleFTPCommands(&ftpConn)
	}
}
