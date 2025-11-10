package snapshot

import (
	"github.com/spf13/cobra"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/utils"
)

func NewSnapshotCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	snapshotCmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Manage a single snapshot.",
	}

	snapshotCmd.AddCommand(
		NewCreateSnapshotCmd(cmdFactory),
		NewDeleteSnapshotCmd(cmdFactory),
		NewGetSnapshotCmd(cmdFactory),
	)

	return snapshotCmd
}
