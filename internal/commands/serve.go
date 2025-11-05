package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/jordangarrison/yoink/internal/core"
	"github.com/jordangarrison/yoink/internal/output"
	"github.com/jordangarrison/yoink/pkg/plugins"
	"github.com/jordangarrison/yoink/pkg/plugins/github"
)

var (
	servePort string
	serveHost string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start webhook server to receive credential revocation requests",
	Long: `Start an HTTP server that listens for webhook notifications about exposed credentials.

The server exposes the following endpoints:
  POST /revoke - Revoke credentials sent in the request body

Request format:
  {
    "credentials": ["ghp_xxx", "gho_yyy"]
  }

Response format:
  {
    "total": 2,
    "successful": 2,
    "failed": 0,
    "results": [...]
  }

Examples:
  # Start server on default port 8080
  yoink serve

  # Start server on custom port
  yoink serve --port 3000

  # Start server with custom host
  yoink serve --host 0.0.0.0 --port 8080`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&servePort, "port", "p", "8080", "port to listen on")
	serveCmd.Flags().StringVarP(&serveHost, "host", "H", "localhost", "host to bind to")
}

func runServe(cmd *cobra.Command, args []string) error {
	// Initialize plugin registry
	registry := plugins.NewRegistry()
	registry.Register(github.New())

	// Initialize engine
	engine := core.NewEngine(registry)
	engine.SetDryRun(dryRun)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/revoke", makeRevokeHandler(engine))
	mux.HandleFunc("/health", healthHandler)

	addr := fmt.Sprintf("%s:%s", serveHost, servePort)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Setup graceful shutdown
	serverErrors := make(chan error, 1)
	go func() {
		fmt.Fprintf(os.Stderr, "Yoink webhook server listening on %s\n", addr)
		if dryRun {
			fmt.Fprintf(os.Stderr, "DRY-RUN mode enabled - no credentials will be revoked\n")
		}
		fmt.Fprintf(os.Stderr, "Endpoints:\n")
		fmt.Fprintf(os.Stderr, "  POST http://%s/revoke  - Revoke credentials\n", addr)
		fmt.Fprintf(os.Stderr, "  GET  http://%s/health  - Health check\n\n", addr)
		fmt.Fprintf(os.Stderr, "Press Ctrl+C to stop\n\n")

		serverErrors <- server.ListenAndServe()
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case <-sigCh:
		fmt.Fprintf(os.Stderr, "\nShutting down server...\n")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Server stopped\n")
		return nil
	}
}

// RevokeRequest represents the request body for the /revoke endpoint
type RevokeRequest struct {
	Credentials []string `json:"credentials"`
}

// RevokeResponse represents the response body for the /revoke endpoint
type RevokeResponse struct {
	Total      int                      `json:"total"`
	Successful int                      `json:"successful"`
	Failed     int                      `json:"failed"`
	DryRun     bool                     `json:"dry_run"`
	Results    []core.RevocationResult  `json:"results"`
}

func makeRevokeHandler(engine *core.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer func() { _ = r.Body.Close() }()

		var req RevokeRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if len(req.Credentials) == 0 {
			http.Error(w, "No credentials provided", http.StatusBadRequest)
			return
		}

		// Log request
		if verbose {
			fmt.Fprintf(os.Stderr, "[%s] Received revocation request for %d credential(s)\n",
				time.Now().Format(time.RFC3339), len(req.Credentials))
		}

		// Revoke credentials
		ctx := context.Background()
		results := engine.RevokeBatch(ctx, req.Credentials)

		// Log results
		for _, result := range results {
			formatter := output.NewHumanFormatter(os.Stderr)
			_ = formatter.WriteResult(result)
		}

		// Prepare response
		stats := core.GetStats(results)
		response := RevokeResponse{
			Total:      stats["total"].(int),
			Successful: stats["successful"].(int),
			Failed:     stats["failed"].(int),
			DryRun:     engine.IsDryRun(),
			Results:    results,
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if response.Failed > 0 {
			w.WriteHeader(http.StatusPartialContent)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		_ = json.NewEncoder(w).Encode(response)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "yoink",
	})
}
