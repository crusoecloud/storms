package main

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"gitlab.com/crusoeenergy/island/storage/storms/storms/internal/app"
	appconfigs "gitlab.com/crusoeenergy/island/storage/storms/storms/internal/app/configs"
)

func main() {
	Execute()
}

func Execute() {
	cmd := NewRootCmd()
	if err := cmd.Execute(); err != nil {
		log.Err(fmt.Errorf("failed to execute root command: %w", err))
	}
}

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "storms",
		Short: "Storage Management Service",
		Args:  cobra.NoArgs,
		RunE:  serveCmdFunc,
	}

	appconfigs.AddFlags(rootCmd)

	return rootCmd
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
