package app

import (
	"fmt"

	"github.com/spf13/cobra"
	admin "gitlab.com/crusoeenergy/island/storage/storms/api/gen/go/admin/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/cmd/utils"
	appconfigs "gitlab.com/crusoeenergy/island/storage/storms/internal/app/configs"
)

func NewShowCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show information about the StorMS application",
		// TODO - consider the case where use reloads after modifying the app config file
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, conn, err := cmdFactory.AdminClientProvider(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to create admin client: %w", err)
			}
			defer conn.Close()

			err = showCmdFn(cmd, client)
			if err != nil {
				return fmt.Errorf("failed command: %w", err)
			}

			return nil
		},
	}

	appconfigs.AddFlags(cmd)

	return cmd
}

func showCmdFn(cmd *cobra.Command, client admin.AdminServiceClient) error {
	resp, err := client.ShowClusters(cmd.Context(), &admin.ShowClustersRequest{})
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	if err := utils.RenderClusters(resp.Clusters); err != nil {
		return fmt.Errorf("failed to render clusters: %w", err)
	}

	return nil
}
