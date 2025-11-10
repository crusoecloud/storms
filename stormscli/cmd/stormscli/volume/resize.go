package volume

import (
	"fmt"

	"github.com/spf13/cobra"
	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/utils"
)

func NewResizeVolumeCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resize",
		Short: "Resize a volume. The new size must be greater than the current size.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, conn, err := cmdFactory.StorMSClientProvider(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to create StorMS client: %w", err)
			}
			defer conn.Close()

			err = resizeVolume(cmd, client)
			if err != nil {
				return fmt.Errorf("failed command: %w", err)
			}

			return nil
		},
	}

	utils.NewFlagBuilder(cmd).
		String(idFlag, "", "id of volume", true).
		String(sizeFlag, "", "size of volume in the format 'X[GiB|TiB]' where X is a non-zero integer", true)

	return cmd
}

func resizeVolume(cmd *cobra.Command, client storms.StorageManagementServiceClient) error {
	id := utils.MustGetStringFlag(cmd, idFlag)
	sizeStr := utils.MustGetStringFlag(cmd, sizeFlag)

	sizeBytes, err := utils.ParseSizeString(sizeStr)
	if err != nil {
		return fmt.Errorf("failed to parse size: %w", err)
	}

	_, err = client.ResizeVolume(cmd.Context(), &storms.ResizeVolumeRequest{
		Uuid: id,
		Size: sizeBytes,
	})
	if err != nil {
		return fmt.Errorf("failed to resize volume: %w", err)
	}

	cmd.Printf("Resized volume: %s (size: %s)\n", id, sizeStr)

	return nil
}
