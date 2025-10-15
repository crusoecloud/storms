package volume

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	storms "gitlab.com/crusoeenergy/island/storage/storms/api/gen/go/storms/v1"
	testutil "gitlab.com/crusoeenergy/island/storage/storms/cmd/testutil"
	"gitlab.com/crusoeenergy/island/storage/storms/cmd/utils"
)

func Test_NewResizeVolumeCmd(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name: "valid",
			args: []string{
				"--id",
				"493ddf53-1794-450e-a576-09cc11399633",
				"--size",
				"256GiB",
			},
			expectErr: false,
		},
		{
			name: "missing id",
			args: []string{
				"--size",
				"256GiB",
			},
			expectErr: true,
		},
		{
			name: "invalid size format",
			args: []string{
				"--id",
				"493ddf53-1794-450e-a576-09cc11399633",
				"--size",
				"239487",
			},
			expectErr: true,
		},
	}

	mockClientProvider := func(context.Context) (storms.StorageManagementServiceClient, io.Closer, error) {
		return &testutil.MockStorMSClient{
			MockResizeVolume: func(ctx context.Context, in *storms.ResizeVolumeRequest, opts ...grpc.CallOption) (*storms.ResizeVolumeResponse, error) {
				return &storms.ResizeVolumeResponse{}, nil
			},
		}, &testutil.MockCloser{}, nil
	}

	mockCmdFactory := &utils.CmdFactory{
		StorMSClientProvider: mockClientProvider,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewResizeVolumeCmd(mockCmdFactory)
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
