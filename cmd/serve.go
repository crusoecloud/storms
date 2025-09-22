package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"gitlab.com/crusoeenergy/island/storage/storms/internal/app"
	appconfigs "gitlab.com/crusoeenergy/island/storage/storms/internal/app/configs"
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the Storage Management Service application",
		RunE:  serveCmdFunc,
	}

	appconfigs.AddFlags(cmd)

	return cmd
}

func serveCmdFunc(cmd *cobra.Command, args []string) error {
	if err := appconfigs.ApplyConfig(cmd, args); err != nil {
		return fmt.Errorf("failed to apply appconfigs: %w", err)
	}

	appConfig := appconfigs.Get()
	log.Info().Msgf("StorMS configuration: %#v\n", appConfig)

	a, err := app.NewApp(&appConfig)
	if err != nil {
		return fmt.Errorf("failed to create new application: %w", err)
	}

	if err = a.Start(cmd.Context()); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return nil
}
