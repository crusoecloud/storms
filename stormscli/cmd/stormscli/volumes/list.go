package volumes

import (
	"fmt"

	"github.com/spf13/cobra"

	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/utils"
)

const idsFlag = "ids"

func NewListVolumesCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List volume(s).",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, conn, err := cmdFactory.StorMSClientProvider(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to create StorMS client: %w", err)
			}
			defer conn.Close()

			err = listVolumes(cmd, client)
			if err != nil {
				return fmt.Errorf("failed command: %w", err)
			}

			return nil
		},
	}

	utils.NewFlagBuilder(cmd).
		StringCSV(idsFlag, "", "uuid of volumes", false)

	return cmd
}

func listVolumes(cmd *cobra.Command, client storms.StorageManagementServiceClient) error {
	ids := utils.MustGetStringCSVFlag(cmd, idsFlag)

	var volumes []*storms.Volume
	if len(ids) != 0 {
		volumes = make([]*storms.Volume, 0)
		for _, volID := range ids {
			volumes = make([]*storms.Volume, 0)
			resp, err := client.GetVolume(cmd.Context(), &storms.GetVolumeRequest{
				Uuid: volID,
			})

			if err != nil {
				return fmt.Errorf("failed to get volume [id=%s]: %w", volID, err)
			}
			volumes = append(volumes, resp.Volume)
		}
	} else {
		resp, err := client.GetVolumes(cmd.Context(), &storms.GetVolumesRequest{})
		if err != nil {
			return fmt.Errorf("failed to get volumes: %w", err)
		}
		volumes = resp.Volumes
	}

	if err := utils.RenderVolumes(volumes); err != nil {
		return fmt.Errorf("failed to render volumes: %w", err)
	}

	return nil
}
