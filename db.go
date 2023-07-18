package main

import (
	"database/sql"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func initDb(curent_dir string) *sql.DB {
	dbPath := filepath.Join(curent_dir, "data.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
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

	return db
}
