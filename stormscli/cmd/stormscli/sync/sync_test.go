package sync

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

func Test_NewSyncCmd(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name: "valid sync all",
			args: []string{
				"--all",
			},
			expectErr: false,
		},
		{
			name: "valid sync a volume",
			args: []string{
				"--volume-id",
				"d8ba36d6-f949-45b2-babe-dc65c26a9a13",
				"--cluster-id",
				"6a898a70-8675-4c48-acd9-54cb167c3a3e",
			},
			expectErr: false,
		},
		{
			name: "valid sync a snapshot",
			args: []string{
				"--snapshot-id",
				"959b1026-f4a1-4f93-b62f-92d9870fe2f7",
				"--cluster-id",
				"6a898a70-8675-4c48-acd9-54cb167c3a3e",
			},
			expectErr: false,
		},
		{
			name: "invalid; provided cluster id with no resource",
			args: []string{
				"--cluster-id",
				"6a898a70-8675-4c48-acd9-54cb167c3a3e",
			},
			expectErr: true,
		},
		{
			name: "invalid; provided mutually exclusive arguments",
			args: []string{
				"--all",
				"--snapshot-id",
				"959b1026-f4a1-4f93-b62f-92d9870fe2f7",
				"--cluster-id",
				"6a898a70-8675-4c48-acd9-54cb167c3a3e",
			},
			expectErr: true,
		},
	}

	mockClientProvider := func(context.Context) (storms.StorageManagementServiceClient, io.Closer, error) {
		return &testutil.MockStorMSClient{
			MockSyncResource: func(ctx context.Context, in *storms.SyncResourceRequest, opts ...grpc.CallOption) (*storms.SyncResourceResponse, error) {
				return &storms.SyncResourceResponse{}, nil
			},
			MockSyncAllResources: func(ctx context.Context, in *storms.SyncAllResourcesRequest, opts ...grpc.CallOption) (*storms.SyncAllResourcesResponse, error) {
				return &storms.SyncAllResourcesResponse{}, nil
			},
		}, &testutil.MockCloser{}, nil
	}

	mockCmdFactory := &utils.CmdFactory{
		StorMSClientProvider: mockClientProvider,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewSyncCmd(mockCmdFactory)
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
