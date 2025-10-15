package snapshot

import (
	"fmt"

	"github.com/spf13/cobra"

	storms "gitlab.com/crusoeenergy/island/storage/storms/api/gen/go/storms/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/cmd/utils"
)

func NewGetSnapshotCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a snapshot.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, conn, err := cmdFactory.StorMSClientProvider(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to create StorMS client: %w", err)
			}
			defer conn.Close()

			err = getSnapshots(cmd, client)
			if err != nil {
				return fmt.Errorf("failed command: %w", err)
			}

			return nil
		},
	}

	utils.NewFlagBuilder(cmd).
		String(idFlag, "", "uuid of snapshot", true)

	return cmd
}

func getSnapshots(cmd *cobra.Command, client storms.StorageManagementServiceClient) error {
	id := utils.MustGetStringFlag(cmd, idFlag)

	resp, err := client.GetSnapshot(cmd.Context(), &storms.GetSnapshotRequest{
		Uuid: id,
	})
	if err != nil {
		return fmt.Errorf("failed to get snapshot [id=%s]: %w", id, err)
	}

	if err := utils.RenderSnapshots([]*storms.Snapshot{resp.Snapshot}); err != nil {
		return fmt.Errorf("failed to render snapshots: %w", err)
	}

	return nil
}
