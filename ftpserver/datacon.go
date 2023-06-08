package ftpserver

import (
	"net"

	"fmt"
	"strconv"
	"strings"
)

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
	if !IsAuthenticated(conn) {
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
	if !IsAuthenticated(conn) {
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
