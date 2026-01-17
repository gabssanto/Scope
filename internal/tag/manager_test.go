package tag

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/gabssanto/Scope/internal/db"
)

// setupTestEnv creates a test environment with temporary database
func setupTestEnv(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp directories
	tmpDir, err := os.MkdirTemp("", "scope-tag-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	testFolder := filepath.Join(tmpDir, "test-folder")
	if err := os.MkdirAll(testFolder, 0755); err != nil {
		t.Fatalf("Failed to create test folder: %v", err)
	}

	// Override HOME for database
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	// Initialize database
	if err := db.InitDB(); err != nil {
		t.Fatalf("Failed to init database: %v", err)
	}

	cleanup := func() {
		db.Close()
		db.ResetForTesting()
		os.Setenv("HOME", originalHome)
		os.RemoveAll(tmpDir)
	}

	return testFolder, cleanup
}

func TestAddTag(t *testing.T) {
	testFolder, cleanup := setupTestEnv(t)
	defer cleanup()

	err := AddTag(testFolder, "work")
	if err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	// Verify tag was added
	tags, err := GetTagsForFolder(testFolder)
	if err != nil {
		t.Fatalf("GetTagsForFolder failed: %v", err)
	}

	if len(tags) != 1 || tags[0] != "work" {
		t.Errorf("Expected tags [work], got %v", tags)
	}
}

func TestAddTagNonExistentFolder(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	err := AddTag("/nonexistent/folder", "test")
	if err == nil {
		t.Error("AddTag should fail for non-existent folder")
	}
}

func TestAddMultipleTags(t *testing.T) {
	testFolder, cleanup := setupTestEnv(t)
	defer cleanup()

	// Add multiple tags to same folder
	err := AddTag(testFolder, "work")
	if err != nil {
		t.Fatalf("AddTag 'work' failed: %v", err)
	}

	err = AddTag(testFolder, "urgent")
	if err != nil {
		t.Fatalf("AddTag 'urgent' failed: %v", err)
	}

	err = AddTag(testFolder, "backend")
	if err != nil {
		t.Fatalf("AddTag 'backend' failed: %v", err)
	}

	// Verify all tags
	tags, err := GetTagsForFolder(testFolder)
	if err != nil {
		t.Fatalf("GetTagsForFolder failed: %v", err)
	}

	sort.Strings(tags)
	expected := []string{"backend", "urgent", "work"}

	if !reflect.DeepEqual(tags, expected) {
		t.Errorf("Expected tags %v, got %v", expected, tags)
	}
}

func TestAddTagIdempotent(t *testing.T) {
	testFolder, cleanup := setupTestEnv(t)
	defer cleanup()

	// Add same tag multiple times
	for i := 0; i < 3; i++ {
		err := AddTag(testFolder, "work")
		if err != nil {
			t.Fatalf("AddTag iteration %d failed: %v", i, err)
		}
	}

	// Should only have one tag
	tags, err := GetTagsForFolder(testFolder)
	if err != nil {
		t.Fatalf("GetTagsForFolder failed: %v", err)
	}

	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags))
	}
}

func TestRemoveTag(t *testing.T) {
	testFolder, cleanup := setupTestEnv(t)
	defer cleanup()

	// Add tags
	AddTag(testFolder, "work")
	AddTag(testFolder, "urgent")

	// Remove one tag
	err := RemoveTag(testFolder, "work")
	if err != nil {
		t.Fatalf("RemoveTag failed: %v", err)
	}

	// Verify only 'urgent' remains
	tags, err := GetTagsForFolder(testFolder)
	if err != nil {
		t.Fatalf("GetTagsForFolder failed: %v", err)
	}

	if len(tags) != 1 || tags[0] != "urgent" {
		t.Errorf("Expected tags [urgent], got %v", tags)
	}
}

