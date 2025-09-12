package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

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
	}

	rootCmd.AddCommand(
		newServeCmd(),
	)

	return rootCmd
}
