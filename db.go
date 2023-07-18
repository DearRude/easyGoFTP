package main

import (
	"database/sql"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type DbAPI struct {
	DB *sql.DB
}

func initDb(curent_dir string) *DbAPI {
	dbPath := filepath.Join(curent_dir, "db/data.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	createTable := `CREATE TABLE IF NOT EXISTS users (
                       id INTEGER PRIMARY KEY AUTOINCREMENT,
                       username TEXT NOT NULL,
                       password TEXT NOT NULL
                   );

                   CREATE TABLE IF NOT EXISTS admins (
                       id INTEGER PRIMARY KEY AUTOINCREMENT,
                       username TEXT NOT NULL,
                       password TEXT NOT NULL
                   );`
	_, err = db.Exec(createTable)
	if err != nil {
		panic(err)
	}

	return &DbAPI{DB: db}
}