func TestRemoveTagNonExistent(t *testing.T) {
	testFolder, cleanup := setupTestEnv(t)
	defer cleanup()

	// Try to remove tag that doesn't exist
	err := RemoveTag(testFolder, "nonexistent")
	if err == nil {
		t.Error("RemoveTag should fail for non-existent tag")
	}
}

func TestRemoveTagFromNonExistentFolder(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	err := RemoveTag("/nonexistent/folder", "work")
	if err == nil {
		t.Error("RemoveTag should fail for non-existent folder")
	}
}

func TestDeleteTag(t *testing.T) {
	testFolder, cleanup := setupTestEnv(t)
	defer cleanup()

	// Add tag to multiple folders
	tmpDir := filepath.Dir(testFolder)
	folder2 := filepath.Join(tmpDir, "folder2")
	os.MkdirAll(folder2, 0755)

	AddTag(testFolder, "work")
	AddTag(folder2, "work")

	// Delete tag entirely
	err := DeleteTag("work")
	if err != nil {
		t.Fatalf("DeleteTag failed: %v", err)
	}

	// Verify tag is gone from all folders
	tags1, _ := GetTagsForFolder(testFolder)
	tags2, _ := GetTagsForFolder(folder2)

	if len(tags1) != 0 || len(tags2) != 0 {
		t.Error("DeleteTag should remove tag from all folders")
	}

	// Verify tag not in list
	allTags, _ := ListTags()
	if _, exists := allTags["work"]; exists {
		t.Error("Deleted tag should not be in list")
	}
}

func TestDeleteTagNonExistent(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	err := DeleteTag("nonexistent")
	if err == nil {
		t.Error("DeleteTag should fail for non-existent tag")
	}
}

func TestListTags(t *testing.T) {
	testFolder, cleanup := setupTestEnv(t)
	defer cleanup()

	tmpDir := filepath.Dir(testFolder)
	folder2 := filepath.Join(tmpDir, "folder2")
	os.MkdirAll(folder2, 0755)

	// Add various tags
	AddTag(testFolder, "work")
	AddTag(testFolder, "urgent")
	AddTag(folder2, "work")
	AddTag(folder2, "personal")

	tags, err := ListTags()
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	// Check counts
	if tags["work"] != 2 {
		t.Errorf("Expected 2 folders with 'work', got %d", tags["work"])
	}
	if tags["urgent"] != 1 {
		t.Errorf("Expected 1 folder with 'urgent', got %d", tags["urgent"])
	}
	if tags["personal"] != 1 {
		t.Errorf("Expected 1 folder with 'personal', got %d", tags["personal"])
	}
}

func TestListTagsEmpty(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	tags, err := ListTags()
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	if len(tags) != 0 {
		t.Errorf("Expected empty tags map, got %v", tags)
	}
}

func TestListFoldersByTag(t *testing.T) {
	testFolder, cleanup := setupTestEnv(t)
	defer cleanup()

	tmpDir := filepath.Dir(testFolder)
	folder2 := filepath.Join(tmpDir, "folder2")
	folder3 := filepath.Join(tmpDir, "folder3")
	os.MkdirAll(folder2, 0755)
	os.MkdirAll(folder3, 0755)

	// Tag folders
	AddTag(testFolder, "work")
	AddTag(folder2, "work")
	AddTag(folder3, "personal")

	folders, err := ListFoldersByTag("work")
	if err != nil {
		t.Fatalf("ListFoldersByTag failed: %v", err)
	}

	if len(folders) != 2 {
		t.Errorf("Expected 2 folders, got %d", len(folders))
	}

	// Check if both folders are in the list
	foundFolder1 := false
	foundFolder2 := false
	for _, f := range folders {
		if f == testFolder {
			foundFolder1 = true
		}
		if f == folder2 {
			foundFolder2 = true
		}
	}

	if !foundFolder1 || !foundFolder2 {
		t.Error("Not all expected folders found in list")
	}
}

