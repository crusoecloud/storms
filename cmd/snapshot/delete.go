package snapshot

import (
	"fmt"

	"github.com/spf13/cobra"

	storms "gitlab.com/crusoeenergy/island/storage/storms/api/gen/go/storms/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/cmd/utils"
)

func NewDeleteSnapshotCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a snapshot.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, conn, err := cmdFactory.StorMSClientProvider(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to create StorMS client: %w", err)
			}
			defer conn.Close()

			err = deleteSnapshot(cmd, client)
			if err != nil {
				return fmt.Errorf("failed to delete snapshot: %w", err)
			}

			return nil
		},
	}

	utils.NewFlagBuilder(cmd).
		String("id", "", "id of snapshot", true)

	return cmd
}

func deleteSnapshot(cmd *cobra.Command, client storms.StorageManagementServiceClient) error {
	id := utils.MustGetStringFlag(cmd, "id")

	_, err := client.DeleteSnapshot(cmd.Context(), &storms.DeleteSnapshotRequest{
		Uuid: id,
	})
	if err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	cmd.Printf("Deleted snapshot: %s\n", id)

	return nil
}
