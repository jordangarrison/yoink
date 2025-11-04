package core

import (
	"context"
	"errors"
	"testing"

	"github.com/jordangarrison/yoink/pkg/plugins"
)

// mockPlugin is a mock implementation for testing
type mockPlugin struct {
	name         string
	shouldDetect bool
	shouldError  bool
}

func (m *mockPlugin) Name() string {
	return m.name
}

func (m *mockPlugin) Detect(credential string) bool {
	return m.shouldDetect
}

func (m *mockPlugin) Revoke(ctx context.Context, credential string) error {
	if m.shouldError {
		return errors.New("mock revocation error")
	}
	return nil
}

func (m *mockPlugin) Validate(credential string) error {
	if m.shouldError {
		return errors.New("mock validation error")
	}
	return nil
}

func TestEngine_SetDryRun(t *testing.T) {
	registry := plugins.NewRegistry()
	engine := NewEngine(registry)

	if engine.IsDryRun() {
		t.Error("expected dry-run to be false by default")
	}

	engine.SetDryRun(true)
	if !engine.IsDryRun() {
		t.Error("expected dry-run to be true after setting")
	}

	engine.SetDryRun(false)
	if engine.IsDryRun() {
		t.Error("expected dry-run to be false after setting")
	}
}

func TestEngine_Revoke(t *testing.T) {
	tests := []struct {
		name           string
		credential     string
		plugin         *mockPlugin
		dryRun         bool
		wantSuccess    bool
		wantPluginName string
	}{
		{
			name:       "successful revocation",
			credential: "test_credential",
			plugin: &mockPlugin{
				name:         "TestPlugin",
				shouldDetect: true,
				shouldError:  false,
			},
			dryRun:         false,
			wantSuccess:    true,
			wantPluginName: "TestPlugin",
		},
		{
			name:       "failed revocation",
			credential: "test_credential",
			plugin: &mockPlugin{
				name:         "TestPlugin",
				shouldDetect: true,
				shouldError:  true,
			},
			dryRun:         false,
			wantSuccess:    false,
			wantPluginName: "TestPlugin",
		},
		{
			name:           "no plugin found",
			credential:     "test_credential",
			plugin:         &mockPlugin{shouldDetect: false},
			dryRun:         false,
			wantSuccess:    false,
			wantPluginName: "none",
		},
		{
			name:       "dry-run mode success",
			credential: "test_credential",
			plugin: &mockPlugin{
				name:         "TestPlugin",
				shouldDetect: true,
				shouldError:  false,
			},
			dryRun:         true,
			wantSuccess:    true,
			wantPluginName: "TestPlugin",
		},
		{
			name:       "dry-run mode validation failure",
			credential: "test_credential",
			plugin: &mockPlugin{
				name:         "TestPlugin",
				shouldDetect: true,
				shouldError:  true,
			},
			dryRun:         true,
			wantSuccess:    false,
			wantPluginName: "TestPlugin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := plugins.NewRegistry()
			registry.Register(tt.plugin)

			engine := NewEngine(registry)
			engine.SetDryRun(tt.dryRun)

			result := engine.Revoke(context.Background(), tt.credential)

			if result.Success != tt.wantSuccess {
				t.Errorf("Revoke() success = %v, want %v", result.Success, tt.wantSuccess)
			}

			if result.Plugin != tt.wantPluginName {
				t.Errorf("Revoke() plugin = %v, want %v", result.Plugin, tt.wantPluginName)
			}

			if result.DryRun != tt.dryRun {
				t.Errorf("Revoke() dryRun = %v, want %v", result.DryRun, tt.dryRun)
			}

			if result.Credential != tt.credential {
				t.Errorf("Revoke() credential = %v, want %v", result.Credential, tt.credential)
			}
		})
	}
}

