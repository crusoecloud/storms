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

func Test_NewDeleteVolumeCmd(t *testing.T) {
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
			MockDeleteVolume: func(ctx context.Context, in *storms.DeleteVolumeRequest, opts ...grpc.CallOption) (*storms.DeleteVolumeResponse, error) {
				return &storms.DeleteVolumeResponse{}, nil
			},
		}, &testutil.MockCloser{}, nil
	}

	mockCmdFactory := &utils.CmdFactory{
		StorMSClientProvider: mockClientProvider,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewDeleteVolumeCmd(mockCmdFactory)
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
