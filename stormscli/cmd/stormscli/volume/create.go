package volume

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/utils"
)

const (
	createVolCmdLongMsg = `
Create a volume in one of two mutually exclusive ways:

1. Create empty volume:
Specify a new size and sector size for an empty volume.
This mode requires the --id, --size, and --sector-size flags.

2. From a Snapshot:
Create a volume as a copy of an existing snapshot. The size and
sector size will be inherited from the snapshot.
This mode requires the --id and --snapshot-id flags.
		`
	createVolCmdShortMsg   = "Create a volume. Can either be a new empty volume or from a snapshot"
	createVolCmdExampleMsg = `
# Create a 10GiB volume new empty volume
create --id <volume-uuid> --size 10GiB --sector-size 4096

# Create a volume from an existing snapshot
create --id <volume-uuid> --snapshot-id <snapshot-uuid>
		`
)

var (
	errCmdUsage = errors.New("cmd usage error")
)

func NewCreateVolumeCmd(cmdFactory *utils.CmdFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create ",
		Short:   createVolCmdShortMsg,
		Long:    createVolCmdLongMsg,
		Example: createVolCmdExampleMsg,

		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			return validateVolSourceMethod(cmd)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, conn, err := cmdFactory.StorMSClientProvider(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to create StorMS client: %w", err)
			}
			defer conn.Close()

			err = createVolume(cmd, client)
			if err != nil {
				return fmt.Errorf("failed command: %w", err)
			}

			return nil
		},
	}

	utils.NewFlagBuilder(cmd).
		String(idFlag, "", "id of volume", true).
		String(sizeFlag, "", "size of volume in the format 'X[GiB|TiB]' where X is a non-zero integer", false).
		Uint(sectorSizeFlag, "", "sector size of volumes: either '512' or '4096'", false).
		String(srcSnapshotIDFlag, "", "id of the source snapshot", false)

	return cmd
}

func createVolume(cmd *cobra.Command, client storms.StorageManagementServiceClient) error {
	id := utils.MustGetStringFlag(cmd, idFlag)
	srcSnapshotID := utils.MustGetStringFlag(cmd, srcSnapshotIDFlag)

	var req *storms.CreateVolumeRequest
	var err error

	// Populate the 'oneof source' field based on the provided flags.
	if srcSnapshotID != "" {
		req = createCreateVolFromSnapshotRequest(id, srcSnapshotID)
	} else {
		sizeStr := utils.MustGetStringFlag(cmd, sizeFlag)
		sectorSize := utils.MustGetUintFlag(cmd, sectorSizeFlag)
		req, err = createCreateEmptyVolRequest(id, sizeStr, sectorSize)
		if err != nil {
			return fmt.Errorf("failed to create request to create new empty volume: %w", err)
		}
	}

	// Send the fully constructed request to the client.
	_, err = client.CreateVolume(cmd.Context(), req)
	if err != nil {
		return fmt.Errorf("failed to create volume: %w", err)
	}

	cmd.Printf("Created volume: %s\n", id)

	return nil
}

// Validate that either (size, sector size) XOR (source snapshot ID) if provided.
func validateVolSourceMethod(cmd *cobra.Command) error {
	srcSnapshotID := utils.MustGetStringFlag(cmd, srcSnapshotIDFlag)
	sizeStr := utils.MustGetStringFlag(cmd, sizeFlag)
	sectorSize := utils.MustGetUintFlag(cmd, sectorSizeFlag)

	fromSnapshot := srcSnapshotID != ""
	fromNew := sizeStr != "" || sectorSize != 0

	if fromSnapshot && fromNew {
		return fmt.Errorf("use either --snapshot-id or (--size and --sector-size), but not both: %w", errCmdUsage)
	}
	if !fromSnapshot && !fromNew {
		return fmt.Errorf("either --snapshot-id or both --size and --sector-size must be provided: %w", errCmdUsage)
	}
	if fromNew && (sizeStr == "" || sectorSize == 0) {
		return fmt.Errorf("when creating a new volume, both --size and --sector-size are required: %w", errCmdUsage)
	}

	return nil
}

func createCreateVolFromSnapshotRequest(id, srcSnapshotID string) *storms.CreateVolumeRequest {
	req := &storms.CreateVolumeRequest{
		Uuid: id,
		Source: &storms.CreateVolumeRequest_FromSnapshot{
			FromSnapshot: &storms.SnapshotSourceVolumeSpec{
				SnapshotUuid: srcSnapshotID,
			},
		},
	}

	return req
}

func createCreateEmptyVolRequest(id, size string, sectorSize uint) (*storms.CreateVolumeRequest, error) {
	sizeBytes, err := utils.ParseSizeString(size)
	if err != nil {
		return nil, fmt.Errorf("failed to parse size: %w", err)
	}

	sectorSizeEnum, err := utils.ParseSectorSizeUint(sectorSize)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sector size: %w", err)
	}

	req := &storms.CreateVolumeRequest{
		Uuid: id,
		Source: &storms.CreateVolumeRequest_FromNew{
			FromNew: &storms.NewVolumeSpec{
				Size:       sizeBytes,
				SectorSize: sectorSizeEnum,
			},
		},
	}

	return req, nil
}