func TestListFoldersByTagEmpty(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	folders, err := ListFoldersByTag("nonexistent")
	if err != nil {
		t.Fatalf("ListFoldersByTag failed: %v", err)
	}

	if len(folders) != 0 {
		t.Errorf("Expected empty folders list, got %v", folders)
	}
}

func TestGetTagsForFolder(t *testing.T) {
	testFolder, cleanup := setupTestEnv(t)
	defer cleanup()

	// Add multiple tags
	AddTag(testFolder, "work")
	AddTag(testFolder, "urgent")
	AddTag(testFolder, "backend")

	tags, err := GetTagsForFolder(testFolder)
	if err != nil {
		t.Fatalf("GetTagsForFolder failed: %v", err)
	}

	sort.Strings(tags)
	expected := []string{"backend", "urgent", "work"}

	if !reflect.DeepEqual(tags, expected) {
		t.Errorf("Expected %v, got %v", expected, tags)
	}
}

func TestGetTagsForFolderEmpty(t *testing.T) {
	testFolder, cleanup := setupTestEnv(t)
	defer cleanup()

	tags, err := GetTagsForFolder(testFolder)
	if err != nil {
		t.Fatalf("GetTagsForFolder failed: %v", err)
	}

	if len(tags) != 0 {
		t.Errorf("Expected empty tags list, got %v", tags)
	}
}

func TestGetTagsForNonExistentFolder(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	tags, err := GetTagsForFolder("/nonexistent/folder")
	if err != nil {
		t.Fatalf("GetTagsForFolder failed: %v", err)
	}

	// Should return empty list, not error
	if len(tags) != 0 {
		t.Errorf("Expected empty tags list for non-existent folder, got %v", tags)
	}
}

func TestConcurrentTagOperations(t *testing.T) {
	testFolder, cleanup := setupTestEnv(t)
	defer cleanup()

	tmpDir := filepath.Dir(testFolder)
	folder2 := filepath.Join(tmpDir, "folder2")
	os.MkdirAll(folder2, 0755)

	// Add tags sequentially first to ensure they exist
	if err := AddTag(testFolder, "work"); err != nil {
		t.Fatalf("Failed to add tag to folder1: %v", err)
	}
	if err := AddTag(folder2, "work"); err != nil {
		t.Fatalf("Failed to add tag to folder2: %v", err)
	}

	// Test concurrent read operations don't crash
	done := make(chan bool, 4)

	go func() {
		ListTags()
		done <- true
	}()

	go func() {
		ListFoldersByTag("work")
		done <- true
	}()

	go func() {
		GetTagsForFolder(testFolder)
		done <- true
	}()

	go func() {
		GetTagsForFolder(folder2)
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}

	// Verify data integrity after concurrent reads
	tags, err := ListTags()
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}
	if tags["work"] != 2 {
		t.Errorf("Expected 2 folders with 'work' after concurrent ops, got %d", tags["work"])
	}
}

func BenchmarkAddTag(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "scope-tag-bench-*")
	defer os.RemoveAll(tmpDir)

	testFolder := filepath.Join(tmpDir, "test")
	os.MkdirAll(testFolder, 0755)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer func() {
		db.Close()
		db.ResetForTesting()
		os.Setenv("HOME", originalHome)
	}()

	db.InitDB()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AddTag(testFolder, "work")
	}
}

func BenchmarkListTags(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "scope-tag-bench-*")
	defer os.RemoveAll(tmpDir)

	testFolder := filepath.Join(tmpDir, "test")
	os.MkdirAll(testFolder, 0755)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer func() {
		db.Close()
		db.ResetForTesting()
		os.Setenv("HOME", originalHome)
	}()

	db.InitDB()

	// Setup: add some tags
	for i := 0; i < 10; i++ {
		folder := filepath.Join(tmpDir, filepath.Join("folder", string(rune(i))))
		os.MkdirAll(folder, 0755)
		AddTag(folder, "work")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ListTags()
	}
}
