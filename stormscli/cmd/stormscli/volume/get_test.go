package volume

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
	testutil "gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/testutil"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/utils"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_NewGetVolume(t *testing.T) {
	expectedTime := time.Date(2025, 11, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name: "valid",
			args: []string{
				"--id",
				"4141c8b6-9a6d-47ff-9bba-e047d131c9a6",
			},
			expectErr: false,
		},
		{
			name:      "invalid; missing id",
			args:      []string{},
			expectErr: true,
		},
	}

	mockClientProvider := func(context.Context) (storms.StorageManagementServiceClient, io.Closer, error) {
		return &testutil.MockStorMSClient{
			MockGetVolume: func(ctx context.Context, in *storms.GetVolumeRequest, opts ...grpc.CallOption) (*storms.GetVolumeResponse, error) {
				return &storms.GetVolumeResponse{
					Volume: &storms.Volume{
						Uuid:               in.Uuid,
						Size:               137438953472,
						SectorSize:         512,
						IsAvailable:        true,
						SourceSnapshotUuid: "1f2ea4fe-d8dd-4469-972b-81d166fd2084",
						CreatedAt:          timestamppb.New(expectedTime),
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
			cmd := NewGetVolumeCmd(mockCmdFactory)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.expectErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			client, closer, err := mockClientProvider(context.Background())
			require.NoError(t, err)
			defer closer.Close()

			resp, err := client.GetVolume(context.Background(), &storms.GetVolumeRequest{Uuid: "4141c8b6-9a6d-47ff-9bba-e047d131c9a6"})
			require.NoError(t, err)
			require.NotNil(t, resp.Volume.CreatedAt)
			require.Equal(t, expectedTime, resp.Volume.CreatedAt.AsTime())
		})
	}
}