func TestEngine_RevokeBatch(t *testing.T) {
	registry := plugins.NewRegistry()
	plugin := &mockPlugin{
		name:         "TestPlugin",
		shouldDetect: true,
		shouldError:  false,
	}
	registry.Register(plugin)

	engine := NewEngine(registry)

	credentials := []string{"cred1", "cred2", "cred3"}
	results := engine.RevokeBatch(context.Background(), credentials)

	if len(results) != len(credentials) {
		t.Errorf("RevokeBatch() returned %d results, want %d", len(results), len(credentials))
	}

	for i, result := range results {
		if result.Credential != credentials[i] {
			t.Errorf("RevokeBatch() result[%d] credential = %v, want %v", i, result.Credential, credentials[i])
		}
		if !result.Success {
			t.Errorf("RevokeBatch() result[%d] should have succeeded", i)
		}
	}
}

func TestEngine_RevokeWithCallback(t *testing.T) {
	registry := plugins.NewRegistry()
	plugin := &mockPlugin{
		name:         "TestPlugin",
		shouldDetect: true,
		shouldError:  false,
	}
	registry.Register(plugin)

	engine := NewEngine(registry)

	credentials := []string{"cred1", "cred2", "cred3"}
	resultCount := 0

	engine.RevokeWithCallback(context.Background(), credentials, func(result RevocationResult) {
		resultCount++
		if !result.Success {
			t.Errorf("callback received failed result: %v", result.Error)
		}
	})

	if resultCount != len(credentials) {
		t.Errorf("callback was called %d times, want %d", resultCount, len(credentials))
	}
}

func TestEngine_ValidateCredential(t *testing.T) {
	tests := []struct {
		name       string
		credential string
		plugin     *mockPlugin
		wantPlugin string
		wantErr    bool
	}{
		{
			name:       "valid credential",
			credential: "test_credential",
			plugin: &mockPlugin{
				name:         "TestPlugin",
				shouldDetect: true,
				shouldError:  false,
			},
			wantPlugin: "TestPlugin",
			wantErr:    false,
		},
		{
			name:       "invalid credential",
			credential: "test_credential",
			plugin: &mockPlugin{
				name:         "TestPlugin",
				shouldDetect: true,
				shouldError:  true,
			},
			wantPlugin: "",
			wantErr:    true,
		},
		{
			name:       "no plugin found",
			credential: "test_credential",
			plugin: &mockPlugin{
				shouldDetect: false,
			},
			wantPlugin: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := plugins.NewRegistry()
			registry.Register(tt.plugin)

			engine := NewEngine(registry)

			pluginName, err := engine.ValidateCredential(tt.credential)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCredential() error = %v, wantErr %v", err, tt.wantErr)
			}

			if pluginName != tt.wantPlugin {
				t.Errorf("ValidateCredential() plugin = %v, want %v", pluginName, tt.wantPlugin)
			}
		})
	}
}

func TestGetStats(t *testing.T) {
	results := []RevocationResult{
		{Success: true, Plugin: "Plugin1"},
		{Success: true, Plugin: "Plugin2"},
		{Success: false, Plugin: "Plugin1"},
		{Success: true, Plugin: "Plugin1"},
	}

	stats := GetStats(results)

	if stats["total"] != 4 {
		t.Errorf("expected total to be 4, got %v", stats["total"])
	}

	if stats["successful"] != 3 {
		t.Errorf("expected successful to be 3, got %v", stats["successful"])
	}

	if stats["failed"] != 1 {
		t.Errorf("expected failed to be 1, got %v", stats["failed"])
	}

	byPlugin := stats["by_plugin"].(map[string]int)
	if byPlugin["Plugin1"] != 3 {
		t.Errorf("expected Plugin1 count to be 3, got %v", byPlugin["Plugin1"])
	}
	if byPlugin["Plugin2"] != 1 {
		t.Errorf("expected Plugin2 count to be 1, got %v", byPlugin["Plugin2"])
	}
}

func TestFormatError(t *testing.T) {
	tests := []struct {
		name   string
		result RevocationResult
		want   string
	}{
		{
			name: "with error",
			result: RevocationResult{
				Plugin: "TestPlugin",
				Error:  errors.New("test error"),
			},
			want: "Error revoking credential (plugin: TestPlugin): test error",
		},
		{
			name: "no error",
			result: RevocationResult{
				Plugin: "TestPlugin",
				Error:  nil,
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatError(tt.result)
			if got != tt.want {
				t.Errorf("FormatError() = %v, want %v", got, tt.want)
			}
		})
	}
}
