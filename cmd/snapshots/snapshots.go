package snapshots

import (
	"github.com/spf13/cobra"
	"gitlab.com/crusoeenergy/island/storage/storms/cmd/utils"
)

func NewSnapshotsCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	snapshotsCmd := &cobra.Command{
		Use:   "snapshots",
		Short: "Manage multiple snapshots.",
	}

	snapshotsCmd.AddCommand(
		NewListSnapshotsCmd(cmdFactory),
	)

	return snapshotsCmd
}
