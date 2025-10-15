package volume

import (
	"github.com/spf13/cobra"
	"gitlab.com/crusoeenergy/island/storage/storms/cmd/utils"
)

func NewVolumeCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	volumesCmd := &cobra.Command{
		Use:   "volume",
		Short: "Manage a single volume.",
	}

	volumesCmd.AddCommand(
		NewAttachVolumeCmd(cmdFactory),
		NewCreateVolumeCmd(cmdFactory),
		NewDeleteVolumeCmd(cmdFactory),
		NewDetachVolumeCmd(cmdFactory),
		NewGetVolumeCmd(cmdFactory),
		NewResizeVolumeCmd(cmdFactory),
	)

	return volumesCmd
}
