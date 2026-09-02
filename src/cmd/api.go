// Package cmd provides CLI command definitions for the IPFS encrypted storage system
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"ipfs-encrypted-storage/src/api"
	"ipfs-encrypted-storage/src/config"
)

// NewAPIServerCmd creates the API server command
func NewAPIServerCmd() *cobra.Command {
	var (
		host string
		port int
	)

	cmd := &cobra.Command{
		Use:   "api",
		Short: "Start REST API server",
		Long:  `Start the REST API server for programmatic access to IPFS encrypted storage operations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAPIServer(host, port)
		},
	}

	cmd.Flags().StringVar(&host, "host", "localhost", "Server host address")
	cmd.Flags().IntVar(&port, "port", 8080, "Server port")

	return cmd
}

func runAPIServer(host string, port int) error {
	if os.Getenv("IPFS_API_KEY") == "" {
		return fmt.Errorf("IPFS_API_KEY must be set")
	}

	// Get configuration
	cfg := GetGlobalConfig()
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	fmt.Printf("Starting Simple API server on %s:%d\n", host, port)
	fmt.Println("API endpoints available at:")
	fmt.Printf("  Health: http://%s:%d/api/v1/health\n", host, port)
	fmt.Printf("  Files:  http://%s:%d/api/v1/files\n", host, port)
	fmt.Printf("  P2P:    http://%s:%d/api/v1/p2p\n", host, port)
	fmt.Println()
	fmt.Println("Use the configured IPFS_API_KEY in the X-API-Key header")
	fmt.Println("Press Ctrl+C to stop the server")

	return api.RunSimple(cfg, host, port)
}
