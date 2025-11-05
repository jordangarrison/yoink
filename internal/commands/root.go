package commands

import (
	"github.com/spf13/cobra"
)

var (
	// Global flags
	dryRun  bool
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "yoink",
	Short: "Yoink - Automatically revoke exposed credentials",
	Long: `Yoink is a tool that automatically detects and revokes exposed credentials.

It supports multiple input methods:
  - Direct input via command line arguments or stdin
  - File watching to monitor directories for exposed credentials
  - Webhook server to receive credential leak notifications

Currently supports:
  - GitHub Personal Access Tokens (PATs)
  - GitHub OAuth Tokens
  - More providers coming soon!`,
	Version: "0.0.1",
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "validate credentials without actually revoking them")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
