package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
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
	// TODO: wire syncer + API server in Step 5
	fmt.Println("server not yet implemented")
	return nil
}
