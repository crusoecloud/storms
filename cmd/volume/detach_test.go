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

func Test_NewDetachVolumeCmd(t *testing.T) {
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
				"--acl",
				"19fe65b9-db48-4bd7-8d39-dc1a0b008bbd",
			},
			expectErr: false,
		},
		{
			name: "missing id",
			args: []string{
				"--acl",
				"19fe65b9-db48-4bd7-8d39-dc1a0b008bbd",
			},
			expectErr: true,
		},
		{
			name: "missing acl",
			args: []string{
				"--id",
				"493ddf53-1794-450e-a576-09cc11399633",
			},
			expectErr: true,
		},
		{
			name: "valid; len(acl) > 1",
			args: []string{
				"--id",
				"493ddf53-1794-450e-a576-09cc11399633",
				"--acl",
				"19fe65b9-db48-4bd7-8d39-dc1a0b008bbd,6352656f-d69c-4f4f-8a6c-578fc7e30102",
			},
			expectErr: false,
		},
	}

	mockClientProvider := func(context.Context) (storms.StorageManagementServiceClient, io.Closer, error) {
		return &testutil.MockStorMSClient{
			MockDetachVolume: func(ctx context.Context, in *storms.DetachVolumeRequest, opts ...grpc.CallOption) (*storms.DetachVolumeResponse, error) {
				return &storms.DetachVolumeResponse{}, nil
			},
		}, &testutil.MockCloser{}, nil
	}

	mockCmdFactory := &utils.CmdFactory{
		StorMSClientProvider: mockClientProvider,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewDetachVolumeCmd(mockCmdFactory)
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
