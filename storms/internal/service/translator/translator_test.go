package translator

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gitlab.com/crusoeenergy/island/storage/storms/client/models"
	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
)

const (
	defaultOSDiskSizeBytes = 137438953472
	sectorSize512          = 512
	sectorSize4096         = 4096
)

type mockClient struct {
	mockGetVolume      func(ctx context.Context, req *models.GetVolumeRequest) (*models.GetVolumeResponse, error)
	mockGetVolumes     func(ctx context.Context, req *models.GetVolumesRequest) (*models.GetVolumesResponse, error)
	mockCreateVolume   func(ctx context.Context, req *models.CreateVolumeRequest) (*models.CreateVolumeResponse, error)
	mockResizeVolume   func(ctx context.Context, req *models.ResizeVolumeRequest) (*models.ResizeVolumeResponse, error)
	mockDeleteVolume   func(ctx context.Context, req *models.DeleteVolumeRequest) (*models.DeleteVolumeResponse, error)
	mockAttachVolume   func(ctx context.Context, req *models.AttachVolumeRequest) (*models.AttachVolumeResponse, error)
	mockDetachVolume   func(ctx context.Context, req *models.DetachVolumeRequest) (*models.DetachVolumeResponse, error)
	mockGetSnapshot    func(ctx context.Context, req *models.GetSnapshotRequest) (*models.GetSnapshotResponse, error)
	mockGetSnapshots   func(ctx context.Context, req *models.GetSnapshotsRequest) (*models.GetSnapshotsResponse, error)
	mockCreateSnapshot func(ctx context.Context, req *models.CreateSnapshotRequest) (*models.CreateSnapshotResponse, error)
	mockDeleteSnapshot func(ctx context.Context, req *models.DeleteSnapshotRequest) (*models.DeleteSnapshotResponse, error)
}

func (m *mockClient) GetVolume(ctx context.Context, req *models.GetVolumeRequest) (*models.GetVolumeResponse, error) {
	return m.mockGetVolume(ctx, req)
}

func (m *mockClient) GetVolumes(ctx context.Context, req *models.GetVolumesRequest) (*models.GetVolumesResponse, error) {
	return m.mockGetVolumes(ctx, req)
}

func (m *mockClient) CreateVolume(ctx context.Context, req *models.CreateVolumeRequest) (*models.CreateVolumeResponse, error) {
	return m.mockCreateVolume(ctx, req)
}

func (m *mockClient) ResizeVolume(ctx context.Context, req *models.ResizeVolumeRequest) (*models.ResizeVolumeResponse, error) {
	return m.mockResizeVolume(ctx, req)
}

func (m *mockClient) DeleteVolume(ctx context.Context, req *models.DeleteVolumeRequest) (*models.DeleteVolumeResponse, error) {
	return m.mockDeleteVolume(ctx, req)
}

func (m *mockClient) AttachVolume(ctx context.Context, req *models.AttachVolumeRequest) (*models.AttachVolumeResponse, error) {
	return m.mockAttachVolume(ctx, req)
}

func (m *mockClient) DetachVolume(ctx context.Context, req *models.DetachVolumeRequest) (*models.DetachVolumeResponse, error) {
	return m.mockDetachVolume(ctx, req)
}

func (m *mockClient) GetSnapshot(ctx context.Context, req *models.GetSnapshotRequest) (*models.GetSnapshotResponse, error) {
	return m.mockGetSnapshot(ctx, req)
}

func (m *mockClient) GetSnapshots(ctx context.Context, req *models.GetSnapshotsRequest) (*models.GetSnapshotsResponse, error) {
	return m.mockGetSnapshots(ctx, req)
}

func (m *mockClient) CreateSnapshot(ctx context.Context, req *models.CreateSnapshotRequest) (*models.CreateSnapshotResponse, error) {
	return m.mockCreateSnapshot(ctx, req)
}

