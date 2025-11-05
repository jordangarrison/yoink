package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/jordangarrison/yoink/pkg/plugins"
)

const (
	// GitHubRevokeEndpoint is the GitHub API endpoint for credential revocation
	GitHubRevokeEndpoint = "https://api.github.com/credentials/revoke"

	// Token prefixes
	prefixPersonalAccessToken = "ghp_"
	prefixOAuthToken          = "gho_"
	prefixServerToken         = "ghs_"
	prefixUserToken           = "ghu_"
	prefixRefreshToken        = "ghr_"
)

var (
	// tokenPattern matches GitHub token formats
	// GitHub tokens start with gh{type}_ followed by alphanumeric characters
	tokenPattern = regexp.MustCompile(`^gh[prsouh]_[a-zA-Z0-9]{36,}$`)
)

// Plugin implements the plugins.Plugin interface for GitHub credentials.
type Plugin struct {
	httpClient *http.Client
}

// New creates a new GitHub plugin instance.
func New() *Plugin {
	return &Plugin{
		httpClient: &http.Client{},
	}
}

// NewWithClient creates a new GitHub plugin with a custom HTTP client.
func NewWithClient(client *http.Client) *Plugin {
	return &Plugin{
		httpClient: client,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "GitHub"
}

// Detect checks if the credential is a GitHub token.
func (p *Plugin) Detect(credential string) bool {
	credential = strings.TrimSpace(credential)
	return tokenPattern.MatchString(credential)
}

// Validate checks if the credential format is valid.
func (p *Plugin) Validate(credential string) error {
	credential = strings.TrimSpace(credential)

	if !p.Detect(credential) {
		return fmt.Errorf("%w: not a valid GitHub token format", plugins.ErrInvalidCredential)
	}

	// Check minimum length
	if len(credential) < 40 {
		return fmt.Errorf("%w: GitHub token too short", plugins.ErrInvalidCredential)
	}

	return nil
}

// Revoke revokes the GitHub credential using the GitHub API.
func (p *Plugin) Revoke(ctx context.Context, credential string) error {
	credential = strings.TrimSpace(credential)

	// Validate before attempting revocation
	if err := p.Validate(credential); err != nil {
		return err
	}

	// Prepare request body
	reqBody := map[string][]string{
		"credentials": {credential},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, GitHubRevokeEndpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body for error messages
	respBody, _ := io.ReadAll(resp.Body)

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// GetTokenType returns the human-readable type of the GitHub token.
func GetTokenType(credential string) string {
	credential = strings.TrimSpace(credential)

	switch {
	case strings.HasPrefix(credential, prefixPersonalAccessToken):
		return "Personal Access Token"
	case strings.HasPrefix(credential, prefixOAuthToken):
		return "OAuth Token"
	case strings.HasPrefix(credential, prefixServerToken):
		return "Server Token"
	case strings.HasPrefix(credential, prefixUserToken):
		return "User Access Token"
	case strings.HasPrefix(credential, prefixRefreshToken):
		return "Refresh Token"
	default:
		return "Unknown Token Type"
	}
}
