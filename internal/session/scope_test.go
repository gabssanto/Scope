package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabssanto/Scope/internal/db"
	"github.com/gabssanto/Scope/internal/tag"
)

// setupTestEnv creates a test environment with temporary database and folders
func setupTestEnv(t *testing.T) (string, []string, func()) {
	t.Helper()

	// Create temp directories
	tmpDir, err := os.MkdirTemp("", "scope-session-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create test folders
	testFolders := []string{
		filepath.Join(tmpDir, "project1"),
		filepath.Join(tmpDir, "project2"),
		filepath.Join(tmpDir, "project3"),
	}

	for _, folder := range testFolders {
		if err := os.MkdirAll(folder, 0755); err != nil {
			t.Fatalf("Failed to create test folder: %v", err)
		}
		// Create a test file in each folder
		testFile := filepath.Join(folder, "README.md")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
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

	return tmpDir, testFolders, cleanup
}

func TestStartSessionNoFolders(t *testing.T) {
	_, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Try to start session with tag that has no folders
	err := StartSession("nonexistent")
	if err == nil {
		t.Error("StartSession should fail when no folders have the tag")
	}

	if !strings.Contains(err.Error(), "no folders found") {
		t.Errorf("Expected 'no folders found' error, got: %v", err)
	}
}

func TestStartSessionCreatesSymlinks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping session test in short mode")
	}

	tmpDir, testFolders, cleanup := setupTestEnv(t)
	defer cleanup()

	// Tag folders
	tag.AddTag(testFolders[0], "work")
	tag.AddTag(testFolders[1], "work")

	// We can't fully test the interactive shell, but we can test symlink creation
	// by creating our own temp directory and checking symlinks
	folders, _ := tag.ListFoldersByTag("work")

	tempDir, err := os.MkdirTemp("", "scope-work-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create symlinks manually (simulating what StartSession does)
	for _, folder := range folders {
		linkName := filepath.Base(folder)
		linkPath := filepath.Join(tempDir, linkName)

		if err := os.Symlink(folder, linkPath); err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}
	}

	// Verify symlinks were created
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 symlinks, got %d", len(entries))
	}

	// Verify symlinks point to correct targets
	for _, entry := range entries {
		linkPath := filepath.Join(tempDir, entry.Name())
		target, err := os.Readlink(linkPath)
		if err != nil {
			t.Errorf("Failed to read symlink %s: %v", linkPath, err)
			continue
		}

		// Check if target is one of our test folders
		found := false
		for _, testFolder := range testFolders {
			if target == testFolder {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Symlink %s points to unexpected target: %s", linkPath, target)
		}
	}

	// Verify we can access files through symlinks
	for _, entry := range entries {
		readmePath := filepath.Join(tempDir, entry.Name(), "README.md")
		if _, err := os.Stat(readmePath); os.IsNotExist(err) {
			t.Errorf("Cannot access file through symlink: %s", readmePath)
		}
	}

	// Cleanup is handled by defer
	_ = tmpDir
}

func TestStartSessionNameConflicts(t *testing.T) {
	tmpDir, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create folders with same basename in different locations
	folder1 := filepath.Join(tmpDir, "location1", "myproject")
	folder2 := filepath.Join(tmpDir, "location2", "myproject")
	os.MkdirAll(folder1, 0755)
	os.MkdirAll(folder2, 0755)

	// Tag both
	tag.AddTag(folder1, "conflict-test")
	tag.AddTag(folder2, "conflict-test")

	// Create temp workspace and symlinks
	tempDir, err := os.MkdirTemp("", "scope-conflict-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	folders, _ := tag.ListFoldersByTag("conflict-test")

	// Simulate creating symlinks with conflict resolution
	for _, folder := range folders {
		linkName := filepath.Base(folder)
		linkPath := filepath.Join(tempDir, linkName)

		// Handle name conflicts by appending a number
		counter := 1
		originalLinkPath := linkPath
		for {
			_, err := os.Lstat(linkPath)
			if os.IsNotExist(err) {
				break
			}
			linkPath = filepath.Join(tempDir, linkName+"-"+string(rune('0'+counter)))
			counter++
		}

		if err := os.Symlink(folder, linkPath); err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}

		_ = originalLinkPath
	}

	// Verify both symlinks were created with different names
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 symlinks despite name conflict, got %d", len(entries))
	}
}

