// Package main provides the CLI interface for the encrypted IPFS storage system
package main

import (
	"github.com/sirupsen/logrus"
	"ipfs-encrypted-storage/src/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()

	// Add all commands
	rootCmd.AddCommand(cmd.NewUploadCmd())
	rootCmd.AddCommand(cmd.NewDownloadCmd())
	rootCmd.AddCommand(cmd.NewListCmd())
	rootCmd.AddCommand(cmd.NewDeleteCmd())
	rootCmd.AddCommand(cmd.NewP2PCmd())
	rootCmd.AddCommand(cmd.NewInitCmd())
	rootCmd.AddCommand(cmd.NewVerifyCmd())
	rootCmd.AddCommand(cmd.NewAPIServerCmd())

	if err := rootCmd.Execute(); err != nil {
		logrus.Error(err)
	}
}
