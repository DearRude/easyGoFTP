package ftpserver

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
)

func handleLPRTCommand(conn *FTPServer, args []string) {
	if len(args) == 0 {
		_, _ = conn.Write([]byte("501 Syntax error in parameters or arguments\r\n"))
		return
	}

	err := establishActiveDataConnection(conn, args[0])
	if err != nil {
		_, _ = conn.Write([]byte("425 Can't open data connection\r\n"))
		return
	}

	_, _ = conn.Write([]byte("200 Active data connection established\r\n"))
}

func handleEPRTCommand(conn *FTPServer, args []string) {
	if len(args) == 0 {
		_, _ = conn.Write([]byte("501 Syntax error in parameters or arguments\r\n"))
		return
	}

	_, _ = conn.Write([]byte("200 Active data connection established\r\n"))
	err := establishActiveDataConnection(conn, args[0])
	if err != nil {
		_, _ = conn.Write([]byte("425 Can't open data connection\r\n"))
		return
	}
}

func handleEPSVCommand(conn *FTPServer, args []string) {
	if len(args) > 0 {
		_, _ = conn.Write([]byte("501 Syntax error in parameters or arguments\r\n"))
		return
	}

	err := establishPassiveDataConnection(conn)
	if err != nil {
		_, _ = conn.Write([]byte("425 Can't open data connection\r\n"))
		return
	}

	port := conn.DataListener.Addr().(*net.TCPAddr).Port
	_, _ = conn.Write([]byte(fmt.Sprintf("229 Entering Extended Passive Mode (|||%d|)\r\n", port)))

	if conn.DataConn, err = conn.DataListener.Accept(); err != nil {
		_, _ = conn.Write([]byte("425 Can't open data connection\r\n"))
	}
}

func handlePASVCommand(conn *FTPServer, args []string) {
	if len(args) > 0 {
		_, _ = conn.Write([]byte("501 Syntax error in parameters or arguments\r\n"))
		return
	}

	err := establishPassiveDataConnection(conn)
	if err != nil {
		_, _ = conn.Write([]byte("425 Can't open data connection\r\n"))
		return
	}

	ip := conn.DataConn.LocalAddr().(*net.TCPAddr).IP
	port := conn.DataConn.LocalAddr().(*net.TCPAddr).Port

	ipParts := strings.Split(ip.String(), ".")
	portHigh := port / 256
	portLow := port % 256

	_, _ = conn.Write([]byte(fmt.Sprintf("227 Entering Passive Mode (%s,%s,%s,%s,%d,%d)\r\n", ipParts[0], ipParts[1], ipParts[2], ipParts[3], portHigh, portLow)))
	if conn.DataConn, err = conn.DataListener.Accept(); err != nil {
		_, _ = conn.Write([]byte("425 Can't open data connection\r\n"))
	}
}

func establishActiveDataConnection(conn *FTPServer, addr string) error {
	protocol, ip, port, err := parseAddress(addr)
	if err != nil {
		return err
	}

	host := net.JoinHostPort(ip, port)

	switch protocol {
	case "1": // Use TCP
		dataConn, err := net.Dial("tcp", host)
		if err != nil {
			return err
		}
		conn.DataConn = dataConn
	case "2": // Use TCP with SSL/TLS
		var dataConn net.Conn
		if !conn.UseTLS { // If TLS is disabled, use TCP
			if dataConn, err = net.Dial("tcp", host); err != nil {
				return err
			}
		} else {
			if dataConn, err = tls.Dial("tcp", host, conn.TLSConf); err != nil {
				return err
			}
		}
		conn.DataConn = dataConn
	default:
		return errors.New("Unsupported data connection protocol")
	}

	conn.Logger.Println("Active connection set")
	return nil
}

func establishPassiveDataConnection(conn *FTPServer) error {
	var listener net.Listener
	var err error

	if !conn.UseTLS { // If TLS is disabled, use TCP
		if listener, err = net.Listen("tcp", ":0"); err != nil {
			return err
		}
	} else {
		if listener, err = tls.Listen("tcp", ":0", conn.TLSConf); err != nil {
			return err
		}
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
