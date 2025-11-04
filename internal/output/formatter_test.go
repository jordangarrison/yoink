package output

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/jordangarrison/yoink/internal/core"
)

func TestHumanFormatter_FormatResult(t *testing.T) {
	tests := []struct {
		name   string
		result core.RevocationResult
		want   []string // strings that should appear in output
	}{
		{
			name: "successful revocation",
			result: core.RevocationResult{
				Credential: "ghp_1234567890abcdefghijklmnopqrstuvwxyz123456",
				Plugin:     "GitHub",
				Success:    true,
				Error:      nil,
				DryRun:     false,
			},
			want: []string{"✓", "Successfully revoked", "GitHub"},
		},
		{
			name: "failed revocation",
			result: core.RevocationResult{
				Credential: "ghp_1234567890abcdefghijklmnopqrstuvwxyz123456",
				Plugin:     "GitHub",
				Success:    false,
				Error:      errors.New("test error"),
				DryRun:     false,
			},
			want: []string{"✗", "Failed", "Error:", "test error"},
		},
		{
			name: "dry-run success",
			result: core.RevocationResult{
				Credential: "ghp_1234567890abcdefghijklmnopqrstuvwxyz123456",
				Plugin:     "GitHub",
				Success:    true,
				Error:      nil,
				DryRun:     true,
			},
			want: []string{"[DRY-RUN]", "✓", "Would revoke", "GitHub"},
		},
		{
			name: "dry-run failure",
			result: core.RevocationResult{
				Credential: "ghp_1234567890abcdefghijklmnopqrstuvwxyz123456",
				Plugin:     "GitHub",
				Success:    false,
				Error:      errors.New("validation error"),
				DryRun:     true,
			},
			want: []string{"[DRY-RUN]", "✗", "Failed", "validation error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewHumanFormatter(&buf)

			output := formatter.FormatResult(tt.result)

			for _, wantStr := range tt.want {
				if !strings.Contains(output, wantStr) {
					t.Errorf("FormatResult() output missing '%s'\nGot: %s", wantStr, output)
				}
			}

			// Verify credential is masked
			if strings.Contains(output, "ghp_1234567890abcdefghijklmnopqrstuvwxyz123456") {
				t.Error("FormatResult() should mask the full credential")
			}
		})
	}
}

func TestHumanFormatter_FormatSummary(t *testing.T) {
	tests := []struct {
		name    string
		results []core.RevocationResult
		want    []string
	}{
		{
			name:    "empty results",
			results: []core.RevocationResult{},
			want:    []string{"No credentials processed"},
		},
		{
			name: "mixed results",
			results: []core.RevocationResult{
				{Plugin: "GitHub", Success: true, DryRun: false},
				{Plugin: "GitHub", Success: true, DryRun: false},
				{Plugin: "GitHub", Success: false, DryRun: false},
				{Plugin: "GitLab", Success: true, DryRun: false},
			},
			want: []string{
				"Revocation Summary",
				"Total processed: 4",
				"Successful: 3",
				"Failed: 1",
				"By plugin:",
				"GitHub: 3",
				"GitLab: 1",
			},
		},
		{
			name: "dry-run mode",
			results: []core.RevocationResult{
				{Plugin: "GitHub", Success: true, DryRun: true},
				{Plugin: "GitHub", Success: true, DryRun: true},
			},
			want: []string{
				"Mode: DRY-RUN",
				"Total processed: 2",
				"Successful: 2",
				"Failed: 0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewHumanFormatter(&buf)

			output := formatter.FormatSummary(tt.results)

			for _, wantStr := range tt.want {
				if !strings.Contains(output, wantStr) {
					t.Errorf("FormatSummary() output missing '%s'\nGot: %s", wantStr, output)
				}
			}
		})
	}
}

func TestHumanFormatter_Write(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewHumanFormatter(&buf)

	testStr := "test output"
	err := formatter.Write(testStr)

	if err != nil {
		t.Errorf("Write() error = %v", err)
	}

	if buf.String() != testStr {
		t.Errorf("Write() wrote '%s', want '%s'", buf.String(), testStr)
	}
}

func TestHumanFormatter_WriteResult(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewHumanFormatter(&buf)

	result := core.RevocationResult{
		Credential: "ghp_test123456789012345678901234567890123456",
		Plugin:     "GitHub",
		Success:    true,
		DryRun:     false,
	}

	err := formatter.WriteResult(result)

	if err != nil {
		t.Errorf("WriteResult() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "✓") || !strings.Contains(output, "Successfully revoked") {
		t.Errorf("WriteResult() output incorrect: %s", output)
	}
}

func TestHumanFormatter_WriteSummary(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewHumanFormatter(&buf)

	results := []core.RevocationResult{
		{Plugin: "GitHub", Success: true, DryRun: false},
	}

	err := formatter.WriteSummary(results)

	if err != nil {
		t.Errorf("WriteSummary() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Revocation Summary") {
		t.Errorf("WriteSummary() output incorrect: %s", output)
	}
}

func TestMaskCredential(t *testing.T) {
	tests := []struct {
		name       string
		credential string
		wantPrefix string
		wantSuffix string
		wantMasked bool
	}{
		{
			name:       "long credential",
			credential: "ghp_1234567890abcdefghijklmnopqrstuvwxyz123456",
			wantPrefix: "ghp_1234",
			wantSuffix: "3456",
			wantMasked: true,
		},
		{
			name:       "short credential",
			credential: "short",
			wantPrefix: "",
			wantSuffix: "",
			wantMasked: true,
		},
		{
			name:       "exactly 12 chars",
			credential: "123456789012",
			wantPrefix: "",
			wantSuffix: "",
			wantMasked: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masked := maskCredential(tt.credential)

			// Should not contain the original credential
			if len(tt.credential) > 12 && masked == tt.credential {
				t.Error("maskCredential() should mask the credential")
			}

			// Should contain asterisks
			if tt.wantMasked && !strings.Contains(masked, "*") {
				t.Error("maskCredential() should contain asterisks")
			}

			// Should have correct prefix and suffix for long credentials
			if len(tt.credential) > 12 {
				if !strings.HasPrefix(masked, tt.wantPrefix) {
					t.Errorf("maskCredential() prefix = %s, want prefix %s", masked[:8], tt.wantPrefix)
				}
				if !strings.HasSuffix(masked, tt.wantSuffix) {
					t.Errorf("maskCredential() suffix = %s, want suffix %s", masked[len(masked)-4:], tt.wantSuffix)
				}
			}
		})
	}
}
