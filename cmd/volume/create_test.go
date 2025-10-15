package volume

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

func Test_NewCreateVolumeCmd(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name: "valid; empty volume",
			args: []string{
				"--id",
				"e9131937-f4ff-4188-bcfa-16a03bf6c62e",
				"--size",
				"128GiB",
				"--sector-size",
				"512",
			},
			expectErr: false,
		},
		{
			name: "invalid; bad size format",
			args: []string{
				"--id",
				"e9131937-f4ff-4188-bcfa-16a03bf6c62e",
				"--size",
				"137438953472",
				"--sector-size",
				"512",
			},
			expectErr: true,
		},
		{
			name: "invalid; unsupported sector size",
			args: []string{
				"--id",
				"e9131937-f4ff-4188-bcfa-16a03bf6c62e",
				"--size",
				"137438953472",
				"--sector-size",
				"8192",
			},
			expectErr: true,
		},
		{
			name: "valid; from snapshot",
			args: []string{
				"--id",
				"e9131937-f4ff-4188-bcfa-16a03bf6c62e",
				"--src-snapshot-id",
				"6bdfb9c2-7da9-4a58-bc63-eba2a4f12631",
			},
			expectErr: false,
		},
		{
			name: "invalid; usually mutually exclusive arguments",
			args: []string{
				"--id",
				"e9131937-f4ff-4188-bcfa-16a03bf6c62e",
				"--size",
				"128GiB",
				"--sector-size",
				"512",
				"--src-snapshot-id",
				"6bdfb9c2-7da9-4a58-bc63-eba2a4f12631",
			},
			expectErr: true,
		},
	}

	mockClientProvider := func(context.Context) (storms.StorageManagementServiceClient, io.Closer, error) {
		return &testutil.MockStorMSClient{
			MockCreateVolume: func(ctx context.Context, in *storms.CreateVolumeRequest, opts ...grpc.CallOption) (*storms.CreateVolumeResponse, error) {
				return &storms.CreateVolumeResponse{}, nil
			},
		}, &testutil.MockCloser{}, nil
	}

	mockCmdFactory := &utils.CmdFactory{
		StorMSClientProvider: mockClientProvider,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCreateVolumeCmd(mockCmdFactory)
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
