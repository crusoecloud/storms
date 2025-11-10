package app

import (
	"github.com/spf13/cobra"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/utils"
)

func NewAppCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	appCmd := &cobra.Command{
		Use:   "app",
		Short: "Manage StorMS application.",
	}

	appCmd.AddCommand(
		NewReloadCmd(cmdFactory),
		NewShowCmd(cmdFactory),
	)

	return appCmd
}
