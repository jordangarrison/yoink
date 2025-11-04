# Contributing to Yoink

Thank you for your interest in contributing to Yoink! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Make
- [Devbox](https://www.jetify.com/devbox) (optional but recommended)

### Getting Started

1. Fork and clone the repository:
```bash
git clone https://github.com/jordangarrison/yoink.git
cd yoink
```

2. Set up your development environment:

**With Devbox (Recommended):**
```bash
devbox shell
```

**Without Devbox:**
```bash
go mod download
```

3. Run tests to verify setup:
```bash
make test
```

## Development Workflow

### Running Tests

```bash
# Unit tests
make test

# Integration tests
make integration-test

# All tests
make all
```

### Building

```bash
# Build the binary
make build

# Install locally
make install

# Clean build artifacts
make clean
```

### Code Style

- Follow standard Go conventions
- Run `gofmt` before committing
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions small and focused

### Testing Requirements

All contributions must include tests:

- **Unit tests**: For individual functions and methods
- **Integration tests**: For end-to-end functionality
- **Coverage**: Aim for >80% code coverage

Example test structure:
```go
func TestFeatureName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        // Test cases here
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Adding a New Provider Plugin

To add support for a new credential provider:

1. Create a new directory under `pkg/plugins/`:
```bash
mkdir -p pkg/plugins/yourprovider
```

2. Implement the `Plugin` interface:
```go
package yourprovider

import (
    "context"
    "github.com/jordangarrison/yoink/pkg/plugins"
)

type Plugin struct {
    // Plugin fields
}

func New() *Plugin {
    return &Plugin{}
}

func (p *Plugin) Name() string {
    return "YourProvider"
}

func (p *Plugin) Detect(credential string) bool {
    // Implement credential detection
}

func (p *Plugin) Revoke(ctx context.Context, credential string) error {
    // Implement revocation logic
}

func (p *Plugin) Validate(credential string) error {
    // Implement validation logic
}
```

3. Add comprehensive tests:
```go
// pkg/plugins/yourprovider/yourprovider_test.go
```

4. Register the plugin in command files:
```go
registry.Register(yourprovider.New())
```

5. Update documentation

## Commit Message Format

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test additions or modifications
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks

Examples:
```
feat(github): add support for fine-grained tokens
fix(core): handle empty credential list correctly
docs: update README with webhook examples
test(plugins): add integration tests for GitLab plugin
```

## Pull Request Process

1. Create a feature branch:
```bash
git checkout -b feature/your-feature-name
```

2. Make your changes and commit:
```bash
git add .
git commit -m "feat: add your feature"
```

3. Push to your fork:
```bash
git push origin feature/your-feature-name
```

4. Create a Pull Request with:
   - Clear description of changes
   - Link to related issues
   - Test coverage report
   - Screenshots (if applicable)

5. Ensure CI passes:
   - All tests pass
   - No linting errors
   - Build succeeds

## Code Review Process

- All PRs require at least one approval
- Address reviewer comments promptly
- Keep PRs focused and reasonably sized
- Squash commits before merging (if requested)

## Documentation

- Update README.md for user-facing changes
- Add inline code comments for complex logic
- Update CONTRIBUTING.md for development process changes
- Include usage examples

## Getting Help

- Check existing issues and PRs
- Open an issue for bugs or feature requests
- Join discussions on GitHub

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
