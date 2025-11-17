package app

import (
	"fmt"

	"github.com/spf13/cobra"
	admin "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/admin/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/utils"
)

func NewReloadCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reload",
		Short: "Triggers a zero-downtime cluster configuration reload.",
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, conn, err := cmdFactory.AdminClientProvider(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to create admin client: %w", err)
			}
			defer conn.Close()

			err = reloadCmdFn(cmd, client)
			if err != nil {
				return fmt.Errorf("failed command: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func reloadCmdFn(cmd *cobra.Command, client admin.AdminServiceClient) error {
	_, err := client.ReloadConfig(cmd.Context(), &admin.ReloadConfigRequest{})
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	return nil
}
