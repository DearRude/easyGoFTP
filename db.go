package main

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func initDb(curent_dir, adminName, adminPass string) (*sql.DB, error) {
	dbPath := filepath.Join(curent_dir, "data.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if _, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			role TEXT NOT NULL
		)
	`); err != nil {
		panic(err)
	}

	err = insertDefaultUser(db, adminName, adminPass)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Insert the default user into the database if no user with role "admin" is present
func insertDefaultUser(db *sql.DB, username, password string) error {
	query := `
		INSERT INTO users (username, password, role)
		SELECT ?, ?, 'admin'
		WHERE NOT EXISTS (SELECT 1 FROM users WHERE role = 'admin')
	`

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec(query, username, string(hash), "admin")
	return err
}
