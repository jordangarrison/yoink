# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Yoink is a credential revocation tool that automatically detects and revokes exposed credentials using a plugin-based architecture. It operates as both a CLI tool and a webhook server, supporting multiple input methods (direct args, stdin, file watching, HTTP endpoints).

## Build Commands

```bash
# Primary build workflow
make build              # Build binary to bin/yoink
make test              # Run unit tests
make integration-test  # Run integration tests (bash script)
make lint              # Run golangci-lint (if installed)
make all               # Run test + lint + build
make clean             # Remove bin/ directory

# Development workflow
devbox shell           # Enter dev environment (optional)
go mod download        # Download dependencies

# Running the tool
./bin/yoink --help
./bin/yoink revoke --dry-run <credential>
./bin/yoink watch --recursive <path>
./bin/yoink serve --port 8080
```

## Testing Commands

```bash
# Run all tests with coverage
go test ./... -v -cover

# Run single package tests
go test ./pkg/plugins/github -v

# Run specific test
go test ./pkg/plugins/github -run TestPlugin_Detect -v

# Integration tests (requires built binary)
bash test/integration_test.sh
```

## Architecture

### Core Data Flow

```
Input (CLI/stdin/file/webhook)
  → internal/commands (cobra commands)
  → internal/core.Engine (orchestration)
  → pkg/plugins.Registry (plugin selection)
  → pkg/plugins/github.Plugin (provider-specific logic)
  → GitHub API
  → internal/output.Formatter (result display)
```

### Plugin System Architecture

The plugin system is the heart of Yoink. All credential providers must implement the `plugins.Plugin` interface:

```go
type Plugin interface {
    Name() string                                      // Human-readable name
    Detect(credential string) bool                     // Pattern matching
    Revoke(ctx context.Context, credential string) error  // API call
    Validate(credential string) error                 // Format validation
}
```

**Key concept**: The `Registry` iterates through registered plugins calling `Detect()` until one matches. Only one plugin handles each credential.

### Engine Concurrency Model

`internal/core.Engine` provides three revocation methods:
- `Revoke()`: Single credential, synchronous
- `RevokeBatch()`: Multiple credentials, concurrent with goroutines + mutex for results
- `RevokeWithCallback()`: Multiple credentials, concurrent with streaming callback

**Important**: Batch operations use goroutines for concurrent revocations. Results maintain input order in `RevokeBatch()` but callback order is non-deterministic in `RevokeWithCallback()`.

### Command Structure

All CLI commands live in `internal/commands/`:
- `root.go`: Global flags (--dry-run, --verbose), cobra setup
- `revoke.go`: Direct revocation, handles stdin detection
- `watch.go`: File watcher using fsnotify, scans content with regex
- `serve.go`: HTTP server with POST /revoke and GET /health endpoints

**Critical**: Each command creates its own plugin registry and registers available plugins. When adding new plugins, register them in all command files.

### Dry-Run Mode

Dry-run is a global flag that changes engine behavior:
- When enabled: `Engine.Revoke()` only calls `plugin.Validate()`, never `plugin.Revoke()`
- Results still return `RevocationResult` but with `DryRun: true`
- No actual API calls are made

## Adding New Provider Plugins

1. Create `pkg/plugins/<provider>/` directory
2. Implement `Plugin` interface with proper error handling
3. Add comprehensive tests (detect, validate, revoke scenarios)
4. Register in all command files:
   ```go
   registry.Register(yourprovider.New())
   ```
5. Update README.md with supported token types

**Testing pattern**: Use table-driven tests with mock HTTP servers for API testing (see `github_test.go` for reference).

## GitHub Actions & Release

- CI workflow: `.github/workflows/ci.yml` runs on all pushes/PRs
- Release workflow: `.github/workflows/release.yml` uses GoReleaser
- GoReleaser config: `.goreleaser.yml` builds for multiple platforms
- Create release: `git tag v1.x.x && git push origin v1.x.x`

## Token Detection Patterns

GitHub tokens use prefix-based detection:
- `ghp_` = Personal Access Token
- `gho_` = OAuth Token
- `ghs_` = Server Token
- `ghu_` = User Access Token
- `ghr_` = Refresh Token

Pattern: `^gh[prsouh]_[a-zA-Z0-9]{36,}$` (see `pkg/plugins/github/github.go`)

## Security Practices

- Credentials are masked in output (first 8 + last 4 chars shown)
- GitHub API endpoint is unauthenticated: `https://api.github.com/credentials/revoke`
- No credentials are logged in full anywhere
- File watcher scans content but doesn't store matches

## Common Patterns

### Creating a new command
1. Create file in `internal/commands/<name>.go`
2. Implement cobra.Command with RunE function
3. Initialize registry + engine + formatter in RunE
4. Add command to rootCmd in init(): `rootCmd.AddCommand(yourCmd)`

### Testing concurrent behavior
Use `sync.WaitGroup` and mutex pattern from `engine.go` and test with goroutines + channels.

### Output formatting
`internal/output.Formatter` interface supports multiple formats. Currently only `HumanFormatter` implemented; JSON formatter is placeholder for future webhook server enhancement.

## Development Notes

- Go 1.21+ required for testing features
- Uses conventional commits (feat/fix/docs/test/refactor)
- Test coverage target: >80% (currently 87.5%-100% across packages)
- Devbox provides reproducible environment but is optional
