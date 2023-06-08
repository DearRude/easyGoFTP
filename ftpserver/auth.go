package ftpserver

import (
	"log"
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

	// Change maindir to user
	user_dir := filepath.Join(conn.MainDir, conn.Username)

	// Check if user subdirectory exists, if not create it
	if _, err := os.Stat(user_dir); os.IsNotExist(err) {
		err := os.Mkdir(user_dir, 0755)
		if err != nil {
			log.Printf("Failed to create subdirectory %s: %v", user_dir, err)
			return
		}
		log.Printf("Subdirectory created: %s", user_dir)
	} else {
		conn.MainDir = user_dir
		conn.CurrDir = user_dir
	}

	log.Printf("Connection authenticaed with username %s", conn.Username)
	conn.Write([]byte("230 User logged in, proceed\r\n"))
}

func IsAuthenticated(conn *FTPConn) bool {
	// Check if the username exists in the password database
	_, ok := userPasswords[conn.Username]
	return ok
}
