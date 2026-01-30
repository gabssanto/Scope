package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/huh"
	"gopkg.in/yaml.v3"

	"github.com/gabssanto/Scope/internal/completions"
	"github.com/gabssanto/Scope/internal/db"
	"github.com/gabssanto/Scope/internal/scan"
	"github.com/gabssanto/Scope/internal/session"
	"github.com/gabssanto/Scope/internal/tag"
	"github.com/gabssanto/Scope/internal/update"
)

// Version is set at build time via ldflags
var Version = "dev"

const usage = `Scope - Fast folder navigation with tags

Usage:
  scope tag <path> <tag>        Tag a folder (use . for current directory)
  scope untag <path> <tag>      Remove a tag from a folder
  scope tags <path>             Show all tags for a folder
  scope list [tag]              List all tags or folders with specific tag
  scope start <tag>             Start a scoped session
  scope scan [path]             Scan for .scope files and apply tags
  scope go <tag>                Jump to a tagged folder (outputs path)
  scope pick [tag]              Interactive folder picker
  scope open <tag>              Open tagged folder(s) in file manager
  scope edit <tag>              Open tagged folder(s) in editor
  scope each <tag> <cmd>        Run command in each tagged folder
  scope status <tag>            Git status across tagged folders
  scope pull <tag>              Git pull across tagged folders
  scope rename <old> <new>      Rename a tag
  scope remove-tag <tag>        Delete a tag entirely
  scope prune [--dry-run]       Remove folders that no longer exist
  scope export                  Export all tags to YAML
  scope import <file>           Import tags from YAML file
  scope update [--check]        Update to latest version
  scope completions <shell>     Generate shell completions (bash/zsh/fish)
  scope debug                   Show debug information
  scope help                    Show this help message
  scope version                 Show version information

Sessions:
  When you run 'scope start <tag>', a new shell opens in a temporary
  workspace containing symlinks to all folders with that tag.

  To exit a session, simply type 'exit' or press Ctrl+D.
  The temporary workspace is automatically cleaned up when you exit.

Navigation:
  'scope go' outputs a path for shell integration. Add to your .bashrc/.zshrc:
    sg() { cd "$(scope go "$@")" 2>/dev/null || scope go "$@"; }

Examples:
  scope tag . work              Tag current directory with 'work'
  scope tag ~/projects/app dev  Tag a specific folder
  scope tags .                  Show tags for current directory
  scope list                    Show all tags
  scope list work               Show all folders tagged 'work'
  scope start work              Open scoped session with 'work' folders
  scope go work                 Output path to 'work' folder (for cd)
  scope open work               Open 'work' folders in Finder/Explorer
  scope edit work               Open 'work' folders in $EDITOR
  scope each work "git status"  Run git status in each 'work' folder
  scope each work -p "go test"  Run tests in parallel across folders
  scope untag . work            Remove 'work' tag from current directory
  scope rename old new          Rename 'old' tag to 'new'
  scope remove-tag old          Delete 'old' tag entirely
  scope prune --dry-run         Preview folders to be removed
`

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// showUpdateNotice displays update notification if available
func showUpdateNotice() {
	// Skip for certain commands that output paths (for shell integration)
	if len(os.Args) >= 2 {
		cmd := os.Args[1]
		// Skip for commands where stdout is used for data
		if cmd == "go" || cmd == "version" || cmd == "--version" || cmd == "-v" {
			return
		}
	}

	// Check if running in a non-interactive context
	if os.Getenv("SCOPE_NO_UPDATE_CHECK") != "" {
		return
	}

	notice := update.GetUpdateNotice(Version)
	if notice != "" {
		fmt.Fprint(os.Stderr, notice)
	}
}

func run() error {
	// Initialize database
	if err := db.InitDB(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Show update notice at the end (only for interactive commands)
	defer showUpdateNotice()

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
	case "tags":
		return handleTags()
	case "list":
		return handleList()
	case "start":
		return handleStart()
	case "scan":
		return handleScan()
	case "go":
		return handleGo()
	case "pick":
		return handlePick()
	case "open":
		return handleOpen()
	case "edit":
		return handleEdit()
	case "each":
		return handleEach()
	case "status":
		return handleStatus()
	case "pull":
		return handlePull()
	case "rename":
		return handleRename()
	case "remove-tag":
		return handleRemoveTag()
	case "prune":
		return handlePrune()
	case "export":
		return handleExport()
	case "import":
		return handleImport()
	case "update":
		return handleUpdate()
	case "completions":
		return handleCompletions()
	case "debug":
		return handleDebug()
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

func handleTags() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: scope tags <path>")
	}

	path := os.Args[2]

	// Resolve path
	absPath, err := resolvePath(path)
	if err != nil {
		return err
	}

	tags, err := tag.GetTagsForFolder(absPath)
	if err != nil {
		return err
	}

	if len(tags) == 0 {
		fmt.Printf("No tags found for '%s'\n", absPath)
		return nil
	}

	fmt.Printf("Tags for '%s':\n", absPath)
	for _, t := range tags {
		fmt.Printf("  %s\n", t)
	}
	return nil
}

