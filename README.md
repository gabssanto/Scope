# Scope

A fast CLI tool to organize and navigate your filesystem by tagging folders. Stop hunting for deeply nested project directories—tag them once, access them instantly.

## What is Scope?

Scope lets you tag folders with custom labels and create temporary workspaces containing only the folders you need. It's like bookmarks for your filesystem, but better.

## Features

- **Tag any folder** - Mark folders with meaningful labels like `work`, `personal`, `project-x`
- **Multi-tag support** - One folder can have multiple tags
- **Instant workspaces** - Create temporary environments showing only tagged folders
- **Bulk operations** - Run commands across all tagged folders at once
- **Project scanning** - Auto-tag projects using `.scope` files
- **Self-updating** - Built-in update system with version checks
- **Fast & lightweight** - Single binary, no dependencies, SQLite backend
- **Cross-platform** - Works on Linux, macOS, and Windows

## Quick Start

```bash
# Tag your current directory
scope tag . work

# Tag specific paths
scope tag ~/projects/my-app work,backend
scope tag ~/Documents/taxes personal

# List all tags
scope list

# List folders with a specific tag
scope list work

# Start a scoped session
scope start work
# You're now in a temp directory with symlinks to all 'work' folders
# Just use ls, cd, and navigate normally!

# Remove a tag from current directory
scope untag . work

# Remove a tag from specific path
scope untag ~/projects/my-app backend

# Exit scoped session
exit
```

## Installation

### Linux/macOS

```bash
curl -sf https://raw.githubusercontent.com/gabssanto/Scope/main/install.sh | sh
```

### From Source

```bash
git clone https://github.com/gabssanto/Scope.git
cd Scope
go build -o scope cmd/scope/main.go
sudo mv scope /usr/local/bin/
```

### From Releases

