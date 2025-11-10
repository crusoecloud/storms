package snapshot

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
	testutil "gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/testutil"
	"gitlab.com/crusoeenergy/island/storage/storms/stormscli/cmd/utils"
	"google.golang.org/grpc"
)

func Test_NewGetSnapshotCmd(t *testing.T) {
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
		})
	}
}
