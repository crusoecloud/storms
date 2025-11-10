package volumes

import (
	"github.com/spf13/cobra"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/utils"
)

func NewVolumesCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	volumesCmd := &cobra.Command{
		Use:   "volumes",
		Short: "Manage multiple volumes.",
	}

	volumesCmd.AddCommand(
		NewListVolumesCmd(cmdFactory),
	)

	return volumesCmd
}
