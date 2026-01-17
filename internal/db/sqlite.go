package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db   *sql.DB
	once sync.Once
)

// InitDB initializes the database connection and creates tables if needed
func InitDB() error {
	var err error
	once.Do(func() {
		// Get user's home directory
		homeDir, e := os.UserHomeDir()
		if e != nil {
			err = fmt.Errorf("failed to get home directory: %w", e)
			return
		}

		// Create config directory
		configDir := filepath.Join(homeDir, ".config", "scope")
		if e := os.MkdirAll(configDir, 0755); e != nil {
			err = fmt.Errorf("failed to create config directory: %w", e)
			return
		}

		// Open database
		dbPath := filepath.Join(configDir, "scope.db")
		db, e = sql.Open("sqlite3", dbPath)
		if e != nil {
			err = fmt.Errorf("failed to open database: %w", e)
			return
		}

		// Create tables
		err = createTables()
	})
	return err
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	return db
}

// Close closes the database connection
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// createTables creates the necessary database tables
func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS folders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT UNIQUE NOT NULL,
		created_at INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		created_at INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS folder_tags (
		folder_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		PRIMARY KEY (folder_id, tag_id),
		FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_folder_tags_tag ON folder_tags(tag_id);
	CREATE INDEX IF NOT EXISTS idx_folder_tags_folder ON folder_tags(folder_id);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}
