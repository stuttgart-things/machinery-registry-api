package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/stuttgart-things/machinery-registry-api/internal/api"
	isync "github.com/stuttgart-things/machinery-registry-api/internal/sync"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the API server",
	Long:  `Start the HTTP API server for serving the claim registry.`,
	RunE:  runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

func runServer(cmd *cobra.Command, args []string) error {
	fmt.Println(logo)
	fmt.Printf("Version:    %s\n", Version)
	fmt.Printf("Commit:     %s\n", Commit)
	fmt.Printf("Build Date: %s\n\n", Date)

	// Resolve configuration from environment
	repo := os.Getenv("REGISTRY_REPO")
	if repo == "" {
		return fmt.Errorf("REGISTRY_REPO environment variable is required")
	}

	regPath := os.Getenv("REGISTRY_PATH")
	if regPath == "" {
		regPath = "claims/registry.yaml"
	}

	branch := os.Getenv("REGISTRY_BRANCH")
	if branch == "" {
		branch = "main"
	}

	interval := 60 * time.Second
	if v := os.Getenv("SYNC_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("invalid SYNC_INTERVAL %q: %w", v, err)
		}
		interval = d
	}

	token := os.Getenv("GITHUB_TOKEN")

	fmt.Printf("Registry:   %s/%s@%s\n", repo, regPath, branch)
	fmt.Printf("Sync:       every %s\n", interval)

	// Create and run initial sync
	syncer := isync.NewSyncer(isync.Config{
		Repo:     repo,
		Path:     regPath,
		Branch:   branch,
		Token:    token,
		Interval: interval,
	})

	ctx := context.Background()
	if err := syncer.InitialSync(ctx); err != nil {
		return fmt.Errorf("initial sync failed: %w", err)
	}

	// Start background sync
	syncer.Start(ctx)

	// Create and start API server
	server := api.NewServer(syncer)

	go func() {
		if err := server.Start(); err != nil {
			if err.Error() != "http: Server closed" {
				log.Printf("Server error: %v", err)
			}
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("\nAPI server listening on http://localhost:%s\n", port)
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  GET  /health                     - Health check")
	fmt.Println("  GET  /version                    - Version info")
	fmt.Println("  GET  /api/v1/claims              - List claims")
	fmt.Println("  GET  /api/v1/claims/{name}       - Get claim by name")
	fmt.Println("  GET  /openapi.yaml               - OpenAPI spec")
	fmt.Println("  GET  /docs                       - API docs")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	fmt.Printf("\nReceived signal: %v\n", sig)

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	syncer.Stop()

	if err := server.Stop(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	fmt.Println("Server stopped gracefully")
	return nil
}
