package sync

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/utils"
)

var (
	errUsage = errors.New("usage error")
)

const (
	allFlag        = "all"
	volumeIDFlag   = "volume-id"
	snapshotIDFlag = "snapshot-id"
	clusterIDFlag  = "cluster-id"

	syncCmdExMsg = `
sync --all
sync --volume-id <volume-id> --cluster-id <cluster-id>
sync --snapshot-id <snapshot-id> --cluster-id <cluster-id>
	`
)

func NewSyncCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sync",
		Short:   "Sync all resources or a single snapshot or volume.",
		Example: syncCmdExMsg,
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateFlags(cmd); err != nil {
				return fmt.Errorf("failed to validate flags: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, conn, err := cmdFactory.StorMSClientProvider(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to create StorMS client: %w", err)
			}
			defer conn.Close()

			if utils.MustGetBoolFlag(cmd, allFlag) {
				err1 := syncAllResources(cmd, client)
				if err1 != nil {
					return fmt.Errorf("failed command: %w", err1)
				}

				return nil
			}

			err = syncResource(cmd, client)
			if err != nil {
				return fmt.Errorf("failed command: %w", err)
			}

			return nil
		},
	}

	utils.NewFlagBuilder(cmd).
		String(clusterIDFlag, "", "cluster id of resource", false).
		String(volumeIDFlag, "", "uuid of volume", false).
		String(snapshotIDFlag, "", "uuid of snapshot", false).
		Bool(allFlag, "", "provide this flag to sync all reources", false)

	return cmd
}

func syncAllResources(cmd *cobra.Command, client storms.StorageManagementServiceClient) error {
	_, err := client.SyncAllResources(cmd.Context(), &storms.SyncAllResourcesRequest{})
	if err != nil {
		return fmt.Errorf("failed to sync all resources: %w", err)
	}

	return nil
}

func syncResource(cmd *cobra.Command, client storms.StorageManagementServiceClient) error {
	var resourceType storms.ResourceType
	resourceID := ""

	volID := utils.MustGetStringFlag(cmd, volumeIDFlag)
	snapshotID := utils.MustGetStringFlag(cmd, snapshotIDFlag)
	clusterID := utils.MustGetStringFlag(cmd, clusterIDFlag)

	if volID != "" {
		resourceID = volID
	} else if snapshotID != "" {
		resourceID = snapshotID
	}
	_, err := client.SyncResource(
		cmd.Context(),
		&storms.SyncResourceRequest{
			ResourceType: resourceType,
			Uuid:         resourceID,
			ClusterUuid:  clusterID,
		})
	if err != nil {
		return fmt.Errorf("failed to sync resource: %w", err)
	}

	return nil
}

func validateFlags(cmd *cobra.Command) error {
	all := utils.MustGetBoolFlag(cmd, allFlag)
	volID := utils.MustGetStringFlag(cmd, volumeIDFlag)
	snapshotID := utils.MustGetStringFlag(cmd, snapshotIDFlag)
	clusterID := utils.MustGetStringFlag(cmd, clusterIDFlag)

	setFlags := 0
	if volID != "" {
		setFlags++
	}
	if snapshotID != "" {
		setFlags++
	}
	if setFlags != 0 && clusterID == "" {
		return fmt.Errorf("when specifying volume or snapshot ID, cluster ID must be provided: %w", errUsage)
	}
	if all {
		setFlags++
	}

	if setFlags == 0 {
		return fmt.Errorf("must specify exactly one of --volume-id, --snapshot-id, or --all: %w", errUsage)
	}
	if setFlags > 1 {
		return fmt.Errorf("only one of --volume-id, --snapshot-id, or --all can be specified: %w", errUsage)
	}

	return nil
}
