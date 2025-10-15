package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"gitlab.com/crusoeenergy/island/storage/storms/cmd/app"
	"gitlab.com/crusoeenergy/island/storage/storms/cmd/snapshot"
	"gitlab.com/crusoeenergy/island/storage/storms/cmd/snapshots"
	"gitlab.com/crusoeenergy/island/storage/storms/cmd/sync"
	"gitlab.com/crusoeenergy/island/storage/storms/cmd/utils"
	"gitlab.com/crusoeenergy/island/storage/storms/cmd/volume"
	"gitlab.com/crusoeenergy/island/storage/storms/cmd/volumes"
)

func Execute() {
	cmd := NewRootCmd()
	if err := cmd.Execute(); err != nil {
		log.Err(fmt.Errorf("failed to execute root command: %w", err))
	}
}

func NewRootCmd() *cobra.Command {
	cmdFactory := utils.NewCmdFactory(
		utils.DefaultAdminProvider,
		utils.DefaultStorMSProvider,
	)

	rootCmd := &cobra.Command{
		Use:   "storms",
		Short: "Storage Management Service",
		Args:  cobra.NoArgs,
	}

	rootCmd.AddCommand(
		app.NewAppCmd(cmdFactory),
		volume.NewVolumeCmd(cmdFactory),
		volumes.NewVolumesCmd(cmdFactory),
		snapshot.NewSnapshotCmd(cmdFactory),
		snapshots.NewSnapshotsCmd(cmdFactory),
		sync.NewSyncCmd(cmdFactory),
	)

	return rootCmd
}
