package db

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestDB creates a temporary database for testing
func setupTestDB(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "scope-db-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Override config directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	// Cleanup function
	cleanup := func() {
		Close()
		ResetForTesting()
		os.Setenv("HOME", originalHome)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestInitDB(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB()
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Check if database file was created
	homeDir, _ := os.UserHomeDir()
	dbPath := filepath.Join(homeDir, ".config", "scope", "scope.db")

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file was not created at %s", dbPath)
	}
}

func TestInitDBCreatesConfigDirectory(t *testing.T) {
	tmpDir, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB()
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Check if config directory exists
	configDir := filepath.Join(tmpDir, ".config", "scope")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config directory was not created at %s", configDir)
	}
}

func TestInitDBCreatesTables(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB()
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	database := GetDB()
	if database == nil {
		t.Fatal("GetDB returned nil")
	}

	// Test that tables exist by querying them
	tables := []string{"folders", "tags", "folder_tags"}
	for _, table := range tables {
		var count int
		err := database.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
		if err != nil {
			t.Errorf("Table %s does not exist or query failed: %v", table, err)
		}
	}
}

func TestInitDBIdempotent(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	// Call InitDB multiple times
	err := InitDB()
	if err != nil {
		t.Fatalf("First InitDB failed: %v", err)
	}

	// Reset singleton for this test
	ResetForTesting()

	err = InitDB()
	if err != nil {
		t.Fatalf("Second InitDB failed: %v", err)
	}
}

func TestGetDB(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB()
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	database := GetDB()
	if database == nil {
		t.Error("GetDB returned nil after InitDB")
	}
}

func TestGetDBBeforeInit(t *testing.T) {
	// Don't initialize database
	database := GetDB()
	if database != nil {
		t.Error("GetDB should return nil before InitDB is called")
	}
}

func TestClose(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB()
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	err = Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestCloseBeforeInit(t *testing.T) {
	err := Close()
	if err != nil {
		t.Errorf("Close should not error when called before InitDB: %v", err)
	}
}

func TestDatabaseIndexes(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB()
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	database := GetDB()

	// Check if indexes exist
	indexes := []string{"idx_folder_tags_tag", "idx_folder_tags_folder"}
	for _, index := range indexes {
		var count int
		err := database.QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?",
			index,
		).Scan(&count)
		if err != nil {
			t.Errorf("Failed to query for index %s: %v", index, err)
		}
		if count == 0 {
			t.Errorf("Index %s does not exist", index)
		}
	}
}

func TestDatabaseForeignKeys(t *testing.T) {
	_, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB()
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	database := GetDB()

	// Insert test data
	result, err := database.Exec("INSERT INTO folders (path, created_at) VALUES (?, ?)", "/test", 123)
	if err != nil {
		t.Fatalf("Failed to insert folder: %v", err)
	}
	folderID, _ := result.LastInsertId()

	result, err = database.Exec("INSERT INTO tags (name, created_at) VALUES (?, ?)", "test-tag", 123)
	if err != nil {
		t.Fatalf("Failed to insert tag: %v", err)
	}
	tagID, _ := result.LastInsertId()

	// Insert folder_tag
	_, err = database.Exec(
		"INSERT INTO folder_tags (folder_id, tag_id, created_at) VALUES (?, ?, ?)",
		folderID, tagID, 123,
	)
	if err != nil {
		t.Fatalf("Failed to insert folder_tag: %v", err)
	}

	// Delete the folder (should cascade to folder_tags)
	_, err = database.Exec("DELETE FROM folders WHERE id = ?", folderID)
	if err != nil {
		t.Fatalf("Failed to delete folder: %v", err)
	}

	// Check that folder_tag was also deleted
	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM folder_tags WHERE folder_id = ?", folderID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query folder_tags: %v", err)
	}
	if count != 0 {
		t.Error("Foreign key cascade delete did not work for folder_tags")
	}
}

func BenchmarkInitDB(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "scope-db-bench-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer func() {
		Close()
		ResetForTesting()
		os.Setenv("HOME", originalHome)
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ResetForTesting()
		InitDB()
	}
}