func (m *mockClient) DeleteSnapshot(ctx context.Context, req *models.DeleteSnapshotRequest) (*models.DeleteSnapshotResponse, error) {
	return m.mockDeleteSnapshot(ctx, req)
}

func Test_AttachVolume(t *testing.T) {
	tests := []struct {
		name      string
		input     *storms.AttachVolumeRequest
		expectErr bool
	}{
		{
			name: "0 acl",
			input: &storms.AttachVolumeRequest{
				Uuid: "e63b6b55-acfe-446b-95c4-f6b37d94b4b4",

				Acl: []string{},
			},
			expectErr: false,
		},
		{
			name: "1 acl",
			input: &storms.AttachVolumeRequest{
				Uuid: "e63b6b55-acfe-446b-95c4-f6b37d94b4b4",
				Acl: []string{
					"e4c2e3f8-a0bd-4e37-b5d0-a107137651fc",
				},
			},
			expectErr: false,
		},
		{
			name: "2 acl",
			input: &storms.AttachVolumeRequest{
				Uuid: "e63b6b55-acfe-446b-95c4-f6b37d94b4b4",
				Acl: []string{
					"e4c2e3f8-a0bd-4e37-b5d0-a107137651fc",
					"1ce99ac6-5f62-4a75-83ff-0bc7f7109d1a",
				},
			},
			expectErr: false,
		},
	}

	ct := NewClientTranslator()
	mc := &mockClient{
		mockAttachVolume: func(ctx context.Context, req *models.AttachVolumeRequest) (*models.AttachVolumeResponse, error) {
			return &models.AttachVolumeResponse{}, nil
		},
	}

	for _, tt := range tests {
		_, err := ct.AttachVolume(context.Background(), mc, tt.input)
		if tt.expectErr {
			require.Error(t, err)

			continue
		}

		require.NoError(t, err)
	}
}

func Test_CreateSnapshot(t *testing.T) {
	tests := []struct {
		name      string
		input     *storms.CreateSnapshotRequest
		expect    *storms.CreateSnapshotResponse
		expectErr bool
	}{
		{
			name: "valid request",
			input: &storms.CreateSnapshotRequest{
				Uuid:          uuid.NewString(),
				SrcVolumeUuid: uuid.NewString(),
			},
		},
	}

	ct := NewClientTranslator()
	mc := &mockClient{
		mockCreateSnapshot: func(ctx context.Context, req *models.CreateSnapshotRequest) (*models.CreateSnapshotResponse, error) {
			return &models.CreateSnapshotResponse{}, nil
		},
	}

	for _, tt := range tests {
		_, err := ct.CreateSnapshot(context.Background(), mc, tt.input)
		if tt.expectErr {
			require.Error(t, err)

			continue
		}

		require.NoError(t, err)
	}
}

func Test_CreateVolume(t *testing.T) {
	tests := []struct {
		name      string
		input     *storms.CreateVolumeRequest
		expect    *storms.CreateVolumeResponse
		expectErr bool
	}{
		{
			name: "valid new volume",
			input: &storms.CreateVolumeRequest{
				Uuid: uuid.NewString(),
				Source: &storms.CreateVolumeRequest_FromNew{
					FromNew: &storms.NewVolumeSpec{
						Size:       defaultOSDiskSizeBytes,
						SectorSize: storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512,
					},
				},
			},
			expect:    &storms.CreateVolumeResponse{},
			expectErr: false,
		},
		{
			name: "valid new volume",
			input: &storms.CreateVolumeRequest{
				Uuid: uuid.NewString(),
				Source: &storms.CreateVolumeRequest_FromSnapshot{
					FromSnapshot: &storms.SnapshotSourceVolumeSpec{
						SnapshotUuid: uuid.NewString(),
					},
				},
			},
			expect:    &storms.CreateVolumeResponse{},
			expectErr: false,
		},
	}

	ct := NewClientTranslator()
	mc := &mockClient{
		mockCreateVolume: func(ctx context.Context, req *models.CreateVolumeRequest) (*models.CreateVolumeResponse, error) {
			return &models.CreateVolumeResponse{}, nil
		},
	}

	for _, tt := range tests {
		res, err := ct.CreateVolume(context.Background(), mc, tt.input)
		if tt.expectErr {
			require.Error(t, err)

			continue
		}
		require.NoError(t, err)
		require.NotNil(t, res)
	}
}

