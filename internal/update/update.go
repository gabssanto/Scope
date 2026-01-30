package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	repoOwner       = "gabssanto"
	repoName        = "Scope"
	checkInterval   = 24 * time.Hour
	githubAPIURL    = "https://api.github.com/repos/%s/%s/releases/latest"
	releaseAssetURL = "https://github.com/%s/%s/releases/download/%s/scope-%s-%s"
)

// Release represents a GitHub release
type Release struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
}

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	ReleaseURL      string
	ReleaseNotes    string
}

// getConfigDir returns the scope config directory
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "scope"), nil
}

// getCacheFile returns the path to the update cache file
func getCacheFile() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, ".update-check"), nil
}

// shouldCheck determines if we should check for updates based on cache
func shouldCheck() bool {
	cacheFile, err := getCacheFile()
	if err != nil {
		return true
	}

	info, err := os.Stat(cacheFile)
	if err != nil {
		return true
	}

	// Check if cache is older than check interval
	return time.Since(info.ModTime()) > checkInterval
}

// fetchLatestRelease fetches the latest release from GitHub
func fetchLatestRelease() (*Release, error) {
	url := fmt.Sprintf(githubAPIURL, repoOwner, repoName)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &release, nil
}

// saveCache saves the latest version to cache
func saveCache(version string) error {
	cacheFile, err := getCacheFile()
	if err != nil {
		return err
	}
	return os.WriteFile(cacheFile, []byte(version), 0644)
}

// readCache reads the cached version info
func readCache() (version string, hasUpdate bool) {
	cacheFile, err := getCacheFile()
	if err != nil {
		return "", false
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return "", false
	}

	parts := strings.Split(string(data), "\n")
	if len(parts) >= 1 {
		version = strings.TrimSpace(parts[0])
	}
	if len(parts) >= 2 {
		hasUpdate = parts[1] == "update"
	}

	return version, hasUpdate
}

// compareVersions compares two version strings (simple comparison)
// Returns true if latest > current
func compareVersions(current, latest string) bool {
	// Strip 'v' prefix
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// Simple string comparison works for semver in most cases
	// For more robust comparison, use a proper semver library
	return latest > current
}

// CheckForUpdate checks if a new version is available
func CheckForUpdate(currentVersion string) (*UpdateInfo, error) {
	release, err := fetchLatestRelease()
	if err != nil {
		return nil, err
	}

	info := &UpdateInfo{
		CurrentVersion:  currentVersion,
		LatestVersion:   release.TagName,
		UpdateAvailable: compareVersions(currentVersion, release.TagName),
		ReleaseURL:      release.HTMLURL,
		ReleaseNotes:    release.Body,
	}

	// Save to cache
	cacheContent := release.TagName
	if info.UpdateAvailable {
		cacheContent += "\nupdate"
	}
	_ = saveCache(cacheContent)

	return info, nil
}

// CheckForUpdateAsync checks for updates in the background
// Returns a channel that will receive the result
func CheckForUpdateAsync(currentVersion string) <-chan *UpdateInfo {
	ch := make(chan *UpdateInfo, 1)

	go func() {
		defer close(ch)

		// Skip if we checked recently
		if !shouldCheck() {
			// Check cache for pending update notification
			version, hasUpdate := readCache()
			if hasUpdate && compareVersions(currentVersion, version) {
				ch <- &UpdateInfo{
					CurrentVersion:  currentVersion,
					LatestVersion:   version,
					UpdateAvailable: true,
				}
			}
			return
		}

		info, err := CheckForUpdate(currentVersion)
		if err != nil {
			return
		}

		if info.UpdateAvailable {
			ch <- info
		}
	}()

	return ch
}

// GetUpdateNotice returns a formatted update notice if available
func GetUpdateNotice(currentVersion string) string {
	version, hasUpdate := readCache()
	if !hasUpdate || !compareVersions(currentVersion, version) {
		return ""
	}
	return fmt.Sprintf("\n\033[33m%s\033[0m scope %s available (current: %s) - run \033[1mscope update\033[0m\n",
		"!", version, currentVersion)
}

// PerformUpdate downloads and installs the latest version
func PerformUpdate(currentVersion string) error {
	fmt.Println("Checking for updates...")

	info, err := CheckForUpdate(currentVersion)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !info.UpdateAvailable {
		fmt.Printf("Already up to date (version %s)\n", currentVersion)
		return nil
	}

	fmt.Printf("New version available: %s (current: %s)\n", info.LatestVersion, info.CurrentVersion)
	fmt.Printf("Release notes: %s\n\n", info.ReleaseURL)

	// Determine platform
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Build download URL
	assetName := fmt.Sprintf("scope-%s-%s", goos, goarch)
	if goos == "windows" {
		assetName += ".exe"
	}

	downloadURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s",
		repoOwner, repoName, info.LatestVersion, assetName)

	fmt.Printf("Downloading %s...\n", assetName)

	// Download the binary
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d (asset may not exist for your platform)", resp.StatusCode)
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Create temp file for download
	tmpFile, err := os.CreateTemp("", "scope-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	// Download to temp file
	_, err = io.Copy(tmpFile, resp.Body)
	_ = tmpFile.Close()
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}

	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Backup current binary
	backupPath := execPath + ".backup"
	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Move new binary into place
	if err := os.Rename(tmpPath, execPath); err != nil {
		// Try to restore backup
		_ = os.Rename(backupPath, execPath)
		return fmt.Errorf("failed to install update: %w", err)
	}

	// Remove backup
	_ = os.Remove(backupPath)

	// Clear update cache
	cacheFile, _ := getCacheFile()
	_ = os.Remove(cacheFile)

	fmt.Printf("\nSuccessfully updated to %s!\n", info.LatestVersion)
	return nil
}
