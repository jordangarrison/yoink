package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/jordangarrison/yoink/internal/core"
)

// Formatter defines the interface for output formatters.
type Formatter interface {
	FormatResult(result core.RevocationResult) string
	FormatSummary(results []core.RevocationResult) string
}

// HumanFormatter formats output in a human-readable format.
type HumanFormatter struct {
	writer io.Writer
}

// NewHumanFormatter creates a new human-readable formatter.
func NewHumanFormatter(w io.Writer) *HumanFormatter {
	return &HumanFormatter{
		writer: w,
	}
}

// FormatResult formats a single revocation result.
func (f *HumanFormatter) FormatResult(result core.RevocationResult) string {
	var sb strings.Builder

	// Mask the credential for security (show first 8 and last 4 chars)
	maskedCred := maskCredential(result.Credential)

	if result.DryRun {
		sb.WriteString("[DRY-RUN] ")
	}

	if result.Success {
		if result.DryRun {
			sb.WriteString(fmt.Sprintf("✓ Would revoke %s credential: %s\n", result.Plugin, maskedCred))
		} else {
			sb.WriteString(fmt.Sprintf("✓ Successfully revoked %s credential: %s\n", result.Plugin, maskedCred))
		}
	} else {
		sb.WriteString(fmt.Sprintf("✗ Failed to revoke credential: %s\n", maskedCred))
		if result.Error != nil {
			sb.WriteString(fmt.Sprintf("  Error: %v\n", result.Error))
		}
	}

	return sb.String()
}

// FormatSummary formats a summary of multiple revocation results.
func (f *HumanFormatter) FormatSummary(results []core.RevocationResult) string {
	if len(results) == 0 {
		return "No credentials processed.\n"
	}

	stats := core.GetStats(results)
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n")
	sb.WriteString("Revocation Summary\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n")

	isDryRun := len(results) > 0 && results[0].DryRun

	if isDryRun {
		sb.WriteString("Mode: DRY-RUN (no actual revocations performed)\n")
	}

	sb.WriteString(fmt.Sprintf("Total processed: %d\n", stats["total"]))
	sb.WriteString(fmt.Sprintf("Successful: %d\n", stats["successful"]))
	sb.WriteString(fmt.Sprintf("Failed: %d\n", stats["failed"]))

	// Show breakdown by plugin
	byPlugin := stats["by_plugin"].(map[string]int)
	if len(byPlugin) > 0 {
		sb.WriteString("\nBy plugin:\n")
		for plugin, count := range byPlugin {
			sb.WriteString(fmt.Sprintf("  %s: %d\n", plugin, count))
		}
	}

	sb.WriteString(strings.Repeat("=", 50) + "\n")

	return sb.String()
}

// Write writes the formatted output to the writer.
func (f *HumanFormatter) Write(s string) error {
	_, err := f.writer.Write([]byte(s))
	return err
}

// WriteResult formats and writes a single result.
func (f *HumanFormatter) WriteResult(result core.RevocationResult) error {
	return f.Write(f.FormatResult(result))
}

// WriteSummary formats and writes a summary.
func (f *HumanFormatter) WriteSummary(results []core.RevocationResult) error {
	return f.Write(f.FormatSummary(results))
}

// maskCredential masks a credential for display, showing only the first 8 and last 4 characters.
func maskCredential(credential string) string {
	if len(credential) <= 12 {
		return strings.Repeat("*", len(credential))
	}

	prefix := credential[:8]
	suffix := credential[len(credential)-4:]
	middle := strings.Repeat("*", len(credential)-12)

	return prefix + middle + suffix
}

// JSONFormatter formats output in JSON format (placeholder for future implementation).
type JSONFormatter struct {
	writer io.Writer
}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter(w io.Writer) *JSONFormatter {
	return &JSONFormatter{
		writer: w,
	}
}

// FormatResult formats a single revocation result as JSON.
func (f *JSONFormatter) FormatResult(result core.RevocationResult) string {
	// TODO: Implement JSON formatting for future webhook server
	return ""
}

// FormatSummary formats a summary as JSON.
func (f *JSONFormatter) FormatSummary(results []core.RevocationResult) string {
	// TODO: Implement JSON formatting for future webhook server
	return ""
}
