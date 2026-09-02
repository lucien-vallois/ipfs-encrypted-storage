// Package cmd provides CLI command definitions for the IPFS encrypted storage system
package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"ipfs-encrypted-storage/src/config"
	"ipfs-encrypted-storage/src/utils"
)

var (
	ipfsURL    string
	password   string
	verbose    bool
	configFile string
	appConfig  *config.Config
)

// NewRootCmd creates the root command for the CLI
func NewRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "ipfs-encrypted-storage",
		Short: "Encrypted IPFS storage system",
		Long:  `Encrypted storage using an external IPFS daemon, with a local in-memory P2P stub`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Load configuration
			var err error
			appConfig, err = config.LoadConfig(configFile)
			if err != nil {
				logrus.WithError(err).Warn("Failed to load config, using defaults")
				appConfig = config.DefaultConfig()
			}

			// Validate configuration
			if err := config.ValidateConfig(appConfig); err != nil {
				logrus.WithError(err).Warn("Invalid configuration, using defaults")
				appConfig = config.DefaultConfig()
			}

			// Apply configuration defaults to flags if not set
			if ipfsURL == "localhost:5001" && appConfig.IPFS.URL != "" {
				ipfsURL = appConfig.IPFS.URL
			}

			// Validate IPFS URL
			if ipfsURL != "" && ipfsURL != "localhost:5001" {
				if err := utils.ValidateIPFSEndpoint(ipfsURL); err != nil {
					logrus.WithError(err).Warn("Invalid IPFS URL in configuration")
					// Don't fail here, just warn and use default
					ipfsURL = "localhost:5001"
				}
			}

			// Configure logging
			if verbose {
				logrus.SetLevel(logrus.DebugLevel)
			} else {
				// Set log level from config
				switch appConfig.Logging.Level {
				case "debug":
					logrus.SetLevel(logrus.DebugLevel)
				case "info":
					logrus.SetLevel(logrus.InfoLevel)
				case "warn":
					logrus.SetLevel(logrus.WarnLevel)
				case "error":
					logrus.SetLevel(logrus.ErrorLevel)
				}
			}

			// Configure log format
			if appConfig.Logging.Format == "json" {
				logrus.SetFormatter(&logrus.JSONFormatter{})
			} else {
				logrus.SetFormatter(&logrus.TextFormatter{})
			}

			// Set output file if specified
			if appConfig.Logging.OutputFile != "" {
				// Note: This would need proper file handling, but keeping simple for now
				logrus.Info("Log file output configured but not implemented in this refactor")
			}
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&ipfsURL, "ipfs-url", "localhost:5001", "IPFS API URL")
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "Encryption password (will prompt if not provided)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Config file path")

	return rootCmd
}

// GetGlobalConfig returns the global application configuration
func GetGlobalConfig() *config.Config {
	return appConfig
}

// GetIPFSURL returns the configured IPFS URL
func GetIPFSURL() string {
	return ipfsURL
}

// GetPassword returns the configured password
func GetPassword() string {
	return password
}
