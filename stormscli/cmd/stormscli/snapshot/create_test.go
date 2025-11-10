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

func Test_NewCreateSnapshotCmd(t *testing.T) {
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
				"--src-vol-id",
				"27b5909c-4044-41d9-a601-667c68bf0858",
			},
			expectErr: false,
		},
		{
			name: "missing id",
			args: []string{
				"--src-vol-id",
				"27b5909c-4044-41d9-a601-667c68bf0858",
			},
			expectErr: true,
		},
		{
			name: "missing src vol id",
			args: []string{
				"id",
				"4141c8b6-9a6d-47ff-9bba-e047d131c9a6",
			},
			expectErr: true,
		},
	}

	mockClientProvider := func(context.Context) (storms.StorageManagementServiceClient, io.Closer, error) {
		return &testutil.MockStorMSClient{
			MockCreateSnapshot: func(ctx context.Context, in *storms.CreateSnapshotRequest, opts ...grpc.CallOption) (*storms.CreateSnapshotResponse, error) {
				return nil, nil
			},
		}, &testutil.MockCloser{}, nil
	}

	mockCmdFactory := &utils.CmdFactory{
		StorMSClientProvider: mockClientProvider,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCreateSnapshotCmd(mockCmdFactory)
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
