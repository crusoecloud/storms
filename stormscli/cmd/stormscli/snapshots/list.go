package snapshots

import (
	"fmt"

	"github.com/spf13/cobra"

	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/utils"
)

const idsFlag = "ids"

func NewListSnapshotsCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List snapshot(s).",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, conn, err := cmdFactory.StorMSClientProvider(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to create StorMS client: %w", err)
			}
			defer conn.Close()

			err = listSnapshots(cmd, client)
			if err != nil {
				return fmt.Errorf("failed command: %w", err)
			}

			return nil
		},
	}

	utils.NewFlagBuilder(cmd).
		StringCSV(idsFlag, "", "uuid of snapshot(s)", false)

	return cmd
}

func listSnapshots(cmd *cobra.Command, client storms.StorageManagementServiceClient) error {
	ids := utils.MustGetStringCSVFlag(cmd, idsFlag)

	var snapshots []*storms.Snapshot
	if len(ids) != 0 {
		snapshots = make([]*storms.Snapshot, 0)
		for _, volID := range ids {
			resp, err := client.GetSnapshot(cmd.Context(), &storms.GetSnapshotRequest{
				Uuid: volID,
			})

			if err != nil {
				return fmt.Errorf("failed to get snapshot [id=%s]: %w", volID, err)
			}
			snapshots = append(snapshots, resp.Snapshot)
		}
	} else {
		resp, err := client.GetSnapshots(cmd.Context(), &storms.GetSnapshotsRequest{})
		if err != nil {
			return fmt.Errorf("failed to get snapshots: %w", err)
		}
		snapshots = resp.Snapshots
	}

	if err := utils.RenderSnapshots(snapshots); err != nil {
		return fmt.Errorf("failed to print snapshot table: %w", err)
	}

	return nil
}
