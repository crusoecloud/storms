package snapshot

import (
	"fmt"

	"github.com/spf13/cobra"

	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/utils"
)

func NewCreateSnapshotCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create ",
		Short: "Create a snapshot.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, conn, err := cmdFactory.StorMSClientProvider(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to create StorMS client: %w", err)
			}
			defer conn.Close()

			err = createSnapshot(cmd, client)
			if err != nil {
				return fmt.Errorf("command failed: %w", err)
			}

			return nil
		},
	}

	utils.NewFlagBuilder(cmd).
		String(idFlag, "", "id of snapshot", true).
		String(srcVolIDFlag, "", "source volume uuid", true)

	return cmd
}

func createSnapshot(cmd *cobra.Command, client storms.StorageManagementServiceClient) error {
	id := utils.MustGetStringFlag(cmd, idFlag)
	srcVolID := utils.MustGetStringFlag(cmd, srcVolIDFlag)

	_, err := client.CreateSnapshot(cmd.Context(), &storms.CreateSnapshotRequest{
		Uuid:          id,
		SrcVolumeUuid: srcVolID,
	})
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	cmd.Printf("Created snapshot: %s\n", id)

	return nil
}
