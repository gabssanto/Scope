package scan

// ScopeConfig represents the structure of a .scope YAML file
type ScopeConfig struct {
	Tags []string `yaml:"tags"`
}

// DiscoveredScope represents a discovered .scope file and its parsed content
type DiscoveredScope struct {
	FolderPath string   // The directory containing the .scope file
	FilePath   string   // Full path to the .scope file
	Tags       []string // Parsed tags from the file
}

// ScanResult contains all discovered .scope files from a scan
type ScanResult struct {
	Scopes []DiscoveredScope
	Errors []ScanError
}

// ScanError represents a non-fatal error during scanning
type ScanError struct {
	FilePath string
	Err      error
}
