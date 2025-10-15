package app

import (
	"github.com/spf13/cobra"
	"gitlab.com/crusoeenergy/island/storage/storms/cmd/utils"
)

func NewAppCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	appCmd := &cobra.Command{
		Use:   "app",
		Short: "Manage StorMS application.",
	}

	appCmd.AddCommand(
		NewStartCmd(cmdFactory),
		NewReloadCmd(cmdFactory),
		NewShowCmd(cmdFactory),
	)

	return appCmd
}
