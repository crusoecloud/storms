package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"gitlab.com/crusoeenergy/island/storage/storms/application"
	"gitlab.com/crusoeenergy/island/storage/storms/configs"
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the Storage Management Service application",
		RunE:  serveCmdFunc,
	}

	configs.AddFlags(cmd)

	return cmd
}

func serveCmdFunc(cmd *cobra.Command, args []string) error {
	if err := configs.ApplyConfig(cmd, args); err != nil {
		return fmt.Errorf("failed to apply configs: %w", err)
	}

	appConfig := configs.Get()
	log.Info().Msgf("StorMS configuration: %#v\n", appConfig)

	app, err := application.NewApp(&appConfig)
	if err != nil {
		return fmt.Errorf("failed to create new application: %w", err)
	}

	if err = app.Start(cmd.Context()); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return nil
}
