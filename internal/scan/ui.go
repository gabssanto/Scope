package scan

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

// ShowScanSummary displays what was found during the scan
func ShowScanSummary(result *ScanResult) {
	fmt.Printf("Found %d .scope files:\n\n", len(result.Scopes))

	for _, scope := range result.Scopes {
		fmt.Printf("  %s\n", scope.FolderPath)
		fmt.Printf("    Tags: %s\n", strings.Join(scope.Tags, ", "))
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\nWarnings (%d files had parsing errors):\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Printf("  %s: %v\n", e.FilePath, e.Err)
		}
	}

	fmt.Println()
}

// SelectScopes presents an interactive multi-select UI for selecting which scopes to apply
func SelectScopes(scopes []DiscoveredScope) ([]DiscoveredScope, error) {
	if len(scopes) == 0 {
		return nil, nil
	}

	// Build options for the multi-select
	options := make([]huh.Option[int], len(scopes))
	for i, scope := range scopes {
		label := fmt.Sprintf("%s [%s]", scope.FolderPath, strings.Join(scope.Tags, ", "))
		options[i] = huh.NewOption(label, i).Selected(true)
	}

	var selectedIndices []int

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[int]().
				Title("Select folders to tag (all selected by default)").
				Description("space: toggle, enter: confirm, /: filter").
				Options(options...).
				Value(&selectedIndices),
		),
	)

	err := form.Run()
	if err != nil {
		return nil, fmt.Errorf("selection cancelled: %w", err)
	}

	// Build result from selected indices
	selected := make([]DiscoveredScope, 0, len(selectedIndices))
	for _, idx := range selectedIndices {
		selected = append(selected, scopes[idx])
	}

	return selected, nil
}
