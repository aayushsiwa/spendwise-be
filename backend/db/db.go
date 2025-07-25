package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(path string) {
	var err error
	DB, err = sql.Open("sqlite", path)
	if err != nil {
		log.Fatal("DB open error:", err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatal("DB connection error (ping failed):", err)
	}

	_, err = DB.Exec(`PRAGMA foreign_keys = ON;`)
	if err != nil {
		log.Fatal("PRAGMA error:", err)
	}

	log.Println("Database connected:", path)
}
