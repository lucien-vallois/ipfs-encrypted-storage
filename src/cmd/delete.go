package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"ipfs-encrypted-storage/src/errors"
	"ipfs-encrypted-storage/src/ipfs"
)

// NewDeleteCmd creates the delete command
func NewDeleteCmd() *cobra.Command {
	var (
		cid string
	)

	cmd := &cobra.Command{
		Use:   "delete [cid]",
		Short: "Unpin a file from local IPFS node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cid = args[0]
			return runDelete(cid)
		},
	}

	return cmd
}

func runDelete(cid string) error {
	client, err := ipfs.NewIPFSClient(GetIPFSURL())
	if err != nil {
		return errors.NewEnhancedError(err, errors.ErrCodeNetworkFailure,
			&errors.ErrorContext{
				Operation: "IPFS_client_creation",
				Resource:  GetIPFSURL(),
				CID:       cid,
				Suggestions: []string{
					"Ensure IPFS daemon is running",
					"Check IPFS API endpoint configuration",
					"Verify network connectivity",
				},
			})
	}
	defer client.Close()

	err = client.UnpinFile(cid)
	if err != nil {
		return errors.NewEnhancedError(err, errors.ErrCodeNetworkFailure,
			&errors.ErrorContext{
				Operation: "IPFS_unpin",
				Resource:  cid,
				CID:       cid,
				Suggestions: []string{
					"Verify CID is pinned on this node",
					"Check IPFS daemon permissions",
					"Ensure file was uploaded through this node",
				},
			})
	}

	fmt.Printf("File %s unpinned successfully\n", cid)
	fmt.Println("Note: File may still be available from other peers")

	return nil
}
