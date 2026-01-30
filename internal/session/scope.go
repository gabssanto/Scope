package session

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/gabssanto/Scope/internal/tag"
)

// StartSession creates a temporary workspace with symlinks and spawns a shell
func StartSession(tagName string) error {
	return StartMultiTagSession([]string{tagName})
}

// StartMultiTagSession creates a workspace with folders from multiple tags
func StartMultiTagSession(tagNames []string) error {
	if len(tagNames) == 0 {
		return fmt.Errorf("no tags provided")
	}

	// Collect folders from all tags (use map to dedupe)
	folderSet := make(map[string]bool)
	for _, tagName := range tagNames {
		folders, err := tag.ListFoldersByTag(tagName)
		if err != nil {
			return fmt.Errorf("failed to list folders for tag '%s': %w", tagName, err)
		}
		for _, f := range folders {
			folderSet[f] = true
		}
	}

	if len(folderSet) == 0 {
		return fmt.Errorf("no folders found with tags: %v", tagNames)
	}

	// Convert map to slice
	folders := make([]string, 0, len(folderSet))
	for f := range folderSet {
		folders = append(folders, f)
	}

	// Create session name from tags
	sessionName := tagNames[0]
	if len(tagNames) > 1 {
		sessionName = fmt.Sprintf("%s+%d", tagNames[0], len(tagNames)-1)
	}

	// Create temp directory
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("scope-%s-", sessionName))
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Cleanup temp directory on exit
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to cleanup temp directory %s: %v\n", tempDir, err)
		}
	}()

	// Create symlinks for all folders
	for _, folder := range folders {
		// Use the basename of the folder as the symlink name
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
			linkPath = fmt.Sprintf("%s-%d", originalLinkPath, counter)
			counter++
		}

		// Create symlink
		if err := os.Symlink(folder, linkPath); err != nil {
			return fmt.Errorf("failed to create symlink for %s: %w", folder, err)
		}
	}

	if len(tagNames) == 1 {
		fmt.Printf("Scope session started with tag '%s'\n", tagNames[0])
	} else {
		fmt.Printf("Scope session started with tags: %v\n", tagNames)
	}
	fmt.Printf("Workspace: %s\n", tempDir)
	fmt.Printf("Folders: %d\n\n", len(folders))
	fmt.Println("Type 'exit' to leave the scoped session")
	fmt.Println("---")

	// Get user's shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	// Spawn shell in the temp directory
	cmd := exec.Command(shell)
	cmd.Dir = tempDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment variables
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("SCOPE_SESSION=%s", sessionName),
		fmt.Sprintf("SCOPE_WORKSPACE=%s", tempDir),
	)

	// Run the shell
	shellErr := cmd.Run()

	// Cleanup happens here via defer before we potentially exit

	if shellErr != nil {
		// Check if it's an exit status error (user exited shell with non-zero)
		if exitErr, ok := shellErr.(*exec.ExitError); ok {
			if _, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				// Return nil - the shell exited normally (possibly with non-zero)
				// The defer cleanup will run, then main() will exit with 0
				// We don't propagate shell exit codes as errors
				fmt.Println("\nScope session ended. Workspace cleaned up.")
				return nil
			}
		}
		return fmt.Errorf("failed to run shell: %w", shellErr)
	}

	fmt.Println("\nScope session ended. Workspace cleaned up.")
	return nil
}
