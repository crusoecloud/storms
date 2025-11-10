package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/stormscli/app"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/stormscli/snapshot"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/stormscli/snapshots"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/stormscli/sync"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/stormscli/volume"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/stormscli/volumes"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/utils"
)

var errUnexpectedHostnameFormat = errors.New("unexpected hostname/ip format")

func main() {
	execute()
}

func execute() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		log.Err(fmt.Errorf("failed to execute root command: %w", err))
	}
}

func newRootCmd() *cobra.Command {
	cmdFactory := utils.NewCmdFactory()

	var targetAddr string
	rootCmd := &cobra.Command{
		Use:   "stormscli",
		Short: "Storage Management Service CLI",
		Args:  cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			targetAddr = cmd.Flags().Lookup("target-addr").Value.String()
			if err := validateTargetAddr(targetAddr); err != nil {
				return fmt.Errorf("failed to validate target addr: %w", err)
			}
			cmdFactory.TargetAddr = targetAddr

			return nil
		},
	}

	rootCmd.AddCommand(
		app.NewAppCmd(cmdFactory),
		volume.NewVolumeCmd(cmdFactory),
		volumes.NewVolumesCmd(cmdFactory),
		snapshot.NewSnapshotCmd(cmdFactory),
		snapshots.NewSnapshotsCmd(cmdFactory),
		sync.NewSyncCmd(cmdFactory),
	)

	rootCmd.PersistentFlags().StringP("target-addr", "", "", "target address of StorMS service")

	return rootCmd
}

func validateTargetAddr(addr string) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid address format: expected <ip>:<port>, got %q: %w", addr, err)
	}

	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port %q: must be an integer: %w", port, err)
	}

	if p < 1 || p > 65535 {
		return fmt.Errorf("port %d out of valid range (1-65535)", p)
	}

	if net.ParseIP(host) == nil && host != "localhost" {
		return errUnexpectedHostnameFormat
	}

	return nil
}
