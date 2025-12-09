package snapshot

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

func Test_NewGetSnapshotCmd(t *testing.T) {
	expectedTime := time.Date(2025, 11, 20, 14, 45, 0, 0, time.UTC)

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
			MockGetSnapshot: func(ctx context.Context, in *storms.GetSnapshotRequest, opts ...grpc.CallOption) (*storms.GetSnapshotResponse, error) {
				return &storms.GetSnapshotResponse{
					Snapshot: &storms.Snapshot{
						Uuid:             in.Uuid,
						Size:             137438953472,
						SectorSize:       512,
						IsAvailable:      true,
						SourceVolumeUuid: "1f2ea4fe-d8dd-4469-972b-81d166fd2084",
						CreatedAt:        timestamppb.New(expectedTime),
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
			cmd := NewGetSnapshotCmd(mockCmdFactory)
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

			resp, err := client.GetSnapshot(context.Background(), &storms.GetSnapshotRequest{Uuid: "4141c8b6-9a6d-47ff-9bba-e047d131c9a6"})
			require.NoError(t, err)
			require.NotNil(t, resp.Snapshot.CreatedAt)
			require.Equal(t, expectedTime, resp.Snapshot.CreatedAt.AsTime())
		})
	}
}
