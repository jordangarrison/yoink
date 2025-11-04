package plugins

import (
	"context"
	"fmt"
)

// Plugin defines the interface that all credential revocation plugins must implement.
type Plugin interface {
	// Name returns the human-readable name of the plugin (e.g., "GitHub", "GitLab")
	Name() string

	// Detect checks if the given credential matches the pattern for this plugin.
	// Returns true if this plugin can handle the credential.
	Detect(credential string) bool

	// Revoke revokes the given credential using the provider's API.
	// Returns an error if revocation fails.
	Revoke(ctx context.Context, credential string) error

	// Validate checks if the credential format is valid without making API calls.
	// Returns an error if the credential format is invalid.
	Validate(credential string) error
}

// Registry manages all registered plugins.
type Registry struct {
	plugins []Plugin
}

// NewRegistry creates a new plugin registry.
func NewRegistry() *Registry {
	return &Registry{
		plugins: make([]Plugin, 0),
	}
}

// Register adds a plugin to the registry.
func (r *Registry) Register(p Plugin) {
	r.plugins = append(r.plugins, p)
}

// FindPlugin finds the appropriate plugin for the given credential.
// Returns the plugin and true if found, nil and false otherwise.
func (r *Registry) FindPlugin(credential string) (Plugin, bool) {
	for _, p := range r.plugins {
		if p.Detect(credential) {
			return p, true
		}
	}
	return nil, false
}

// ListPlugins returns all registered plugins.
func (r *Registry) ListPlugins() []Plugin {
	return r.plugins
}

// ErrNoPluginFound is returned when no plugin can handle a credential.
var ErrNoPluginFound = fmt.Errorf("no plugin found for credential")

// ErrInvalidCredential is returned when a credential format is invalid.
var ErrInvalidCredential = fmt.Errorf("invalid credential format")
