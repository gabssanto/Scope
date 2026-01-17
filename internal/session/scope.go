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
	// Get all folders for the tag
	folders, err := tag.ListFoldersByTag(tagName)
	if err != nil {
		return fmt.Errorf("failed to list folders: %w", err)
	}

	if len(folders) == 0 {
		return fmt.Errorf("no folders found with tag: %s", tagName)
	}

	// Create temp directory
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("scope-%s-", tagName))
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

	fmt.Printf("Scope session started with tag '%s'\n", tagName)
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
		fmt.Sprintf("SCOPE_SESSION=%s", tagName),
		fmt.Sprintf("SCOPE_WORKSPACE=%s", tempDir),
	)

	// Run the shell
	if err := cmd.Run(); err != nil {
		// Check if it's an exit status error
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				// Exit with the same status code as the shell
				os.Exit(status.ExitStatus())
			}
		}
		return fmt.Errorf("failed to run shell: %w", err)
	}

	fmt.Println("\nScope session ended. Workspace cleaned up.")
	return nil
}