func handleRename() error {
	if len(os.Args) < 4 {
		return fmt.Errorf("usage: scope rename <old> <new>")
	}

	oldName := os.Args[2]
	newName := os.Args[3]

	if err := tag.RenameTag(oldName, newName); err != nil {
		return err
	}

	fmt.Printf("Renamed tag '%s' to '%s'\n", oldName, newName)
	return nil
}

func handlePrune() error {
	dryRun := false
	if len(os.Args) >= 3 && (os.Args[2] == "--dry-run" || os.Args[2] == "-n") {
		dryRun = true
	}

	result, err := tag.Prune(dryRun)
	if err != nil {
		return err
	}

	if result.RemovedCount == 0 {
		fmt.Println("No stale folders found. Everything is clean!")
		return nil
	}

	if dryRun {
		fmt.Printf("Would remove %d stale folder(s):\n", result.RemovedCount)
	} else {
		fmt.Printf("Removed %d stale folder(s):\n", result.RemovedCount)
	}

	for _, path := range result.RemovedFolders {
		fmt.Printf("  %s\n", path)
	}

	return nil
}

// ExportData represents the structure of exported data
type ExportData struct {
	Version int                 `yaml:"version"`
	Tags    map[string][]string `yaml:"tags"`
}

func handleExport() error {
	tags, err := tag.ListTags()
	if err != nil {
		return err
	}

	if len(tags) == 0 {
		fmt.Fprintln(os.Stderr, "No tags to export")
		return nil
	}

	data := ExportData{
		Version: 1,
		Tags:    make(map[string][]string),
	}

	// Get folders for each tag
	for tagName := range tags {
		folders, err := tag.ListFoldersByTag(tagName)
		if err != nil {
			return fmt.Errorf("failed to get folders for tag '%s': %w", tagName, err)
		}
		data.Tags[tagName] = folders
	}

	// Marshal to YAML
	output, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	fmt.Print(string(output))
	return nil
}

func handleImport() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: scope import <file>")
	}

	filePath := os.Args[2]

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse YAML
	var data ExportData
	if err := yaml.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	if len(data.Tags) == 0 {
		fmt.Println("No tags found in import file")
		return nil
	}

	// Import tags
	imported := 0
	skipped := 0

	for tagName, folders := range data.Tags {
		for _, folder := range folders {
			// Check if folder exists
			if _, err := os.Stat(folder); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Skipping non-existent folder: %s\n", folder)
				skipped++
				continue
			}

			if err := tag.AddTag(folder, tagName); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to add tag '%s' to %s: %v\n", tagName, folder, err)
				continue
			}
			imported++
		}
	}

	fmt.Printf("Imported %d tag assignments (%d skipped)\n", imported, skipped)
	return nil
}

func handleDebug() error {
	homeDir, _ := os.UserHomeDir()
	dbPath := filepath.Join(homeDir, ".config", "scope", "scope.db")

	fmt.Println("Scope Debug Information")
	fmt.Println("=======================")
	fmt.Printf("Version:     %s\n", Version)
	fmt.Printf("OS/Arch:     %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Go version:  %s\n", runtime.Version())
	fmt.Printf("Database:    %s\n", dbPath)

	// Check if db exists
	if _, err := os.Stat(dbPath); err == nil {
		info, _ := os.Stat(dbPath)
		fmt.Printf("DB size:     %d bytes\n", info.Size())
	} else {
		fmt.Printf("DB size:     (not found)\n")
	}

	// Shell info
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "(unknown)"
	}
	fmt.Printf("Shell:       %s\n", shell)

	// Scope session info
	scopeSession := os.Getenv("SCOPE_SESSION")
	if scopeSession != "" {
		fmt.Printf("In session:  %s\n", scopeSession)
		fmt.Printf("Workspace:   %s\n", os.Getenv("SCOPE_WORKSPACE"))
	}

	// Stats
	tags, _ := tag.ListTags()
	totalFolders := 0
	for _, count := range tags {
		totalFolders += count
	}
	fmt.Printf("\nStats:\n")
	fmt.Printf("  Tags:      %d\n", len(tags))
	fmt.Printf("  Folders:   %d tag assignments\n", totalFolders)

	return nil
}

