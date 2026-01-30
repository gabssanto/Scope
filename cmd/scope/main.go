package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gabssanto/Scope/internal/db"
	"github.com/gabssanto/Scope/internal/scan"
	"github.com/gabssanto/Scope/internal/session"
	"github.com/gabssanto/Scope/internal/tag"
)

// Version is set at build time via ldflags
var Version = "dev"

const usage = `Scope - Fast folder navigation with tags

Usage:
  scope tag <path> <tag>        Tag a folder (use . for current directory)
  scope untag <path> <tag>      Remove a tag from a folder
  scope list [tag]              List all tags or folders with specific tag
  scope start <tag>             Start a scoped session
  scope scan [path]             Scan for .scope files and apply tags
  scope remove-tag <tag>        Delete a tag entirely
  scope help                    Show this help message
  scope version                 Show version information

Sessions:
  When you run 'scope start <tag>', a new shell opens in a temporary
  workspace containing symlinks to all folders with that tag.

  To exit a session, simply type 'exit' or press Ctrl+D.
  The temporary workspace is automatically cleaned up when you exit.

Examples:
  scope tag . work              Tag current directory with 'work'
  scope tag ~/projects/app dev  Tag a specific folder
  scope list                    Show all tags
  scope list work               Show all folders tagged 'work'
  scope start work              Open scoped session with 'work' folders
  scope untag . work            Remove 'work' tag from current directory
  scope remove-tag old          Delete 'old' tag entirely
`

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Initialize database
	if err := db.InitDB(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Parse command
	if len(os.Args) < 2 {
		fmt.Print(usage)
		return nil
	}

	command := os.Args[1]

	switch command {
	case "tag":
		return handleTag()
	case "untag":
		return handleUntag()
	case "list":
		return handleList()
	case "start":
		return handleStart()
	case "scan":
		return handleScan()
	case "remove-tag":
		return handleRemoveTag()
	case "help", "--help", "-h":
		fmt.Print(usage)
		return nil
	case "version", "--version", "-v":
		fmt.Printf("scope version %s\n", Version)
		return nil
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		fmt.Print(usage)
		return nil
	}
}

func handleTag() error {
	if len(os.Args) < 4 {
		return fmt.Errorf("usage: scope tag <path> <tag>")
	}

	path := os.Args[2]
	tagName := os.Args[3]

	// Resolve path
	absPath, err := resolvePath(path)
	if err != nil {
		return err
	}

	// Add tag
	if err := tag.AddTag(absPath, tagName); err != nil {
		return err
	}

	fmt.Printf("Tagged '%s' with '%s'\n", absPath, tagName)
	return nil
}

func handleUntag() error {
	if len(os.Args) < 4 {
		return fmt.Errorf("usage: scope untag <path> <tag>")
	}

	path := os.Args[2]
	tagName := os.Args[3]

	// Resolve path
	absPath, err := resolvePath(path)
	if err != nil {
		return err
	}

	// Remove tag
	if err := tag.RemoveTag(absPath, tagName); err != nil {
		return err
	}

	fmt.Printf("Removed tag '%s' from '%s'\n", tagName, absPath)
	return nil
}

func handleList() error {
	// If tag name provided, list folders for that tag
	if len(os.Args) >= 3 {
		tagName := os.Args[2]
		folders, err := tag.ListFoldersByTag(tagName)
		if err != nil {
			return err
		}

		if len(folders) == 0 {
			fmt.Printf("No folders found with tag '%s'\n", tagName)
			return nil
		}

		fmt.Printf("Folders tagged with '%s':\n", tagName)
		for _, folder := range folders {
			fmt.Printf("  %s\n", folder)
		}
		fmt.Printf("\nTotal: %d folders\n", len(folders))
		return nil
	}

	// Otherwise, list all tags
	tags, err := tag.ListTags()
	if err != nil {
		return err
	}

	if len(tags) == 0 {
		fmt.Println("No tags found. Use 'scope tag <path> <tag>' to create one.")
		return nil
	}

	// Sort tags by name
	names := make([]string, 0, len(tags))
	for name := range tags {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Println("Tags:")
	for _, name := range names {
		count := tags[name]
		plural := ""
		if count != 1 {
			plural = "s"
		}
		fmt.Printf("  %-20s %d folder%s\n", name, count, plural)
	}

	fmt.Printf("\nTotal: %d tags\n", len(tags))
	return nil
}

func handleStart() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: scope start <tag>")
	}

	tagName := os.Args[2]
	return session.StartSession(tagName)
}

func handleRemoveTag() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: scope remove-tag <tag>")
	}

	tagName := os.Args[2]

	if err := tag.DeleteTag(tagName); err != nil {
		return err
	}

	fmt.Printf("Removed tag '%s'\n", tagName)
	return nil
}

func handleScan() error {
	// Default to current directory
	path := "."
	if len(os.Args) >= 3 {
		path = os.Args[2]
	}

	// Resolve to absolute path
	absPath, err := resolvePath(path)
	if err != nil {
		return err
	}

	// Verify it's a directory
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("cannot access path: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}

	return scan.RunScan(absPath)
}

// resolvePath converts a path (including .) to an absolute path
func resolvePath(path string) (string, error) {
	// Handle current directory
	if path == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		return cwd, nil
	}

	// Expand home directory
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(homeDir, path[1:])
	}

	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	return absPath, nil
}