func Test_DeleteSnapshot(t *testing.T) {
	tests := []struct {
		name      string
		input     *storms.DeleteSnapshotRequest
		expect    *storms.DeleteSnapshotResponse
		expectErr bool
	}{
		{
			name:      "valid",
			input:     &storms.DeleteSnapshotRequest{},
			expect:    &storms.DeleteSnapshotResponse{},
			expectErr: false,
		},
	}

	ct := NewClientTranslator()
	mc := &mockClient{
		mockDeleteSnapshot: func(ctx context.Context, req *models.DeleteSnapshotRequest) (*models.DeleteSnapshotResponse, error) {
			return &models.DeleteSnapshotResponse{}, nil
		},
	}

	for _, tt := range tests {
		res, err := ct.DeleteSnapshot(context.Background(), mc, tt.input)
		if tt.expectErr {
			require.Error(t, err)

			continue
		}
		require.NoError(t, err)
		require.NotNil(t, res)

	}
}

func Test_DeleteVolume(t *testing.T) {
	tests := []struct {
		name      string
		input     *storms.DeleteVolumeRequest
		expect    *storms.DeleteVolumeResponse
		expectErr bool
	}{
		{
			name:      "valid",
			input:     &storms.DeleteVolumeRequest{},
			expect:    &storms.DeleteVolumeResponse{},
			expectErr: false,
		},
	}

	ct := NewClientTranslator()
	mc := &mockClient{
		mockDeleteVolume: func(ctx context.Context, req *models.DeleteVolumeRequest) (*models.DeleteVolumeResponse, error) {
			return &models.DeleteVolumeResponse{}, nil
		},
	}

	for _, tt := range tests {
		res, err := ct.DeleteVolume(context.Background(), mc, tt.input)
		if tt.expectErr {
			require.Error(t, err)

			continue
		}
		require.NoError(t, err)
		require.NotNil(t, res)

	}
}

func Test_DetachVolume(t *testing.T) {
	tests := []struct {
		name      string
		input     *storms.DetachVolumeRequest
		expect    *storms.DetachVolumeResponse
		expectErr bool
	}{
		{
			name:      "valid",
			input:     &storms.DetachVolumeRequest{},
			expect:    &storms.DetachVolumeResponse{},
			expectErr: false,
		},
	}

	ct := NewClientTranslator()
	mc := &mockClient{
		mockDetachVolume: func(ctx context.Context, req *models.DetachVolumeRequest) (*models.DetachVolumeResponse, error) {
			return &models.DetachVolumeResponse{}, nil
		},
	}

	for _, tt := range tests {
		res, err := ct.DetachVolume(context.Background(), mc, tt.input)
		if tt.expectErr {
			require.Error(t, err)

			continue
		}
		require.NoError(t, err)
		require.NotNil(t, res)

	}
}

