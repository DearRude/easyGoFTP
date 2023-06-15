package ftpserver

import (
	"os"
	"path/filepath"
)

// handleCWDCommand handles the CWD command
func handleCWDCommand(conn *FTPServer, args []string) {
	if !IsAuthenticated(conn) {
		conn.Write([]byte("530 Not logged in\r\n"))
		return
	}

	if len(args) < 1 {
		conn.Write([]byte("501 Syntax error in parameters or arguments\r\n"))
		return
	}

	// Get the target directory
	targetDir := args[0]

	// Update the current working directory for the FTP connection
	if updateWorkingDir(conn, targetDir) {
		conn.Write([]byte("250 Directory successfully changed\r\n"))
	} else {
		conn.Write([]byte("550 Requested action not taken. Directory not found\r\n"))
	}
}

// updateWorkingDir updates the current working directory for the FTP connection
func updateWorkingDir(conn *FTPServer, targetDir string) bool {
	fullPath := filepath.Join(conn.CurrDir, targetDir)
	if filepath.IsAbs(targetDir) {
		fullPath = filepath.Join(conn.MainDir, targetDir)
	}

	// Check of exists
	if _, err := os.Stat(fullPath); err != nil {
		return false
	}

	// Check if it's not parent of maindir
	if !filepath.HasPrefix(fullPath, conn.MainDir) {
		return false
	}
	conn.CurrDir = fullPath
	return true
}

// handleMKDCommand handles the MKD command
func handleMKDCommand(conn *FTPServer, args []string) {
	if !IsAuthenticated(conn) {
		conn.Write([]byte("530 Not logged in\r\n"))
		return
	}

	if len(args) < 1 {
		conn.Write([]byte("501 Syntax error in parameters or arguments\r\n"))
		return
	}

	// Get the directory name to be created
	dirName := args[0]

	// Construct the full path for the new directory
	fullPath := filepath.Join(conn.CurrDir, dirName)

	// Check if it's not parent of maindir
	if !filepath.HasPrefix(fullPath, conn.MainDir) {
		conn.Write([]byte("550 Requested action not taken. Directory localtion is not permissable\r\n"))
		return
	}

	// Create the directory
	err := os.Mkdir(fullPath, 0777)
	if err != nil {
		conn.Write([]byte("550 Requested action not taken. Failed to create directory\r\n"))
		return
	}

	conn.Write([]byte("250 Directory created successfully\r\n"))
}
