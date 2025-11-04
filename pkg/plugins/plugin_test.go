package plugins

import (
	"context"
	"testing"
)

// mockPlugin is a mock implementation of the Plugin interface for testing.
type mockPlugin struct {
	name       string
	detectFunc func(string) bool
	revokeFunc func(context.Context, string) error
}

func (m *mockPlugin) Name() string {
	return m.name
}

func (m *mockPlugin) Detect(credential string) bool {
	if m.detectFunc != nil {
		return m.detectFunc(credential)
	}
	return false
}

func (m *mockPlugin) Revoke(ctx context.Context, credential string) error {
	if m.revokeFunc != nil {
		return m.revokeFunc(ctx, credential)
	}
	return nil
}

func (m *mockPlugin) Validate(credential string) error {
	if m.Detect(credential) {
		return nil
	}
	return ErrInvalidCredential
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	plugin1 := &mockPlugin{name: "test1"}
	plugin2 := &mockPlugin{name: "test2"}

	registry.Register(plugin1)
	registry.Register(plugin2)

	plugins := registry.ListPlugins()
	if len(plugins) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(plugins))
	}
}

func TestRegistry_FindPlugin(t *testing.T) {
	registry := NewRegistry()

	plugin1 := &mockPlugin{
		name: "test1",
		detectFunc: func(s string) bool {
			return s == "test1_credential"
		},
	}
	plugin2 := &mockPlugin{
		name: "test2",
		detectFunc: func(s string) bool {
			return s == "test2_credential"
		},
	}

	registry.Register(plugin1)
	registry.Register(plugin2)

	tests := []struct {
		name       string
		credential string
		wantPlugin string
		wantFound  bool
	}{
		{
			name:       "find plugin1",
			credential: "test1_credential",
			wantPlugin: "test1",
			wantFound:  true,
		},
		{
			name:       "find plugin2",
			credential: "test2_credential",
			wantPlugin: "test2",
			wantFound:  true,
		},
		{
			name:       "no match",
			credential: "unknown_credential",
			wantPlugin: "",
			wantFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin, found := registry.FindPlugin(tt.credential)
			if found != tt.wantFound {
				t.Errorf("FindPlugin() found = %v, want %v", found, tt.wantFound)
			}
			if found && plugin.Name() != tt.wantPlugin {
				t.Errorf("FindPlugin() plugin = %v, want %v", plugin.Name(), tt.wantPlugin)
			}
		})
	}
}

func TestRegistry_ListPlugins(t *testing.T) {
	registry := NewRegistry()

	// Empty registry
	plugins := registry.ListPlugins()
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins in empty registry, got %d", len(plugins))
	}

	// Add plugins
	plugin1 := &mockPlugin{name: "test1"}
	plugin2 := &mockPlugin{name: "test2"}

	registry.Register(plugin1)
	registry.Register(plugin2)

	plugins = registry.ListPlugins()
	if len(plugins) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(plugins))
	}

	// Verify plugin names
	names := make(map[string]bool)
	for _, p := range plugins {
		names[p.Name()] = true
	}

	if !names["test1"] || !names["test2"] {
		t.Error("not all registered plugins were returned")
	}
}
