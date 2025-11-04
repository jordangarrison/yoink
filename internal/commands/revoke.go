package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jordangarrison/yoink/internal/core"
	"github.com/jordangarrison/yoink/internal/output"
	"github.com/jordangarrison/yoink/pkg/plugins"
	"github.com/jordangarrison/yoink/pkg/plugins/github"
)

var revokeCmd = &cobra.Command{
	Use:   "revoke [credentials...]",
	Short: "Revoke one or more credentials",
	Long: `Revoke exposed credentials by passing them as arguments or via stdin.

Examples:
  # Revoke a single credential
  yoink revoke ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

  # Revoke multiple credentials
  yoink revoke ghp_xxx gho_yyy

  # Revoke from stdin
  echo "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" | yoink revoke

  # Dry-run mode (validate without revoking)
  yoink revoke --dry-run ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`,
	RunE: runRevoke,
}

func init() {
	rootCmd.AddCommand(revokeCmd)
}

func runRevoke(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Initialize plugin registry
	registry := plugins.NewRegistry()
	registry.Register(github.New())

	// Initialize engine
	engine := core.NewEngine(registry)
	engine.SetDryRun(dryRun)

	// Initialize output formatter
	formatter := output.NewHumanFormatter(os.Stdout)

	// Collect credentials from args and stdin
	credentials := collectCredentials(args)

	if len(credentials) == 0 {
		return fmt.Errorf("no credentials provided. Use 'yoink revoke --help' for usage")
	}

	if verbose {
		if dryRun {
			fmt.Fprintf(os.Stderr, "DRY-RUN mode enabled - no credentials will be revoked\n")
		}
		fmt.Fprintf(os.Stderr, "Processing %d credential(s)...\n\n", len(credentials))
	}

	// Revoke credentials with callback for streaming output
	var results []core.RevocationResult
	engine.RevokeWithCallback(ctx, credentials, func(result core.RevocationResult) {
		results = append(results, result)
		if err := formatter.WriteResult(result); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		}
	})

	// Print summary
	if len(results) > 1 || verbose {
		if err := formatter.WriteSummary(results); err != nil {
			return fmt.Errorf("error writing summary: %w", err)
		}
	}

	// Exit with error if any revocations failed
	stats := core.GetStats(results)
	if stats["failed"].(int) > 0 {
		return fmt.Errorf("some credentials failed to revoke")
	}

	return nil
}

// collectCredentials collects credentials from command line args and stdin
func collectCredentials(args []string) []string {
	credentials := make([]string, 0)

	// Add credentials from arguments
	for _, arg := range args {
		credential := strings.TrimSpace(arg)
		if credential != "" {
			credentials = append(credentials, credential)
		}
	}

	// Check if stdin has data (not a terminal)
	stat, err := os.Stdin.Stat()
	if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			credential := strings.TrimSpace(scanner.Text())
			if credential != "" {
				credentials = append(credentials, credential)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
		}
	}

	return credentials
}