func Test_GetSnapshot(t *testing.T) {
	snapshotUUID := "4533ae7a-ef23-43c5-8e94-68dfcd5bedd6"
	sourceVolumeUUID := "cc23f964-d2d3-40af-a6f5-82b1e0574c93"
	vendorShapshotID := "6f04092b-e9f9-4dc3-9ad1-4b07385826c3"

	tests := []struct {
		name      string
		input     *storms.GetSnapshotRequest
		expect    *storms.GetSnapshotResponse
		expectErr bool
	}{
		{
			name: "valid",
			input: &storms.GetSnapshotRequest{
				Uuid: snapshotUUID,
			},
			expect: &storms.GetSnapshotResponse{
				Snapshot: &storms.Snapshot{
					Uuid:             snapshotUUID,
					VendorSnapshotId: vendorShapshotID,
					Size:             defaultOSDiskSizeBytes,
					SectorSize:       storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512,
					IsAvailable:      true,
					SourceVolumeUuid: sourceVolumeUUID,
				},
			},
			expectErr: false,
		},
	}

	ct := NewClientTranslator()
	mc := &mockClient{
		mockGetSnapshot: func(ctx context.Context, req *models.GetSnapshotRequest) (*models.GetSnapshotResponse, error) {
			return &models.GetSnapshotResponse{
				Snapshot: &models.Snapshot{
					UUID:             req.UUID,
					VendorSnapshotID: vendorShapshotID,
					Size:             defaultOSDiskSizeBytes,
					SectorSize:       sectorSize512,
					SourceVolumeUUID: sourceVolumeUUID,
					IsAvailable:      true,
				},
			}, nil
		},
	}

	for _, tt := range tests {
		res, err := ct.GetSnapshot(context.Background(), mc, tt.input)
		if tt.expectErr {
			require.Error(t, err)

			continue
		}

		require.NoError(t, err)

		require.NotNil(t, res.Snapshot)
		require.True(t, res.Snapshot.IsAvailable)
		require.Equal(t, res.Snapshot.Uuid, snapshotUUID)
		require.Equal(t, tt.expect.Snapshot.VendorSnapshotId, res.Snapshot.VendorSnapshotId)
	}
}

func Test_GetSnapshots(t *testing.T) {
	tests := []struct {
		name           string
		input          *storms.GetSnapshotsRequest
		clientResponse *models.GetSnapshotsResponse
		clientErr      error
		expectErr      bool
	}{
		{
			name:  "1 snapshot",
			input: &storms.GetSnapshotsRequest{},
			clientResponse: &models.GetSnapshotsResponse{
				Snapshots: []*models.Snapshot{
					{
						UUID:             uuid.NewString(),
						Size:             defaultOSDiskSizeBytes,
						SectorSize:       sectorSize512,
						IsAvailable:      true,
						SourceVolumeUUID: uuid.NewString(),
					},
				},
			},
			clientErr: nil,
			expectErr: false,
		},
		{
			name:  "2 snapshots",
			input: &storms.GetSnapshotsRequest{},
			clientResponse: &models.GetSnapshotsResponse{
				Snapshots: []*models.Snapshot{
					{
						UUID:             uuid.NewString(),
						VendorSnapshotID: uuid.NewString(),
						Size:             defaultOSDiskSizeBytes,
						SectorSize:       sectorSize512,
						IsAvailable:      true,
						SourceVolumeUUID: uuid.NewString(),
					},
					{
						UUID:             uuid.NewString(),
						VendorSnapshotID: uuid.NewString(),
						Size:             defaultOSDiskSizeBytes,
						SectorSize:       sectorSize4096,
						IsAvailable:      false,
						SourceVolumeUUID: uuid.NewString(),
					},
				},
			},
			clientErr: nil,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct := NewClientTranslator()
			mc := &mockClient{
				mockGetSnapshots: func(ctx context.Context, req *models.GetSnapshotsRequest) (*models.GetSnapshotsResponse, error) {
					return tt.clientResponse, tt.clientErr
				},
			}
			resp, err := ct.GetSnapshots(context.Background(), mc, tt.input)
			if tt.expectErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			for i := 0; i < len(tt.clientResponse.Snapshots); i++ {
				require.Equal(t, tt.clientResponse.Snapshots[i].UUID, resp.Snapshots[i].Uuid)
				require.Equal(t, tt.clientResponse.Snapshots[i].VendorSnapshotID, resp.Snapshots[i].VendorSnapshotId)
				require.Equal(t, tt.clientResponse.Snapshots[i].Size, resp.Snapshots[i].Size)
				require.Equal(t, tt.clientResponse.Snapshots[i].IsAvailable, resp.Snapshots[i].IsAvailable)
				require.Equal(t, tt.clientResponse.Snapshots[i].SourceVolumeUUID, resp.Snapshots[i].SourceVolumeUuid)
			}
		})

	}
}