func handleGo() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: scope go <tag>")
	}

	tagName := os.Args[2]

	folders, err := tag.ListFoldersByTag(tagName)
	if err != nil {
		return err
	}

	if len(folders) == 0 {
		return fmt.Errorf("no folders found with tag '%s'", tagName)
	}

	// Single folder - just output the path
	if len(folders) == 1 {
		fmt.Println(folders[0])
		return nil
	}

	// Multiple folders - show picker
	fmt.Fprintf(os.Stderr, "Multiple folders found for '%s':\n", tagName)
	for i, folder := range folders {
		fmt.Fprintf(os.Stderr, "  [%d] %s\n", i+1, folder)
	}
	fmt.Fprintf(os.Stderr, "\nSelect folder (1-%d): ", len(folders))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(folders) {
		return fmt.Errorf("invalid selection: %s", input)
	}

	fmt.Println(folders[choice-1])
	return nil
}

func handlePick() error {
	var folders []string
	var err error

	// If tag provided, filter by tag
	if len(os.Args) >= 3 {
		tagName := os.Args[2]
		folders, err = tag.ListFoldersByTag(tagName)
		if err != nil {
			return err
		}
		if len(folders) == 0 {
			return fmt.Errorf("no folders found with tag '%s'", tagName)
		}
	} else {
		// Get all folders from all tags
		folders, err = tag.ListAllFolders()
		if err != nil {
			return err
		}
		if len(folders) == 0 {
			fmt.Println("No tagged folders found. Use 'scope tag <path> <tag>' to tag folders.")
			return nil
		}
	}

	// Build options for select
	options := make([]huh.Option[string], len(folders))
	for i, folder := range folders {
		folderName := filepath.Base(folder)
		options[i] = huh.NewOption(fmt.Sprintf("%s (%s)", folderName, folder), folder)
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a folder").
				Description("Use / to filter, enter to select").
				Options(options...).
				Value(&selected),
		),
	)

	err = form.Run()
	if err != nil {
		return fmt.Errorf("selection canceled: %w", err)
	}

	// Output the selected path
	fmt.Println(selected)
	return nil
}

func handleOpen() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: scope open <tag>")
	}

	tagName := os.Args[2]

	folders, err := tag.ListFoldersByTag(tagName)
	if err != nil {
		return err
	}

	if len(folders) == 0 {
		return fmt.Errorf("no folders found with tag '%s'", tagName)
	}

	// Determine the open command based on OS
	var openCmd string
	switch runtime.GOOS {
	case "darwin":
		openCmd = "open"
	case "linux":
		openCmd = "xdg-open"
	case "windows":
		openCmd = "explorer"
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// Open each folder
	for _, folder := range folders {
		cmd := exec.Command(openCmd, folder)
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to open '%s': %v\n", folder, err)
			continue
		}
		fmt.Printf("Opened: %s\n", folder)
	}

	return nil
}

func handleEdit() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: scope edit <tag>")
	}

	tagName := os.Args[2]

	folders, err := tag.ListFoldersByTag(tagName)
	if err != nil {
		return err
	}

	if len(folders) == 0 {
		return fmt.Errorf("no folders found with tag '%s'", tagName)
	}

	// Determine editor
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		// Try common editors
		for _, e := range []string{"code", "vim", "nano"} {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	if editor == "" {
		return fmt.Errorf("no editor found. Set $EDITOR or $VISUAL environment variable")
	}

	// Open each folder in editor
	for _, folder := range folders {
		cmd := exec.Command(editor, folder)
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to open '%s' in %s: %v\n", folder, editor, err)
			continue
		}
		fmt.Printf("Opened in %s: %s\n", editor, folder)
	}

	return nil
}

func handleEach() error {
	if len(os.Args) < 4 {
		return fmt.Errorf("usage: scope each <tag> [-p] <command>")
	}

	tagName := os.Args[2]
	parallel := false
	cmdStart := 3

	// Check for parallel flag
	if os.Args[3] == "-p" || os.Args[3] == "--parallel" {
		parallel = true
		cmdStart = 4
		if len(os.Args) < 5 {
			return fmt.Errorf("usage: scope each <tag> [-p] <command>")
		}
	}

	// Join remaining args as command
	command := strings.Join(os.Args[cmdStart:], " ")

	folders, err := tag.ListFoldersByTag(tagName)
	if err != nil {
		return err
	}

	if len(folders) == 0 {
		return fmt.Errorf("no folders found with tag '%s'", tagName)
	}

	if parallel {
		return runEachParallel(folders, command)
	}
	return runEachSequential(folders, command)
}

