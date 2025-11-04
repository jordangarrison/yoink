# Yoink

> Automatically revoke exposed credentials before they can be exploited

Yoink is a command-line tool and webhook server that automatically detects and revokes exposed credentials. It features a plugin-based architecture that makes it easy to support multiple credential providers.

## Features

- **Multiple Input Methods**
  - Direct CLI arguments
  - Stdin piping
  - File watching with pattern matching
  - Webhook server for automation

- **GitHub Support** (v1)
  - Personal Access Tokens (PATs)
  - OAuth Tokens
  - Server Tokens
  - User Access Tokens
  - Refresh Tokens

- **Dry-Run Mode**
  - Test credential detection without actually revoking

- **Plugin Architecture**
  - Easy to extend with new credential providers

## Installation

### Using Go

```bash
go install github.com/jordangarrison/yoink/cmd/yoink@latest
```

### Using Devbox (Recommended for Development)

```bash
devbox shell
make install
```

### From Source

```bash
git clone https://github.com/jordangarrison/yoink.git
cd yoink
make build
./bin/yoink --version
```

## Usage

### Revoke Command

Revoke credentials directly from the command line or stdin.

```bash
# Revoke a single credential
yoink revoke ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# Revoke multiple credentials
yoink revoke ghp_xxx gho_yyy ghs_zzz

# Revoke from stdin
echo "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" | yoink revoke

# Revoke from a file
cat leaked_tokens.txt | yoink revoke

# Dry-run mode (validate without revoking)
yoink revoke --dry-run ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# Verbose output
yoink revoke -v ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

### Watch Command

Monitor files or directories for exposed credentials and automatically revoke them.

```bash
# Watch current directory
yoink watch .

# Watch specific directory recursively
yoink watch --recursive /path/to/repo

# Watch with file pattern matching
yoink watch --pattern "*.log" /var/log

# Watch in dry-run mode
yoink watch --dry-run --recursive /path/to/repo
```

### Serve Command

Start a webhook server to receive credential revocation requests.

```bash
# Start server on default port (8080)
yoink serve

# Start server on custom port
yoink serve --port 3000

# Start server with custom host
yoink serve --host 0.0.0.0 --port 8080

# Start server in dry-run mode
yoink serve --dry-run
```

#### API Endpoints

**POST /revoke**

Revoke one or more credentials.

Request:
```json
{
  "credentials": [
    "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "gho_yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy"
  ]
}
```

Response:
```json
{
  "total": 2,
  "successful": 2,
  "failed": 0,
  "dry_run": false,
  "results": [
    {
      "Credential": "ghp_xxx...xxx",
      "Plugin": "GitHub",
      "Success": true,
      "Error": null,
      "DryRun": false
    }
  ]
}
```

**GET /health**

Health check endpoint.

Response:
```json
{
  "status": "ok",
  "service": "yoink"
}
```

## GitHub Actions Integration

Use Yoink in your GitHub Actions workflows:

```yaml
name: Revoke Exposed Credentials

on:
  push:
    branches: [ main ]

jobs:
  revoke:
    runs-on: ubuntu-latest
    steps:
      - name: Install Yoink
        run: go install github.com/jordangarrison/yoink/cmd/yoink@latest

      - name: Revoke exposed token (dry-run)
        run: |
          echo "${{ secrets.EXPOSED_TOKEN }}" | yoink revoke --dry-run
```

### Using as a GitHub Action

You can also run Yoink directly in GitHub Actions:

```yaml
- name: Scan and revoke credentials
  run: |
    # Install yoink
    go install github.com/jordangarrison/yoink/cmd/yoink@latest

    # Scan repository for exposed credentials
    yoink watch --dry-run --recursive .
```

## Development

### Prerequisites

- Go 1.21+
- Make
- [Devbox](https://www.jetify.com/devbox) (optional but recommended)

### Setup with Devbox

```bash
# Enter development environment
devbox shell

# Run tests
make test

# Run integration tests
make integration-test

# Run linter
make lint

# Build
make build

# Run all checks
make all
```

### Project Structure

```
yoink/
├── cmd/
│   └── yoink/              # Main CLI entry point
├── internal/
│   ├── commands/           # CLI commands (revoke, watch, serve)
│   ├── core/               # Core revocation engine
│   ├── input/              # Input handling
│   └── output/             # Output formatting
├── pkg/
│   └── plugins/            # Plugin system
│       ├── plugin.go       # Plugin interface
│       └── github/         # GitHub plugin
├── test/
│   └── integration_test.sh # Integration tests
├── devbox.json             # Devbox configuration
├── Makefile                # Build automation
└── README.md
```

### Running Tests

```bash
# Unit tests
make test

# Integration tests
make integration-test

# All tests
make all
```

### Adding a New Plugin

1. Create a new directory under `pkg/plugins/`
2. Implement the `Plugin` interface:

```go
type Plugin interface {
    Name() string
    Detect(credential string) bool
    Revoke(ctx context.Context, credential string) error
    Validate(credential string) error
}
```

3. Register the plugin in your command implementations
4. Add tests for the new plugin

## How It Works

### GitHub Token Revocation

Yoink uses GitHub's [Credential Revocation API](https://docs.github.com/en/rest/credentials/revoke) to revoke exposed tokens. This API:

- Does not require authentication
- Works for any valid GitHub token
- Automatically notifies the token owner via email
- Logs the revocation in the owner's audit log

### Token Detection

Yoink detects GitHub tokens by their prefix patterns:

- `ghp_` - Personal Access Tokens
- `gho_` - OAuth Tokens
- `ghs_` - Server Tokens
- `ghu_` - User Access Tokens
- `ghr_` - Refresh Tokens

## Security Considerations

- Credentials are masked in output (showing only first 8 and last 4 characters)
- Dry-run mode allows safe testing
- No credentials are stored or logged in full
- All network communication uses HTTPS
- Webhook server supports rate limiting (future enhancement)

## Roadmap

### v2 Features (Planned)

- [ ] Additional provider plugins (GitLab, AWS, Azure)
- [ ] JSON output format
- [ ] Configuration file support
- [ ] Metrics/telemetry (Prometheus)
- [ ] Rate limiting for webhook server
- [ ] Database persistence for audit logs
- [ ] Slack/Discord notifications
- [ ] Advanced pattern detection using ML

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'feat: add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

Please use conventional commits and ensure all tests pass.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- GitHub for providing the Credential Revocation API
- The Go community for excellent tooling
- [Cobra](https://github.com/spf13/cobra) for CLI framework
- [fsnotify](https://github.com/fsnotify/fsnotify) for file watching

## Support

If you encounter any issues or have questions:

- Open an issue on [GitHub](https://github.com/jordangarrison/yoink/issues)
- Check existing issues for solutions
- Review the documentation

---

Made with ❤️ by Jordan Garrison
