package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitSQLite(path string) {
	var err error
	DB, err = sql.Open("sqlite", path)
	if err != nil {
		log.Fatalf("failed to open SQLite database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("failed to ping SQLite database: %v", err)
	}

	createTables := `
	CREATE TABLE IF NOT EXISTS manga (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		author TEXT,
		artist TEXT,
		genres TEXT,               -- stored as JSON array string
		chapter_count INTEGER,
		published_year INTEGER,
		status TEXT,
		cover_url TEXT,
		description TEXT
	);
	`
	if _, err := DB.Exec(createTables); err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}

	log.Println("✅ SQLite connected and table ready at", path)
}

func Close() {
	if DB != nil {
		DB.Close()
		log.Println("SQLite connection closed.")
	}
}
