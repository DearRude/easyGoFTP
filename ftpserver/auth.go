package ftpserver

import (
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

func handleUSERCommand(conn *FTPServer, args []string) {
	if len(args) < 1 {
		_, _ = conn.Write([]byte("501 Syntax error in parameters\r\n"))
		return
	}

	username := args[0]

	_, err := conn.DB.Exec("SELECT id FROM users WHERE username = ?", username)
	if err != nil {
		_, _ = conn.Write([]byte("530 Not logged in\r\n"))
		return
	}

	// Store the username in the FTPConn for later use
	conn.Username = username

	log.Printf("Connection requested with username %s", conn.Username)
	_, _ = conn.Write([]byte("331 User name okay, need password\r\n"))
}

func handlePASSCommand(conn *FTPServer, args []string) {
	if len(args) < 1 {
		_, _ = conn.Write([]byte("501 Syntax error in parameters\r\n"))
		return
	}

	password := args[0]

	var hashedPassword string
	if err := conn.DB.QueryRow("SELECT password FROM users WHERE username = ?", conn.Username).Scan(&hashedPassword); err != nil {
		_, _ = conn.Write([]byte("530 Not logged in\r\n"))
		return
	}

	// Compare the provided password with the stored hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		_, _ = conn.Write([]byte("530 Not logged in\r\n"))
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

	conn.IsAuthed = true
	log.Printf("Connection authenticaed with username %s", conn.Username)
	_, _ = conn.Write([]byte("230 User logged in, proceed\r\n"))
}
