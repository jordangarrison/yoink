package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/jordangarrison/yoink/pkg/plugins"
)

// RevocationResult represents the result of a credential revocation attempt.
type RevocationResult struct {
	Credential string
	Plugin     string
	Success    bool
	Error      error
	DryRun     bool
}

// Engine is the core revocation engine that processes credentials.
type Engine struct {
	registry *plugins.Registry
	dryRun   bool
}

// NewEngine creates a new revocation engine.
func NewEngine(registry *plugins.Registry) *Engine {
	return &Engine{
		registry: registry,
		dryRun:   false,
	}
}

// SetDryRun enables or disables dry-run mode.
func (e *Engine) SetDryRun(dryRun bool) {
	e.dryRun = dryRun
}

// IsDryRun returns whether dry-run mode is enabled.
func (e *Engine) IsDryRun() bool {
	return e.dryRun
}

// Revoke processes a single credential and attempts to revoke it.
func (e *Engine) Revoke(ctx context.Context, credential string) RevocationResult {
	// Find appropriate plugin
	plugin, found := e.registry.FindPlugin(credential)
	if !found {
		return RevocationResult{
			Credential: credential,
			Plugin:     "none",
			Success:    false,
			Error:      plugins.ErrNoPluginFound,
			DryRun:     e.dryRun,
		}
	}

	// In dry-run mode, only validate
	if e.dryRun {
		err := plugin.Validate(credential)
		return RevocationResult{
			Credential: credential,
			Plugin:     plugin.Name(),
			Success:    err == nil,
			Error:      err,
			DryRun:     true,
		}
	}

	// Attempt revocation
	err := plugin.Revoke(ctx, credential)
	return RevocationResult{
		Credential: credential,
		Plugin:     plugin.Name(),
		Success:    err == nil,
		Error:      err,
		DryRun:     false,
	}
}

// RevokeBatch processes multiple credentials concurrently.
func (e *Engine) RevokeBatch(ctx context.Context, credentials []string) []RevocationResult {
	results := make([]RevocationResult, len(credentials))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, cred := range credentials {
		wg.Add(1)
		go func(index int, credential string) {
			defer wg.Done()

			result := e.Revoke(ctx, credential)

			mu.Lock()
			results[index] = result
			mu.Unlock()
		}(i, cred)
	}

	wg.Wait()
	return results
}

// RevokeWithCallback processes multiple credentials and calls a callback for each result.
// This is useful for streaming results as they complete.
func (e *Engine) RevokeWithCallback(ctx context.Context, credentials []string, callback func(RevocationResult)) {
	var wg sync.WaitGroup

	for _, cred := range credentials {
		wg.Add(1)
		go func(credential string) {
			defer wg.Done()

			result := e.Revoke(ctx, credential)
			callback(result)
		}(cred)
	}

	wg.Wait()
}

// ValidateCredential validates a credential format without attempting revocation.
func (e *Engine) ValidateCredential(credential string) (string, error) {
	plugin, found := e.registry.FindPlugin(credential)
	if !found {
		return "", plugins.ErrNoPluginFound
	}

	err := plugin.Validate(credential)
	if err != nil {
		return "", err
	}

	return plugin.Name(), nil
}

// GetStats returns statistics about the revocation results.
func GetStats(results []RevocationResult) map[string]interface{} {
	total := len(results)
	successful := 0
	failed := 0
	byPlugin := make(map[string]int)

	for _, result := range results {
		if result.Success {
			successful++
		} else {
			failed++
		}
		byPlugin[result.Plugin]++
	}

	return map[string]interface{}{
		"total":      total,
		"successful": successful,
		"failed":     failed,
		"by_plugin":  byPlugin,
	}
}

// FormatError formats an error message for display.
func FormatError(result RevocationResult) string {
	if result.Error == nil {
		return ""
	}

	return fmt.Sprintf("Error revoking credential (plugin: %s): %v", result.Plugin, result.Error)
}
