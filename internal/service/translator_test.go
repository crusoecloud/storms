package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	storms "gitlab.com/crusoeenergy/island/storage/storms/api/gen/go/storms/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/client/models"
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
	mockCloneVolume    func(ctx context.Context, req *models.CloneVolumeRequest) (*models.CloneVolumeResponse, error)
	mockCloneSnapshot  func(ctx context.Context, req *models.CloneSnapshotRequest) (*models.CloneSnapshotResponse, error)
	mockGetCloneStatus func(ctx context.Context, req *models.GetCloneStatusRequest) (*models.GetCloneStatusResponse, error)
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

func (m *mockClient) CloneVolume(ctx context.Context, req *models.CloneVolumeRequest) (*models.CloneVolumeResponse, error) {
	return m.mockCloneVolume(ctx, req)
}

func (m *mockClient) CloneSnapshot(ctx context.Context, req *models.CloneSnapshotRequest) (*models.CloneSnapshotResponse, error) {
	return m.mockCloneSnapshot(ctx, req)
}

func (m *mockClient) GetCloneStatus(ctx context.Context, req *models.GetCloneStatusRequest) (*models.GetCloneStatusResponse, error) {
	return m.mockGetCloneStatus(ctx, req)
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

				Acls: []string{},
			},
			expectErr: false,
		},
		{
			name: "1 acl",
			input: &storms.AttachVolumeRequest{
				Uuid: "e63b6b55-acfe-446b-95c4-f6b37d94b4b4",
				Acls: []string{
					"e4c2e3f8-a0bd-4e37-b5d0-a107137651fc",
				},
			},
			expectErr: false,
		},
		{
			name: "2 acl",
			input: &storms.AttachVolumeRequest{
				Uuid: "e63b6b55-acfe-446b-95c4-f6b37d94b4b4",
				Acls: []string{
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

func Test_CloneSnapshot(t *testing.T) {
	tests := []struct {
		name      string
		input     *storms.CloneSnapshotRequest
		expect    *storms.CloneSnapshotResponse
		expectErr bool
	}{
		{
			name:      "valid",
			input:     &storms.CloneSnapshotRequest{},
			expect:    &storms.CloneSnapshotResponse{},
			expectErr: false,
		},
	}

	ct := NewClientTranslator()
	mc := &mockClient{
		mockCloneSnapshot: func(ctx context.Context, req *models.CloneSnapshotRequest) (*models.CloneSnapshotResponse, error) {
			return &models.CloneSnapshotResponse{}, nil
		},
	}

	for _, tt := range tests {
		_, err := ct.CloneSnapshot(context.Background(), mc, tt.input)

		require.Error(t, err) // Expected, because not implemented.
	}
}
func Test_CloneVolume(t *testing.T) {
	tests := []struct {
		name      string
		input     *storms.CloneVolumeRequest
		expect    *storms.CloneVolumeResponse
		expectErr bool
	}{
		{
			name:      "valid",
			input:     &storms.CloneVolumeRequest{},
			expect:    &storms.CloneVolumeResponse{},
			expectErr: false,
		},
	}

	ct := NewClientTranslator()
	mc := &mockClient{
		mockCloneVolume: func(ctx context.Context, req *models.CloneVolumeRequest) (*models.CloneVolumeResponse, error) {
			return &models.CloneVolumeResponse{}, nil
		},
	}

	for _, tt := range tests {
		_, err := ct.CloneVolume(context.Background(), mc, tt.input)
		require.Error(t, err) // Expected, because not implemented.
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
				Snapshot: &storms.Snapshot{
					Uuid:             uuid.NewString(),
					Size:             137438953472,
					SectorSize:       storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512,
					SourceVolumeUuid: uuid.NewString(),
				},
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
			name: "valid",
			input: &storms.CreateVolumeRequest{
				Volume: &storms.Volume{
					Uuid:        uuid.NewString(),
					Size:        defaultOSDiskSizeBytes,
					SectorSize:  storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512,
					Acls:        []string{},
					IsAvailable: true,
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
func Test_GetCloneStatus(t *testing.T) {
	tests := []struct {
		name      string
		input     *storms.GetCloneStatusRequest
		expect    *storms.GetCloneStatusResponse
		expectErr bool
	}{
		{
			name:      "valid",
			input:     &storms.GetCloneStatusRequest{},
			expect:    &storms.GetCloneStatusResponse{},
			expectErr: false,
		},
	}

	ct := NewClientTranslator()
	mc := &mockClient{
		mockGetCloneStatus: func(ctx context.Context, req *models.GetCloneStatusRequest) (*models.GetCloneStatusResponse, error) {
			return &models.GetCloneStatusResponse{}, nil
		},
	}

	for _, tt := range tests {
		_, err := ct.GetCloneStatus(context.Background(), mc, tt.input)
		require.Error(t, err) // Expected because not implented.
	}
}

func Test_GetSnapshot(t *testing.T) {
	snapshotUUID := "4533ae7a-ef23-43c5-8e94-68dfcd5bedd6"
	sourceVolumeUUID := "cc23f964-d2d3-40af-a6f5-82b1e0574c93"

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
	}

}

func Test_GetSnapshots(t *testing.T) {
	tests := []struct {
		name      string
		input     *storms.GetSnapshotsRequest
		expect    *storms.GetSnapshotsResponse
		expectErr bool
	}{
		{
			name:      "valid",
			input:     &storms.GetSnapshotsRequest{},
			expect:    &storms.GetSnapshotsResponse{},
			expectErr: false,
		},
	}

	ct := NewClientTranslator()
	mc := &mockClient{
		mockGetSnapshots: func(ctx context.Context, req *models.GetSnapshotsRequest) (*models.GetSnapshotsResponse, error) {
			return &models.GetSnapshotsResponse{}, nil
		},
	}

	for _, tt := range tests {
		res, err := ct.GetSnapshots(context.Background(), mc, tt.input)
		if tt.expectErr {
			require.Error(t, err)

			continue
		}
		require.NoError(t, err)
		require.NotNil(t, res)

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
				Uuid: uuid.NewString(),
			},
			expect:    &storms.GetVolumeResponse{},
			expectErr: false,
		},
	}

	ct := NewClientTranslator()
	mc := &mockClient{
		mockGetVolume: func(ctx context.Context, req *models.GetVolumeRequest) (*models.GetVolumeResponse, error) {
			return &models.GetVolumeResponse{
				Volume: &models.Volume{
					UUID:               req.UUID,
					Size:               defaultOSDiskSizeBytes,
					SectorSize:         sectorSize512,
					Acls:               []string{},
					IsAvailable:        true,
					SourceSnapshotUUID: "",
				},
			}, nil
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

	}
}
func Test_GetVolumes(t *testing.T) {
	tests := []struct {
		name      string
		input     *storms.GetVolumesRequest
		expect    *storms.GetVolumesResponse
		expectErr bool
	}{
		{
			name:      "valid",
			input:     &storms.GetVolumesRequest{},
			expect:    &storms.GetVolumesResponse{},
			expectErr: false,
		},
	}

	ct := NewClientTranslator()
	mc := &mockClient{
		mockGetVolumes: func(ctx context.Context, req *models.GetVolumesRequest) (*models.GetVolumesResponse, error) {
			return &models.GetVolumesResponse{}, nil
		},
	}

	for _, tt := range tests {
		res, err := ct.GetVolumes(context.Background(), mc, tt.input)
		if tt.expectErr {
			require.Error(t, err)

			continue
		}
		require.NoError(t, err)
		require.NotNil(t, res)

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

	ct := NewClientTranslator()
	mc := &mockClient{
		mockResizeVolume: func(ctx context.Context, req *models.ResizeVolumeRequest) (*models.ResizeVolumeResponse, error) {
			return &models.ResizeVolumeResponse{}, nil
		},
	}

	for _, tt := range tests {
		res, err := ct.ResizeVolume(context.Background(), mc, tt.input)
		if tt.expectErr {
			require.Error(t, err)

			continue
		}
		require.NoError(t, err)
		require.NotNil(t, res)

	}
}
