package scan

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseScopeFile reads and parses a .scope YAML file
func ParseScopeFile(filePath string) (*ScopeConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var config ScopeConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate and clean tags
	cleanedTags := make([]string, 0, len(config.Tags))
	for _, tag := range config.Tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			cleanedTags = append(cleanedTags, tag)
		}
	}
	config.Tags = cleanedTags

	return &config, nil
}
