package volume

import (
	"fmt"

	"github.com/spf13/cobra"

	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/utils"
)

func NewGetVolumeCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a volume.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, conn, err := cmdFactory.StorMSClientProvider(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to create StorMS client: %w", err)
			}
			defer conn.Close()

			err = getVolume(cmd, client)
			if err != nil {
				return fmt.Errorf("failed command: %w", err)
			}

			return nil
		},
	}

	utils.NewFlagBuilder(cmd).
		String(idFlag, "", "uuid of volume", true)

	return cmd
}

func getVolume(cmd *cobra.Command, client storms.StorageManagementServiceClient) error {
	id := utils.MustGetStringFlag(cmd, idFlag)
	resp, err := client.GetVolume(cmd.Context(), &storms.GetVolumeRequest{
		Uuid: id,
	})
	if err != nil {
		return fmt.Errorf("failed to get volume [id=%s]: %w", id, err)
	}

	if err := utils.RenderVolumes([]*storms.Volume{resp.Volume}); err != nil {
		return fmt.Errorf("failed to render volumes: %w", err)
	}

	return nil
}
