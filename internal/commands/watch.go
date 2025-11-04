package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	"github.com/jordangarrison/yoink/internal/core"
	"github.com/jordangarrison/yoink/internal/output"
	"github.com/jordangarrison/yoink/pkg/plugins"
	"github.com/jordangarrison/yoink/pkg/plugins/github"
)

var (
	watchRecursive bool
	watchPattern   string
)

var watchCmd = &cobra.Command{
	Use:   "watch <path>",
	Short: "Watch a directory for exposed credentials",
	Long: `Watch a file or directory for changes and automatically revoke any exposed credentials found.

Examples:
  # Watch current directory
  yoink watch .

  # Watch specific directory recursively
  yoink watch --recursive /path/to/repo

  # Watch with pattern matching
  yoink watch --pattern "*.log" /var/log

  # Dry-run mode
  yoink watch --dry-run /path/to/watch`,
	Args: cobra.ExactArgs(1),
	RunE: runWatch,
}

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.Flags().BoolVarP(&watchRecursive, "recursive", "r", false, "watch directories recursively")
	watchCmd.Flags().StringVarP(&watchPattern, "pattern", "p", "*", "file pattern to watch (glob)")
}

func runWatch(cmd *cobra.Command, args []string) error {
	path := args[0]

	// Verify path exists
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize plugin registry
	registry := plugins.NewRegistry()
	registry.Register(github.New())

	// Initialize engine
	engine := core.NewEngine(registry)
	engine.SetDryRun(dryRun)

	// Initialize output formatter
	formatter := output.NewHumanFormatter(os.Stdout)

	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	// Add path(s) to watcher
	if info.IsDir() {
		if watchRecursive {
			if err := addRecursive(watcher, path); err != nil {
				return fmt.Errorf("failed to add recursive watch: %w", err)
			}
		} else {
			if err := watcher.Add(path); err != nil {
				return fmt.Errorf("failed to watch directory: %w", err)
			}
		}
	} else {
		if err := watcher.Add(path); err != nil {
			return fmt.Errorf("failed to watch file: %w", err)
		}
	}

	fmt.Fprintf(os.Stderr, "Watching: %s\n", path)
	if dryRun {
		fmt.Fprintf(os.Stderr, "DRY-RUN mode enabled - no credentials will be revoked\n")
	}
	fmt.Fprintf(os.Stderr, "Press Ctrl+C to stop\n\n")

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Watch for file changes
	for {
		select {
		case <-sigCh:
			fmt.Fprintf(os.Stderr, "\nStopping watcher...\n")
			return nil

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Only process write and create events
			if event.Op&fsnotify.Write != fsnotify.Write && event.Op&fsnotify.Create != fsnotify.Create {
				continue
			}

			// Check if file matches pattern
			matched, err := filepath.Match(watchPattern, filepath.Base(event.Name))
			if err != nil || !matched {
				continue
			}

			if verbose {
				fmt.Fprintf(os.Stderr, "File changed: %s\n", event.Name)
			}

			// Scan file for credentials
			credentials, err := scanFileForCredentials(event.Name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error scanning file %s: %v\n", event.Name, err)
				continue
			}

			if len(credentials) > 0 {
				fmt.Fprintf(os.Stderr, "Found %d credential(s) in %s\n", len(credentials), event.Name)

				// Revoke credentials
				engine.RevokeWithCallback(ctx, credentials, func(result core.RevocationResult) {
					if err := formatter.WriteResult(result); err != nil {
						fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
					}
				})
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "Watcher error: %v\n", err)
		}
	}
}

// addRecursive adds all subdirectories to the watcher
func addRecursive(watcher *fsnotify.Watcher, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := watcher.Add(path); err != nil {
				return err
			}
		}
		return nil
	})
}

// scanFileForCredentials scans a file for potential credentials
func scanFileForCredentials(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Pattern to match GitHub tokens
	// This is a simple regex - in production you might want more sophisticated detection
	githubTokenPattern := regexp.MustCompile(`gh[prsouh]_[a-zA-Z0-9]{36,}`)

	matches := githubTokenPattern.FindAllString(string(content), -1)

	// Deduplicate
	seen := make(map[string]bool)
	credentials := make([]string, 0)
	for _, match := range matches {
		if !seen[match] {
			seen[match] = true
			credentials = append(credentials, match)
		}
	}

	return credentials, nil
}