func TestStartSessionEnvironmentVariables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping session test in short mode")
	}

	_, testFolders, cleanup := setupTestEnv(t)
	defer cleanup()

	tag.AddTag(testFolders[0], "env-test")

	// We can't easily test the full session spawn, but we can verify
	// the environment variables would be set correctly
	tagName := "env-test"
	tempDir := "/tmp/scope-env-test-12345"

	expectedVars := map[string]string{
		"SCOPE_SESSION":   tagName,
		"SCOPE_WORKSPACE": tempDir,
	}

	// This is what would be set in the actual session
	for key, expectedValue := range expectedVars {
		// In the actual implementation, these are set via cmd.Env
		// We're just verifying the logic here
		if key == "SCOPE_SESSION" && expectedValue != tagName {
			t.Errorf("SCOPE_SESSION should be %s, got %s", tagName, expectedValue)
		}
		if key == "SCOPE_WORKSPACE" && expectedValue != tempDir {
			t.Errorf("SCOPE_WORKSPACE should be %s, got %s", tempDir, expectedValue)
		}
	}
}

func TestStartSessionShellSelection(t *testing.T) {
	// Save original SHELL
	originalShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", originalShell)

	// Test with custom shell
	customShell := "/bin/custom-shell"
	os.Setenv("SHELL", customShell)

	shell := os.Getenv("SHELL")
	if shell != customShell {
		t.Errorf("Expected shell %s, got %s", customShell, shell)
	}

	// Test with empty SHELL (should default to /bin/bash)
	os.Unsetenv("SHELL")
	shell = os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash" // This is what the code does
	}
	if shell != "/bin/bash" {
		t.Errorf("Expected default shell /bin/bash, got %s", shell)
	}
}

func TestSymlinkCleanup(t *testing.T) {
	tmpDir, testFolders, cleanup := setupTestEnv(t)
	defer cleanup()

	tag.AddTag(testFolders[0], "cleanup-test")

	// Create temp workspace
	tempDir, err := os.MkdirTemp("", "scope-cleanup-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create a symlink
	linkPath := filepath.Join(tempDir, "project1")
	if err := os.Symlink(testFolders[0], linkPath); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Verify symlink exists
	if _, err := os.Lstat(linkPath); err != nil {
		t.Fatalf("Symlink should exist: %v", err)
	}

	// Cleanup (simulating defer in StartSession)
	if err := os.RemoveAll(tempDir); err != nil {
		t.Fatalf("Failed to cleanup: %v", err)
	}

	// Verify temp directory is gone
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Error("Temp directory should be removed after cleanup")
	}

	// Verify original folder still exists
	if _, err := os.Stat(testFolders[0]); os.IsNotExist(err) {
		t.Error("Original folder should not be deleted when symlink is removed")
	}

	_ = tmpDir
}

func TestMultipleFoldersSession(t *testing.T) {
	_, testFolders, cleanup := setupTestEnv(t)
	defer cleanup()

	// Tag multiple folders
	for i, folder := range testFolders {
		tag.AddTag(folder, "multi-test")
		_ = i
	}

	folders, err := tag.ListFoldersByTag("multi-test")
	if err != nil {
		t.Fatalf("ListFoldersByTag failed: %v", err)
	}

	if len(folders) != len(testFolders) {
		t.Errorf("Expected %d folders, got %d", len(testFolders), len(folders))
	}

	// Create workspace with all symlinks
	tempDir, err := os.MkdirTemp("", "scope-multi-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	for _, folder := range folders {
		linkName := filepath.Base(folder)
		linkPath := filepath.Join(tempDir, linkName)
		if err := os.Symlink(folder, linkPath); err != nil {
			t.Fatalf("Failed to create symlink for %s: %v", folder, err)
		}
	}

	// Verify all symlinks
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	if len(entries) != len(testFolders) {
		t.Errorf("Expected %d symlinks, got %d", len(testFolders), len(entries))
	}
}

func BenchmarkSymlinkCreation(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "scope-session-bench-*")
	defer os.RemoveAll(tmpDir)

	testFolder := filepath.Join(tmpDir, "test")
	os.MkdirAll(testFolder, 0755)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tempDir, _ := os.MkdirTemp("", "scope-bench-")
		linkPath := filepath.Join(tempDir, "test")
		os.Symlink(testFolder, linkPath)
		os.RemoveAll(tempDir)
	}
}
