# Contributing to Scope

Thank you for your interest in contributing to Scope! This document provides guidelines and instructions for contributing.

## Code of Conduct

Be respectful, inclusive, and constructive in all interactions.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/Scope.git
   cd Scope
   ```
3. **Set up the development environment**:
   ```bash
   make deps
   make dev-setup  # Installs development tools
   ```

## Development Workflow

### Making Changes

1. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** and write tests

3. **Run tests locally**:
   ```bash
   make test
   make test-integration
   ```

4. **Check code quality**:
   ```bash
   make fmt    # Format code
   make vet    # Run go vet
   make lint   # Run linters
   ```

5. **Run all CI checks**:
   ```bash
   make ci
   ```

### Commit Guidelines

- Write clear, concise commit messages
- Use conventional commits format:
  ```
  feat: add export command
  fix: handle symlink errors correctly
  docs: update installation instructions
  test: add tests for tag package
  refactor: simplify database initialization
  ```

- Keep commits atomic (one logical change per commit)
- Reference issues in commits when applicable: `fixes #123`

### Pull Request Process

1. **Update documentation** if you're changing user-facing behavior

2. **Ensure all tests pass**:
   ```bash
   make test
   make test-integration
   make ci
   ```

3. **Push your branch**:
   ```bash
   git push origin feature/your-feature-name
   ```

4. **Create a Pull Request** on GitHub

5. **Address review feedback** if any

## Testing Guidelines

### Writing Tests

- Write tests for all new features
- Maintain or improve code coverage
- Use table-driven tests when appropriate
- Test edge cases and error conditions

### Test Structure

```go
func TestFeatureName(t *testing.T) {
    // Setup
    testData := setupTestData(t)
    defer cleanup()

    // Execute
    result := functionUnderTest(testData)

    // Assert
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
}
```

### Running Specific Tests

```bash
# Run specific package tests
go test -v ./internal/tag

# Run specific test
go test -v -run TestAddTag ./internal/tag

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Code Style

### Go Guidelines

- Follow the [Effective Go](https://golang.org/doc/effective_go) guidelines
- Use `gofmt` to format code (automatic with `make fmt`)
- Keep functions small and focused
- Write self-documenting code
- Add comments for exported functions and complex logic

### Project Structure

```
Scope/
├── cmd/scope/          # CLI entry point
├── internal/           # Internal packages
│   ├── db/            # Database operations
│   ├── tag/           # Tag management
│   └── session/       # Session management
├── scripts/           # Build and test scripts
└── .github/           # GitHub Actions workflows
```

## Adding New Features

### Planning

1. **Open an issue** to discuss the feature first
2. **Get feedback** from maintainers
3. **Design the implementation** before coding

### Implementation Checklist

- [ ] Feature implemented in appropriate package
- [ ] Unit tests written and passing
- [ ] Integration tests updated if needed
- [ ] Documentation updated (README, godoc comments)
- [ ] Error handling implemented
- [ ] Edge cases considered
- [ ] Code follows project style
- [ ] All tests pass
- [ ] No linter warnings

## Bug Reports

When reporting bugs, include:

1. **Description** of the issue
2. **Steps to reproduce**
3. **Expected behavior**
4. **Actual behavior**
5. **Environment** (OS, Go version, Scope version)
6. **Logs or error messages** if applicable

## Feature Requests

When requesting features:

1. **Describe the problem** you're trying to solve
2. **Explain the proposed solution**
3. **Provide examples** of how it would be used
4. **Discuss alternatives** you've considered

## Release Process

Releases are automated via GitHub Actions when tags are pushed:

```bash
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
```

## Questions?

- Open an issue for questions
- Check existing issues and PRs for similar discussions
- Read the [README](README.md) and [claude.md](claude.md) for project details

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