Download the latest binary for your platform from [GitHub Releases](https://github.com/gabssanto/Scope/releases).

## Commands

### Tagging

#### `scope tag <path> <tag>`

Tag a folder. Use `.` for current directory. Multiple tags can be comma-separated.

```bash
scope tag . work
scope tag ~/my-project work,urgent,backend
```

#### `scope bulk <file> <tag> [--dry-run]`

Bulk tag multiple paths from a file. The file should contain one path per line.
Empty lines and lines starting with `#` are ignored.

```bash
scope bulk paths.txt work           # Tag all paths with 'work'
scope bulk paths.txt work --dry-run # Preview what would be tagged
```

Example paths file:
```
# Services
/home/user/project/services/api
/home/user/project/services/frontend

# Infrastructure
/home/user/project/infra/db
```

#### `scope untag <path> <tag>`

Remove a tag from a folder.

```bash
scope untag . work
scope untag ~/my-project urgent
```

#### `scope tags <path>`

Show all tags for a specific folder.

```bash
scope tags .
scope tags ~/my-project
```

#### `scope rename <old> <new>`

Rename a tag across all folders.

```bash
scope rename old-name new-name
```

#### `scope remove-tag <tag>`

Delete a tag entirely (removes it from all folders).

```bash
scope remove-tag old-project
```

### Listing & Navigation

#### `scope list [tag]`

List all tags and their folder counts, or list all folders with a specific tag.

```bash
scope list          # Show all tags
scope list work     # Show all folders tagged 'work'
```

#### `scope go <tag>`

Quick jump to a tagged folder. Outputs the path for shell integration.

```bash
scope go work       # Outputs path (single folder)
scope go work       # Shows picker (multiple folders)
```

**Shell integration** - Add to your `.bashrc` or `.zshrc`:
```bash
sg() { cd "$(scope go "$@")" 2>/dev/null || scope go "$@"; }
```

Then use `sg work` to instantly cd to your work folder.

#### `scope pick [tag]`

Interactive folder picker with search/filter support.

```bash
scope pick          # Pick from all tagged folders
scope pick work     # Pick from folders with 'work' tag
```

#### `scope open <tag>`

Open tagged folder(s) in your system file manager (Finder/Nautilus/Explorer).

```bash
scope open work
```

#### `scope edit <tag>`

Open tagged folder(s) in your editor (`$EDITOR`, `$VISUAL`, or auto-detected).

```bash
scope edit work
```

### Sessions

#### `scope start <tag>`

Create a temporary workspace with symlinks to all folders matching the tag.

```bash
scope start work
# Opens new shell in /tmp/scope-work-<random>/
# All your 'work' folders are now accessible via ls
# Type 'exit' to leave and auto-cleanup
```

### Bulk Operations

#### `scope each <tag> <command>`

Run a command in each tagged folder. Use `-p` for parallel execution.

```bash
scope each work "git status -s"      # Run sequentially
scope each work -p "npm install"     # Run in parallel
scope each backend "go test ./..."   # Run tests across all backend projects
```

#### `scope status <tag>`

Show git status for all tagged repositories (only shows repos with changes).

```bash
scope status work
```

#### `scope pull <tag>`

Git pull across all tagged repositories (runs in parallel).

```bash
scope pull work
```

### Project Scanning

#### `scope scan [path]`

Scan a directory for `.scope` files and interactively apply tags.

```bash
scope scan              # Scan current directory
scope scan ~/projects   # Scan specific directory
```

### Maintenance

#### `scope prune [--dry-run]`

Remove folders that no longer exist from the database.

```bash
scope prune --dry-run   # Preview what would be removed
scope prune             # Actually remove stale entries
```

#### `scope update [--check]`

Update scope to the latest version.

```bash
scope update --check    # Just check if update available
scope update            # Download and install latest version
```

#### `scope export`

Export all tags to YAML (outputs to stdout).

```bash
scope export > backup.yml
```

Output format:
```yaml
version: 1
tags:
  work:
    - /path/to/project1
    - /path/to/project2
  personal:
    - /path/to/blog
```

#### `scope import <file>`

Import tags from a YAML file.

```bash
scope import backup.yml
```

#### `scope completions <shell>`

Generate shell completion scripts.

```bash
# Bash - add to ~/.bashrc
eval "$(scope completions bash)"

# Zsh - add to ~/.zshrc
eval "$(scope completions zsh)"

# Fish - save to completions directory
scope completions fish > ~/.config/fish/completions/scope.fish
```

#### `scope debug`

Show debug information (version, database path, stats).

```bash
scope debug
```

## Project Configuration (`.scope` files)

You can add a `.scope` file to any project directory to define its tags. This makes it easy to share tagging conventions across teams or set up new machines.

### File Format

Create a `.scope` file in your project root:

```yaml
tags:
  - work
  - backend
  - api
```

### Scanning for Projects

Use `scope scan` to discover and apply tags from `.scope` files:

```bash
# Scan your projects directory
scope scan ~/projects

# Found 3 .scope files:
#   ~/projects/api [work, backend, api]
#   ~/projects/frontend [work, frontend, react]
#   ~/projects/scripts [work, tools]
#
# Select folders to tag (all selected by default)
# space: toggle, enter: confirm
```

The scanner will:
- Recursively find all `.scope` files
- Skip hidden directories (`.git`, `.node_modules`, etc.)
- Show an interactive picker to select which projects to tag
- Apply the tags from each `.scope` file

### Example Project Structure

```
~/projects/
├── my-api/
│   ├── .scope          # tags: [work, backend, go]
│   ├── main.go
│   └── ...
├── my-frontend/
│   ├── .scope          # tags: [work, frontend, react]
│   ├── package.json
│   └── ...
└── scripts/
    ├── .scope          # tags: [work, tools]
    └── ...
```

## How It Works

1. **Database**: Scope stores folder paths and tags in a local SQLite database at `~/.config/scope/scope.db`
2. **Symlinks**: When you run `scope start`, it creates a temp directory with symlinks to all matching folders
3. **New Shell**: You get a fresh shell session in that temp directory
4. **Cleanup**: Temp directories are automatically cleaned up on exit

## Why Scope?

- **Faster navigation**: No more `cd ../../../../../../projects/deeply/nested/folder`
- **Project organization**: Group related projects across different locations
- **Context switching**: Instantly switch between work, personal, or client projects
- **Simple**: Just tags and folders—no complex configuration

## Development

### Quick Start

```bash
# Clone the repository
git clone https://github.com/gabssanto/Scope.git
cd Scope

# Install dependencies
make deps

# Build
make build

# Run tests
make test

# Run integration tests
make test-integration
```

### Makefile Commands

The project includes a comprehensive Makefile with many useful commands:

```bash
make help           # Show all available commands
make build          # Build the binary
make test           # Run unit tests with race detector
make test-coverage  # Run tests with coverage report
make test-integration # Run integration tests
make clean          # Remove build artifacts
make install        # Install to /usr/local/bin
make ci             # Run all CI checks (fmt, vet, lint, test, build)
```

### Testing

Scope has comprehensive test coverage:

- **Unit tests**: Test individual packages (`internal/db`, `internal/tag`, `internal/session`)
- **Integration tests**: End-to-end testing of CLI commands
- **Benchmarks**: Performance testing for critical operations

```bash
# Run all unit tests
make test

# Run with coverage
make test-coverage

# Run integration tests
make test-integration

# Run benchmarks
make benchmark

# Quick test (short mode)
make qtest
```

### Building for Release

```bash
# Build for all platforms
make build-all

# Create release packages
make release

# Or use the build script
./scripts/build-release.sh
```

### Code Quality

```bash
# Format code
make fmt

# Run vet
make vet

# Run linters (requires golangci-lint)
make lint

# Run all checks
make ci
```

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test && make test-integration`)
5. Run code quality checks (`make ci`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Development Guidelines

- Write tests for new features
- Maintain test coverage above 80%
- Follow Go best practices and idioms
- Run `make ci` before submitting PRs
- Update documentation for user-facing changes

## License

MIT
