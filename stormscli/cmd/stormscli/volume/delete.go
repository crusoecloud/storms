package volume

import (
	"fmt"

	"github.com/spf13/cobra"
	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/utils"
)

func NewDeleteVolumeCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a volume.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, conn, err := cmdFactory.StorMSClientProvider(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to create StorMS client: %w", err)
			}
			defer conn.Close()

			err = deleteVolume(cmd, client)
			if err != nil {
				return fmt.Errorf("failed command: %w", err)
			}

			return nil
		},
	}

	utils.NewFlagBuilder(cmd).
		String(idFlag, "", "id of volume", true)

	return cmd
}

func deleteVolume(cmd *cobra.Command, client storms.StorageManagementServiceClient) error {
	id := utils.MustGetStringFlag(cmd, idFlag)
	_, err := client.DeleteVolume(cmd.Context(), &storms.DeleteVolumeRequest{
		Uuid: id,
	})
	if err != nil {
		return fmt.Errorf("failed to delete volume: %w", err)
	}

	cmd.Printf("Deleted volume: %s\n", id)

	return nil
}
