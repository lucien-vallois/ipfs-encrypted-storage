package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"ipfs-encrypted-storage/src/errors"
	"ipfs-encrypted-storage/src/ipfs"
)

// NewListCmd creates the list command
func NewListCmd() *cobra.Command {
	var (
		cid string
	)

	cmd := &cobra.Command{
		Use:   "list [cid]",
		Short: "List contents of IPFS directory or show file info",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cid = args[0]
			}
			return runList(cid)
		},
	}

	return cmd
}

func runList(cid string) error {
	client, err := ipfs.NewIPFSClient(GetIPFSURL())
	if err != nil {
		return errors.NewEnhancedError(err, errors.ErrCodeNetworkFailure,
			&errors.ErrorContext{
				Operation: "IPFS_client_creation",
				Resource:  GetIPFSURL(),
				Suggestions: []string{
					"Ensure IPFS daemon is running",
					"Check IPFS API endpoint configuration",
					"Verify network connectivity",
				},
			})
	}
	defer client.Close()

	if cid == "" {
		// List all pins
		pins, err := client.ListPins()
		if err != nil {
			return errors.NewEnhancedError(err, errors.ErrCodeNetworkFailure,
				&errors.ErrorContext{
					Operation: "IPFS_list_pins",
					Resource:  GetIPFSURL(),
					Suggestions: []string{
						"Check IPFS daemon connectivity",
						"Verify IPFS node has pinning capabilities",
						"Try restarting IPFS daemon",
					},
				})
		}

		fmt.Println("Pinned files:")
		for cid, pinType := range pins {
			fmt.Printf("  %s (%s)\n", cid, pinType)
		}
	} else {
		// List directory contents
		entries, err := client.ListDirectory(cid)
		if err != nil {
			// Try to get file stats instead
			stats, err := client.GetObjectStats(cid)
			if err != nil {
				return errors.NewEnhancedError(err, errors.ErrCodeResourceNotFound,
					&errors.ErrorContext{
						Operation: "IPFS_get_info",
						Resource:  cid,
						CID:       cid,
						Suggestions: []string{
							"Verify CID is correct",
							"Check if content exists on IPFS",
							"Try pinning the content first",
						},
					})
			}

			fmt.Printf("File: %s\n", cid)
			fmt.Printf("Size: %d bytes\n", stats.CumulativeSize)
			fmt.Printf("Block Size: %d bytes\n", stats.BlockSize)
			return nil
		}

		fmt.Printf("Directory: %s\n", cid)
		fmt.Println("Contents:")
		for _, entry := range entries {
			fmt.Printf("  %s (%d bytes)\n", entry.Name, entry.Size)
		}
	}

	return nil
}
