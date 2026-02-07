package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version information (set via ldflags during build)
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "machinery-registry-api",
	Short: "Machinery Registry API - Read-only claim registry service",
	Long: `Machinery Registry API serves the claim inventory from a GitHub-hosted
registry.yaml file via a lightweight REST API.

By default, it starts the API server. Use subcommands for other operations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return serverCmd.RunE(cmd, args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
