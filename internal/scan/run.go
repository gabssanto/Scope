package scan

import (
	"fmt"

	"github.com/gabssanto/Scope/internal/tag"
)

// RunScan orchestrates the entire scan operation
func RunScan(rootPath string) error {
	// Step 1: Scan for .scope files
	fmt.Printf("Scanning %s for .scope files...\n\n", rootPath)

	result, err := Scan(rootPath)
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if len(result.Scopes) == 0 {
		fmt.Println("No .scope files found.")
		return nil
	}

	// Step 2: Show summary
	ShowScanSummary(result)

	// Step 3: Interactive scope selection
	selectedScopes, err := SelectScopes(result.Scopes)
	if err != nil {
		return err
	}

	if len(selectedScopes) == 0 {
		fmt.Println("No folders selected. Nothing to apply.")
		return nil
	}

	// Step 4: Apply tags for selected scopes
	appliedCount := 0
	for _, scope := range selectedScopes {
		for _, t := range scope.Tags {
			if err := tag.AddTag(scope.FolderPath, t); err != nil {
				fmt.Printf("Warning: failed to add tag '%s' to %s: %v\n",
					t, scope.FolderPath, err)
				continue
			}
			appliedCount++
		}
	}

	fmt.Printf("\nApplied %d tag assignments.\n", appliedCount)
	return nil
}