func runEachSequential(folders []string, command string) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	successCount := 0
	failCount := 0

	for _, folder := range folders {
		folderName := filepath.Base(folder)
		fmt.Printf("\n\033[1;34m[%s]\033[0m %s\n", folderName, folder)
		fmt.Println(strings.Repeat("-", 40))

		cmd := exec.Command(shell, "-c", command)
		cmd.Dir = folder
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "\033[1;31mError:\033[0m %v\n", err)
			failCount++
		} else {
			successCount++
		}
	}

	fmt.Printf("\n\033[1mSummary:\033[0m %d succeeded, %d failed\n", successCount, failCount)
	return nil
}

func runEachParallel(folders []string, command string) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	type result struct {
		folder string
		output string
		err    error
	}

	results := make(chan result, len(folders))
	var wg sync.WaitGroup

	for _, folder := range folders {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()

			var stdout, stderr bytes.Buffer
			cmd := exec.Command(shell, "-c", command)
			cmd.Dir = f
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			output := stdout.String()
			if stderr.Len() > 0 {
				output += stderr.String()
			}

			results <- result{folder: f, output: output, err: err}
		}(folder)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and print results
	successCount := 0
	failCount := 0

	for r := range results {
		folderName := filepath.Base(r.folder)
		fmt.Printf("\n\033[1;34m[%s]\033[0m %s\n", folderName, r.folder)
		fmt.Println(strings.Repeat("-", 40))

		if r.output != "" {
			fmt.Print(r.output)
		}

		if r.err != nil {
			fmt.Fprintf(os.Stderr, "\033[1;31mError:\033[0m %v\n", r.err)
			failCount++
		} else {
			successCount++
		}
	}

	fmt.Printf("\n\033[1mSummary:\033[0m %d succeeded, %d failed\n", successCount, failCount)
	return nil
}

func handleStatus() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: scope status <tag>")
	}

	tagName := os.Args[2]

	folders, err := tag.ListFoldersByTag(tagName)
	if err != nil {
		return err
	}

	if len(folders) == 0 {
		return fmt.Errorf("no folders found with tag '%s'", tagName)
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	for _, folder := range folders {
		// Check if it's a git repo
		gitDir := filepath.Join(folder, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			continue
		}

		folderName := filepath.Base(folder)

		// Get git status
		cmd := exec.Command(shell, "-c", "git status -s")
		cmd.Dir = folder
		output, _ := cmd.Output()

		if len(output) > 0 {
			fmt.Printf("\033[1;33m[%s]\033[0m %s\n", folderName, folder)
			fmt.Print(string(output))
			fmt.Println()
		}
	}

	return nil
}

func handlePull() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: scope pull <tag>")
	}

	tagName := os.Args[2]

	folders, err := tag.ListFoldersByTag(tagName)
	if err != nil {
		return err
	}

	if len(folders) == 0 {
		return fmt.Errorf("no folders found with tag '%s'", tagName)
	}

	// Filter to git repos only
	var gitFolders []string
	for _, folder := range folders {
		gitDir := filepath.Join(folder, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			gitFolders = append(gitFolders, folder)
		}
	}

	if len(gitFolders) == 0 {
		fmt.Println("No git repositories found with this tag")
		return nil
	}

	fmt.Printf("Pulling %d repositories...\n", len(gitFolders))
	return runEachParallel(gitFolders, "git pull")
}

func handleCompletions() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: scope completions <shell>\nSupported shells: bash, zsh, fish")
	}

	shell := os.Args[2]
	script, err := completions.Generate(shell)
	if err != nil {
		return err
	}

	fmt.Print(script)
	return nil
}

func handleUpdate() error {
	// Check for --check flag
	checkOnly := false
	if len(os.Args) >= 3 && (os.Args[2] == "--check" || os.Args[2] == "-c") {
		checkOnly = true
	}

	if checkOnly {
		info, err := update.CheckForUpdate(Version)
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}

		if info.UpdateAvailable {
			fmt.Printf("Update available: %s (current: %s)\n", info.LatestVersion, info.CurrentVersion)
			fmt.Printf("Run 'scope update' to install\n")
			fmt.Printf("Release: %s\n", info.ReleaseURL)
		} else {
			fmt.Printf("Already up to date (version %s)\n", Version)
		}
		return nil
	}

	return update.PerformUpdate(Version)
}
