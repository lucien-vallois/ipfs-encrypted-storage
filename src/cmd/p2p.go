package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"ipfs-encrypted-storage/src/errors"
	"ipfs-encrypted-storage/src/handlers"
	"ipfs-encrypted-storage/src/p2p"
)

// NewP2PCmd creates the P2P command
func NewP2PCmd() *cobra.Command {
	var (
		listenAddr string
		bootstrap  bool
	)

	cmd := &cobra.Command{
		Use:   "p2p",
		Short: "Start the local in-memory P2P stub",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runP2P(
				listenAddr,
				bootstrap,
				cmd.Flags().Changed("listen"),
				cmd.Flags().Changed("bootstrap"),
			)
		},
	}

	cmd.Flags().StringVar(&listenAddr, "listen", "/ip4/0.0.0.0/tcp/0", "Address to validate and expose from the local stub")
	cmd.Flags().BoolVar(&bootstrap, "bootstrap", true, "Validate configured bootstrap peer addresses")

	return cmd
}

func runP2P(listenAddr string, bootstrap, listenFlagSet, bootstrapFlagSet bool) error {
	// Validate listen address if provided
	if listenAddr != "/ip4/0.0.0.0/tcp/0" {
		// Basic validation for listen address format
		if !strings.HasPrefix(listenAddr, "/ip4/") && !strings.HasPrefix(listenAddr, "/ip6/") {
			return errors.NewEnhancedError(
				fmt.Errorf("invalid listen address format"),
				errors.ErrCodeInvalidInput,
				&errors.ErrorContext{
					Operation: "p2p_listen_validation",
					Resource:  listenAddr,
					Suggestions: []string{
						"Use format: /ip4/0.0.0.0/tcp/PORT",
						"Example: /ip4/0.0.0.0/tcp/4001",
					},
				})
		}
	}

	// Explicit flags take precedence over configuration values.
	if !listenFlagSet && GetGlobalConfig() != nil && GetGlobalConfig().P2P.ListenAddr != "" {
		listenAddr = GetGlobalConfig().P2P.ListenAddr
	}
	if !bootstrapFlagSet && GetGlobalConfig() != nil {
		bootstrap = GetGlobalConfig().P2P.Bootstrap
	}

	node, err := p2p.NewP2PNode(listenAddr)
	if err != nil {
		return fmt.Errorf("failed to create P2P node: %w", err)
	}
	defer node.Close()

	fmt.Printf("P2P stub initialized with ID: %s\n", node.GetID())
	fmt.Println("Configured address:")
	for _, addr := range node.GetAddresses() {
		fmt.Printf("  %s\n", addr)
	}

	if bootstrap {
		// Use bootstrap peers from config if available
		bootstrapPeers := p2p.DefaultBootstrapPeers
		if GetGlobalConfig() != nil && len(GetGlobalConfig().P2P.BootstrapPeers) > 0 {
			bootstrapPeers = GetGlobalConfig().P2P.BootstrapPeers
		}

		fmt.Println("Validating bootstrap peer addresses...")
		err = node.Bootstrap(bootstrapPeers)
		if err != nil {
			fmt.Printf("Warning: Invalid bootstrap peer configuration: %v\n", err)
		}
	}

	// Register message handlers
	node.RegisterMessageHandler(p2p.ProtocolStorage, handlers.HandleStorageMessage)
	node.RegisterMessageHandler(p2p.ProtocolFileRequest, handlers.HandleFileRequest)

	// Subscribe to topics
	_, err = node.SubscribeToTopic("encrypted-storage", handlers.HandleTopicMessage)
	if err != nil {
		fmt.Printf("Warning: Failed to subscribe to topic: %v\n", err)
	}

	fmt.Println("Press Enter to stop...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	return nil
}
