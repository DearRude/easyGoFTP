package ftpserver

import (
	//	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
)

func handleLPRTCommand(conn *FTPConn, args []string) {
	if len(args) == 0 {
		conn.Write([]byte("501 Syntax error in parameters or arguments\r\n"))
		return
	}

	err := establishActiveDataConnection(conn, args[0])
	if err != nil {
		conn.Write([]byte("425 Can't open data connection\r\n"))
		return
	}

	conn.Write([]byte("200 Active data connection established\r\n"))
}

func handleEPRTCommand(conn *FTPConn, args []string) {
	if len(args) == 0 {
		conn.Write([]byte("501 Syntax error in parameters or arguments\r\n"))
		return
	}

	conn.Write([]byte("200 Active data connection established\r\n"))
	err := establishActiveDataConnection(conn, args[0])
	if err != nil {
		conn.Write([]byte("425 Can't open data connection\r\n"))
		return
	}
}

func handleEPSVCommand(conn *FTPConn, args []string) {
	if len(args) > 0 {
		conn.Write([]byte("501 Syntax error in parameters or arguments\r\n"))
		return
	}

	err := establishPassiveDataConnection(conn)
	if err != nil {
		conn.Write([]byte("425 Can't open data connection\r\n"))
		return
	}

	port := conn.DataListener.Addr().(*net.TCPAddr).Port
	conn.Write([]byte(fmt.Sprintf("229 Entering Extended Passive Mode (|||%d|)\r\n", port)))

	if conn.DataConn, err = conn.DataListener.Accept(); err != nil {
		conn.Write([]byte("425 Can't open data connection\r\n"))
	}
}

func handlePASVCommand(conn *FTPConn, args []string) {
	if len(args) > 0 {
		conn.Write([]byte("501 Syntax error in parameters or arguments\r\n"))
		return
	}

	err := establishPassiveDataConnection(conn)
	if err != nil {
		conn.Write([]byte("425 Can't open data connection\r\n"))
		return
	}

	ip := conn.DataConn.LocalAddr().(*net.TCPAddr).IP
	port := conn.DataConn.LocalAddr().(*net.TCPAddr).Port

	ipParts := strings.Split(ip.String(), ".")
	portHigh := port / 256
	portLow := port % 256

	conn.Write([]byte(fmt.Sprintf("227 Entering Passive Mode (%s,%s,%s,%s,%d,%d)\r\n", ipParts[0], ipParts[1], ipParts[2], ipParts[3], portHigh, portLow)))
	if conn.DataConn, err = conn.DataListener.Accept(); err != nil {
		conn.Write([]byte("425 Can't open data connection\r\n"))
	}
}

func establishActiveDataConnection(conn *FTPConn, addr string) error {
	protocol, ip, port, err := parseAddress(addr)
	if err != nil {
		return err
	}

	host := net.JoinHostPort(ip, port)

	switch protocol {
	case "1", "2": // Use TCP
		dataConn, err := net.Dial("tcp", host)
		if err != nil {
			return err
		}
		conn.DataConn = dataConn
	// case "2": // Use TCP with SSL/TLS
	// 	tlsConfig := &tls.Config{
	// 		InsecureSkipVerify: true, // Disable certificate verification (for testing purposes)
	// 	}

	// 	dataConn, err := tls.Dial("tcp", host, tlsConfig)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	conn.DataConn = dataConn
	default:
		return errors.New("Unsupported data connection protocol")
	}
	log.Println("Active connection set")

	return nil
}

func establishPassiveDataConnection(conn *FTPConn) error {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}

	conn.DataListener = listener
	conn.IsPassive = true
	return nil
}

func parseAddress(addr string) (string, string, string, error) {
	parts := strings.Split(addr, "|")
	if len(parts) < 4 {
		return "", "", "", fmt.Errorf("invalid address format")
	}

	// proto, ip, port, error
	return parts[1], parts[2], parts[3], nil
}