func Test_GetVolume(t *testing.T) {

	tests := []struct {
		name      string
		input     *storms.GetVolumeRequest
		expect    *storms.GetVolumeResponse
		expectErr bool
	}{
		{
			name: "valid",
			input: &storms.GetVolumeRequest{
				Uuid: "e7add145-4529-4802-9be7-dc7d33d1f6f0",
			},
			expect: &storms.GetVolumeResponse{
				Volume: &storms.Volume{
					Uuid:           "e7add145-4529-4802-9be7-dc7d33d1f6f0",
					VendorVolumeId: "3c989dfe-7618-4166-8fc2-99e7571d49ed",
					Size:           defaultOSDiskSizeBytes,
					SectorSize:     storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512,
					Acl: []string{
						"2432c08e-b2a4-4a11-aa00-b07cc9789213",
					},
					IsAvailable:        true,
					SourceSnapshotUuid: "c4599262-6e64-4597-8834-d790dab7be35",
				},
			},
			expectErr: false,
		},
	}

	ct := NewClientTranslator()
	mc := &mockClient{
		mockGetVolume: func(ctx context.Context, req *models.GetVolumeRequest) (*models.GetVolumeResponse, error) {
			switch req.UUID {
			case "e7add145-4529-4802-9be7-dc7d33d1f6f0":
				return &models.GetVolumeResponse{
					Volume: &models.Volume{
						UUID:           req.UUID,
						VendorVolumeID: "3c989dfe-7618-4166-8fc2-99e7571d49ed",
						Size:           defaultOSDiskSizeBytes,
						SectorSize:     sectorSize512,
						ACL: []string{
							"2432c08e-b2a4-4a11-aa00-b07cc9789213",
						},
						IsAvailable:        true,
						SourceSnapshotUUID: "c4599262-6e64-4597-8834-d790dab7be35",
					},
				}, nil
			default:
				return &models.GetVolumeResponse{}, fmt.Errorf("fail")
			}
		},
	}

	for _, tt := range tests {
		res, err := ct.GetVolume(context.Background(), mc, tt.input)
		if tt.expectErr {
			require.Error(t, err)

			continue
		}
		require.NoError(t, err)
		require.NotNil(t, res)

		require.Equal(t, tt.expect.Volume, res.Volume)
	}
}

func Test_GetVolumes(t *testing.T) {
	tests := []struct {
		name           string
		input          *storms.GetVolumesRequest
		clientResponse *models.GetVolumesResponse
		clientErr      error
		expectErr      bool
	}{
		{
			name:  "1 volume",
			input: &storms.GetVolumesRequest{},
			clientResponse: &models.GetVolumesResponse{
				Volumes: []*models.Volume{
					{
						UUID:               uuid.NewString(),
						Size:               uint64(defaultOSDiskSizeBytes),
						SectorSize:         sectorSize512,
						ACL:                []string{},
						IsAvailable:        true,
						SourceSnapshotUUID: "",
					},
				},
			},
			clientErr: nil,
			expectErr: false,
		},
		{
			name:  "2 volumes",
			input: &storms.GetVolumesRequest{},
			clientResponse: &models.GetVolumesResponse{
				Volumes: []*models.Volume{
					{
						UUID:               uuid.NewString(),
						VendorVolumeID:     uuid.NewString(),
						Size:               uint64(defaultOSDiskSizeBytes),
						SectorSize:         sectorSize512,
						ACL:                []string{},
						IsAvailable:        true,
						SourceSnapshotUUID: "",
					},
					{
						UUID:               uuid.NewString(),
						VendorVolumeID:     uuid.NewString(),
						Size:               uint64(defaultOSDiskSizeBytes),
						SectorSize:         sectorSize4096,
						ACL:                []string{uuid.NewString()},
						IsAvailable:        false,
						SourceSnapshotUUID: uuid.NewString(),
					},
				},
			},
			clientErr: nil,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct := NewClientTranslator()
			mc := &mockClient{
				mockGetVolumes: func(ctx context.Context, req *models.GetVolumesRequest) (*models.GetVolumesResponse, error) {
					return tt.clientResponse, tt.clientErr
				},
			}

			resp, err := ct.GetVolumes(context.Background(), mc, tt.input)
			if tt.expectErr {
				require.Error(t, err)
				require.Nil(t, resp)

				return
			}

			require.NoError(t, err)
			require.Equal(t, len(tt.clientResponse.Volumes), len(resp.Volumes))

			for i := 0; i < len(tt.clientResponse.Volumes); i++ {
				require.Equal(t, tt.clientResponse.Volumes[i].UUID, resp.Volumes[i].Uuid)
				require.Equal(t, tt.clientResponse.Volumes[i].VendorVolumeID, resp.Volumes[i].VendorVolumeId)
				require.Equal(t, tt.clientResponse.Volumes[i].Size, resp.Volumes[i].Size)
				require.Equal(t, tt.clientResponse.Volumes[i].IsAvailable, resp.Volumes[i].IsAvailable)
				require.Equal(t, tt.clientResponse.Volumes[i].SourceSnapshotUUID, resp.Volumes[i].SourceSnapshotUuid)
			}
		},
		)

	}
}

