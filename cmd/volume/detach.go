//nolint:dupl // attach and detach logic is intended to be similar
package volume

import (
	"fmt"

	"github.com/spf13/cobra"
	storms "gitlab.com/crusoeenergy/island/storage/storms/api/gen/go/storms/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/cmd/utils"
)

func NewDetachVolumeCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "detach",
		Short: "Detach a volume.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, conn, err := cmdFactory.StorMSClientProvider(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to create StorMS client: %w", err)
			}
			defer conn.Close()

			err = detachVolume(cmd, client)
			if err != nil {
				return fmt.Errorf("failed command: %w", err)
			}

			return nil
		},
	}

	utils.NewFlagBuilder(cmd).
		String(idFlag, "", "id of volume", true).
		StringCSV(aclFlag, "", "comma-separated list uuids to remove from volume ACL", true)

	return cmd
}

func detachVolume(cmd *cobra.Command, client storms.StorageManagementServiceClient) error {
	id := utils.MustGetStringFlag(cmd, idFlag)
	acl := utils.MustGetStringCSVFlag(cmd, aclFlag)

	_, err := client.DetachVolume(cmd.Context(), &storms.DetachVolumeRequest{
		Uuid: id,
		Acls: acl,
	})
	if err != nil {
		return fmt.Errorf("failed to attach volume: %w", err)
	}

	cmd.Printf("Removed from ACL of volume %s: %v\n", id, acl)

	return nil
}
