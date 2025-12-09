# Contributing to yfinance-go

Thank you for your interest in contributing to yfinance-go! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Code Style](#code-style)
- [Testing](#testing)
- [Documentation](#documentation)
- [Submitting Changes](#submitting-changes)
- [Review Process](#review-process)

## Code of Conduct

This project adheres to a Code of Conduct that all contributors are expected to follow. Please be respectful and constructive in all interactions.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/your-username/yfinance-go.git
   cd yfinance-go
   ```
3. **Add the upstream remote**:
   ```bash
   git remote add upstream https://github.com/AmpyFin/yfinance-go.git
   ```

## Development Setup

### Prerequisites

- Go 1.23 or later
- Git
- Make (optional, for using Makefile commands)

### Setup Steps

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Run tests** to verify your setup:
   ```bash
   go test ./...
   ```

3. **Build the CLI**:
   ```bash
   go build -o yfin ./cmd/yfin
   ```

### Development Tools

- **Linting**: We use `golangci-lint`. Run with:
  ```bash
  golangci-lint run
  ```

- **Formatting**: Use `gofmt` or `goimports`:
  ```bash
  gofmt -w .
  ```

- **Testing**: Run tests with coverage:
  ```bash
  go test -v -cover ./...
  ```

## Making Changes

### Branch Strategy

1. **Create a feature branch** from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Keep your branch up to date**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

### Commit Messages

Follow these guidelines for commit messages:

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests after the first line
- Use a blank line between the subject and body

**Good examples:**
```
Add support for intraday data fetching

Implements FetchIntradayBars method with support for 1m, 5m, 15m, 30m, and 60m intervals.
Fixes #123
```

```
Fix hardcoded MIC values in scraping functions

Replaces hardcoded "XNAS" with dynamic MIC inference from exchange information.
This enables proper support for NYSE, international exchanges, and other markets.
```

## Code Style

### General Guidelines

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting
- Keep functions focused and small
- Use meaningful variable and function names
- Add comments for exported functions and types
- Handle errors explicitly (don't ignore them)

### Naming Conventions

- **Packages**: lowercase, single word
- **Exported functions/types**: PascalCase
- **Unexported functions/types**: camelCase
- **Constants**: PascalCase or UPPER_CASE
- **Variables**: camelCase

### Error Handling

Always handle errors explicitly:

```go
// ✅ Good
result, err := someFunction()
if err != nil {
    return nil, fmt.Errorf("failed to do something: %w", err)
}

// ❌ Bad
result, _ := someFunction()
```

### Context Usage

Always accept `context.Context` as the first parameter for functions that:
- Make network requests
- Perform I/O operations
- Can be cancelled or timed out

```go
// ✅ Good
func (c *Client) FetchData(ctx context.Context, symbol string) (*Data, error) {
    // ...
}

// ❌ Bad
func (c *Client) FetchData(symbol string) (*Data, error) {
    // ...
}
```

## Testing

### Test Requirements

- **All new code must include tests**
- **Maintain or improve test coverage**
- **Tests should be fast and isolated**
- **Use table-driven tests when appropriate**

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run specific package tests
go test ./internal/httpx/...
```

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "AAPL",
            want:    "XNAS",
            wantErr: false,
        },
        {
            name:    "invalid input",
            input:   "",
            want:    "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("FunctionName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

Integration tests should be tagged with `//go:build integration`:

```go
//go:build integration

package httpx

import (
    "testing"
)

func TestClientIntegration(t *testing.T) {
    // Integration test code
}
```

Run integration tests with:
```bash
go test -tags=integration ./...
```

## Documentation

### Code Documentation

- **All exported functions, types, and packages must have documentation**
- Use complete sentences
- Start with the name of the thing being documented

```go
// ✅ Good
// FetchDailyBars fetches daily OHLCV data for a symbol within the specified date range.
// It returns normalized bar data with proper scaling and currency information.
func (c *Client) FetchDailyBars(ctx context.Context, symbol string, start, end time.Time, adjusted bool, runID string) (*norm.NormalizedBarBatch, error) {
    // ...
}

// ❌ Bad
// fetches bars
func (c *Client) FetchDailyBars(...) {
    // ...
}
```

### README Updates

- Update README.md if you add new features
- Add examples for new functionality
- Update the API reference section

### Documentation Files

- Update relevant docs in `docs/` directory
- Add examples in `examples/` directory
- Update runbooks if operational procedures change

## Submitting Changes

### Pull Request Process

1. **Ensure your code is ready**:
   - All tests pass
   - Code is formatted (`gofmt`)
   - Linter passes (`golangci-lint`)
   - Documentation is updated

2. **Push your branch**:
   ```bash
   git push origin feature/your-feature-name
   ```

3. **Create a Pull Request** on GitHub:
   - Use a clear, descriptive title
   - Fill out the PR template
   - Reference related issues
   - Add screenshots/examples if applicable

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] No new warnings generated
- [ ] Tests pass locally
```

## Review Process

### What Reviewers Look For

- **Correctness**: Does the code work as intended?
- **Code Quality**: Is it clean, readable, and maintainable?
- **Testing**: Are there adequate tests?
- **Documentation**: Is it well-documented?
- **Performance**: Are there any performance concerns?
- **Security**: Are there any security issues?

### Responding to Feedback

- Be open to feedback and suggestions
- Address all comments before requesting re-review
- Ask questions if something is unclear
- Be patient - reviews take time

### After Approval

Once your PR is approved:
- A maintainer will merge it
- Your contribution will be included in the next release
- Thank you for contributing! 🎉

## Areas for Contribution

We welcome contributions in these areas:

- **Bug Fixes**: Fix issues reported in GitHub Issues
- **New Features**: Implement features from the roadmap
- **Documentation**: Improve existing docs or add new ones
- **Tests**: Increase test coverage
- **Performance**: Optimize existing code
- **Examples**: Add more usage examples

## Getting Help

- **GitHub Issues**: For bug reports and feature requests
- **GitHub Discussions**: For questions and discussions
- **Documentation**: Check `docs/` directory for guides

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to yfinance-go! 🙏