func Test_ResizeVolume(t *testing.T) {
	tests := []struct {
		name      string
		input     *storms.ResizeVolumeRequest
		expect    *storms.ResizeVolumeResponse
		expectErr bool
	}{
		{
			name:      "valid",
			input:     &storms.ResizeVolumeRequest{},
			expect:    &storms.ResizeVolumeResponse{},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct := NewClientTranslator()
			mc := &mockClient{
				mockResizeVolume: func(ctx context.Context, req *models.ResizeVolumeRequest) (*models.ResizeVolumeResponse, error) {
					return &models.ResizeVolumeResponse{}, nil
				},
			}
			res, err := ct.ResizeVolume(context.Background(), mc, tt.input)
			if tt.expectErr {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
		},
		)
	}
}

func Test_translateUint32ToSectorSizeEnum(t *testing.T) {
	tests := []struct {
		name     string
		input    uint32
		expected storms.SectorSizeEnum
	}{
		{
			name:     "4096",
			input:    uint32(4096),
			expected: storms.SectorSizeEnum_SECTOR_SIZE_ENUM_4096,
		},
		{
			name:     "512",
			input:    uint32(512),
			expected: storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512,
		},
		{
			name:     "0",
			input:    uint32(0),
			expected: storms.SectorSizeEnum_SECTOR_SIZE_ENUM_UNSPECIFIED,
		},
		{
			name:     "unsupported",
			input:    uint32(1234),
			expected: storms.SectorSizeEnum_SECTOR_SIZE_ENUM_UNSPECIFIED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := translateUint32ToSectorSizeEnum(tt.input)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func Test_translateSectorSizeEnumToUint32(t *testing.T) {
	tests := []struct {
		name      string
		input     storms.SectorSizeEnum
		expected  uint32
		expectErr bool
	}{
		{
			name:      "sector size 512",
			input:     storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512,
			expected:  512,
			expectErr: false,
		},
		{
			name:      "sector size 4096",
			input:     storms.SectorSizeEnum_SECTOR_SIZE_ENUM_4096,
			expected:  4096,
			expectErr: false,
		},
		{
			name:      "sector size unspecified",
			input:     storms.SectorSizeEnum_SECTOR_SIZE_ENUM_UNSPECIFIED,
			expected:  0,
			expectErr: false,
		},
		{
			name:      "sector size unsupported",
			input:     storms.SectorSizeEnum(7),
			expected:  0,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := translateSectorSizeEnumToUint32(tt.input)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, actual, tt.expected)
		})
	}
}
