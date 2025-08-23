package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	dbPath := "limiter.db"

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Create users table with created_at
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
			quota FLOAT NOT NULL,
            created_at DATETIME NOT NULL
        )
    `)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}

	now := time.Now().Format(time.RFC3339)

	// Insert two users with creation date
	_, err = db.Exec(`
        INSERT OR REPLACE INTO users (id, name, quota, created_at) VALUES
        (1, 'Ionel', 0.5, ?),
        (2, 'Ionela', 1.0, ?)
    `, now, now)
	if err != nil {
		log.Fatal("Failed to insert users:", err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS request_count (
		user_id INTEGER NOT NULL,
		path TEXT NOT NULL,
		window_start INTEGER NOT NULL,
		count INTEGER NOT NULL,
		PRIMARY KEY (user_id, path, window_start)
	);`)

	fmt.Println("Migration completed successfully.")
}
