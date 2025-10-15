package volumes

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	storms "gitlab.com/crusoeenergy/island/storage/storms/api/gen/go/storms/v1"
	testutil "gitlab.com/crusoeenergy/island/storage/storms/cmd/testutil"
	"gitlab.com/crusoeenergy/island/storage/storms/cmd/utils"
	"google.golang.org/grpc"
)

func Test_NewListVolumesCmd(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name: "valid; 1 id",
			args: []string{
				"--ids",
				"4141c8b6-9a6d-47ff-9bba-e047d131c9a6",
			},
			expectErr: false,
		},
		{
			name: "valid; >1 ids",
			args: []string{
				"--ids",
				"8c2a02b0-5e29-4f1f-a11f-cbd07ed17089,710fb780-db3c-458d-9903-479482afa7e0",
			},
			expectErr: false,
		},
		{
			name:      "valid; 0 ids",
			args:      []string{},
			expectErr: false,
		},
	}

	mockClientProvider := func(context.Context) (storms.StorageManagementServiceClient, io.Closer, error) {
		return &testutil.MockStorMSClient{
			MockGetVolume: func(ctx context.Context, in *storms.GetVolumeRequest, opts ...grpc.CallOption) (*storms.GetVolumeResponse, error) {
				return &storms.GetVolumeResponse{
					Volume: &storms.Volume{
						Uuid:               "0fb8966f-9224-40dd-9b9c-febbf21e163d",
						Size:               137438953472,
						SectorSize:         512,
						IsAvailable:        true,
						SourceSnapshotUuid: "1f2ea4fe-d8dd-4469-972b-81d166fd2084",
					},
				}, nil
			},
			MockGetVolumes: func(ctx context.Context, in *storms.GetVolumesRequest, opts ...grpc.CallOption) (*storms.GetVolumesResponse, error) {
				return &storms.GetVolumesResponse{
					Volumes: []*storms.Volume{
						{
							Uuid:               "0fb8966f-9224-40dd-9b9c-febbf21e163d",
							Size:               137438953472,
							SectorSize:         512,
							IsAvailable:        true,
							SourceSnapshotUuid: "1f2ea4fe-d8dd-4469-972b-81d166fd2084",
						},
						{
							Uuid:               "798123a5-dc8c-4837-b744-54e4e63ebe56",
							Size:               137438953472,
							SectorSize:         512,
							IsAvailable:        true,
							SourceSnapshotUuid: "",
						},
					},
				}, nil
			},
		}, &testutil.MockCloser{}, nil
	}

	mockCmdFactory := &utils.CmdFactory{
		StorMSClientProvider: mockClientProvider,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewListVolumesCmd(mockCmdFactory)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.expectErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
		})
	}
}
