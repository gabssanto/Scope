package scan

import (
	"fmt"
	"os"
	"path/filepath"
)

const scopeFileName = ".scope"

// Scan walks the directory tree starting from rootPath and discovers all .scope files
func Scan(rootPath string) (*ScanResult, error) {
	result := &ScanResult{
		Scopes: make([]DiscoveredScope, 0),
		Errors: make([]ScanError, 0),
	}

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip directories we can't access
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden directories (except the root if it's hidden)
		if d.IsDir() && path != rootPath {
			if len(d.Name()) > 0 && d.Name()[0] == '.' {
				return filepath.SkipDir
			}
		}

		// Check for .scope file
		if d.Name() == scopeFileName && !d.IsDir() {
			config, parseErr := ParseScopeFile(path)
			if parseErr != nil {
				result.Errors = append(result.Errors, ScanError{
					FilePath: path,
					Err:      parseErr,
				})
				return nil
			}

			if len(config.Tags) > 0 {
				folderPath := filepath.Dir(path)
				result.Scopes = append(result.Scopes, DiscoveredScope{
					FolderPath: folderPath,
					FilePath:   path,
					Tags:       config.Tags,
				})
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	return result, nil
}
