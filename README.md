# Scope

A fast CLI tool to organize and navigate your filesystem by tagging folders. Stop hunting for deeply nested project directories—tag them once, access them instantly.

## What is Scope?

Scope lets you tag folders with custom labels and create temporary workspaces containing only the folders you need. It's like bookmarks for your filesystem, but better.

## Features

- **Tag any folder** - Mark folders with meaningful labels like `work`, `personal`, `project-x`
- **Multi-tag support** - One folder can have multiple tags
- **Instant workspaces** - Create temporary environments showing only tagged folders
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

### `scope tag <path> <tag>`

Tag a folder. Use `.` for current directory. Multiple tags can be comma-separated.

```bash
scope tag . work
scope tag ~/my-project work,urgent,backend
```

### `scope untag <path> <tag>`

Remove a tag from a folder. Use `.` for current directory.

```bash
scope untag . work
scope untag ~/my-project urgent
```

### `scope list [tag]`

List all tags and their folder counts, or list all folders with a specific tag.

```bash
scope list          # Show all tags
scope list work     # Show all folders tagged 'work'
```

### `scope start <tag>`

Create a temporary workspace with symlinks to all folders matching the tag.

```bash
scope start work
# Opens new shell in /tmp/scope-work-<random>/
# All your 'work' folders are now accessible via ls
```

### `scope remove-tag <tag>`

Delete a tag entirely (removes it from all folders).

```bash
scope remove-tag old-project
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
