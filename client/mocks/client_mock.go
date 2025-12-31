package mocks

import (
	"context"

	"gitlab.com/crusoeenergy/island/storage/storms/client"
	"gitlab.com/crusoeenergy/island/storage/storms/client/models"
)

var (
	_ client.Client = (*MockClient)(nil)
)

type MockClient struct {
	MockGetVolume func(ctx context.Context, req *models.GetVolumeRequest,
	) (*models.GetVolumeResponse, error)
	MockGetVolumes func(ctx context.Context, req *models.GetVolumesRequest,
	) (*models.GetVolumesResponse, error)
	MockCreateVolume func(ctx context.Context, req *models.CreateVolumeRequest,
	) (*models.CreateVolumeResponse, error)
	MockResizeVolume func(ctx context.Context, req *models.ResizeVolumeRequest,
	) (*models.ResizeVolumeResponse, error)
	MockDeleteVolume func(ctx context.Context, req *models.DeleteVolumeRequest,
	) (*models.DeleteVolumeResponse, error)
	MockAttachVolume func(ctx context.Context, req *models.AttachVolumeRequest,
	) (*models.AttachVolumeResponse, error)
	MockDetachVolume func(ctx context.Context, req *models.DetachVolumeRequest,
	) (*models.DetachVolumeResponse, error)
	MockGetSnapshot func(ctx context.Context, req *models.GetSnapshotRequest,
	) (*models.GetSnapshotResponse, error)
	MockGetSnapshots func(ctx context.Context, req *models.GetSnapshotsRequest,
	) (*models.GetSnapshotsResponse, error)
	MockCreateSnapshot func(ctx context.Context, req *models.CreateSnapshotRequest,
	) (*models.CreateSnapshotResponse, error)
	MockDeleteSnapshot func(ctx context.Context, req *models.DeleteSnapshotRequest,
	) (*models.DeleteSnapshotResponse, error)
}

func (m *MockClient) GetVolume(
	ctx context.Context, req *models.GetVolumeRequest,
) (*models.GetVolumeResponse, error) {
	return m.MockGetVolume(ctx, req)
}

func (m *MockClient) GetVolumes(
	ctx context.Context, req *models.GetVolumesRequest,
) (*models.GetVolumesResponse, error) {
	return m.MockGetVolumes(ctx, req)
}

func (m *MockClient) CreateVolume(
	ctx context.Context, req *models.CreateVolumeRequest,
) (*models.CreateVolumeResponse, error) {
	return m.MockCreateVolume(ctx, req)
}

func (m *MockClient) ResizeVolume(
	ctx context.Context, req *models.ResizeVolumeRequest,
) (*models.ResizeVolumeResponse, error) {
	return m.MockResizeVolume(ctx, req)
}

func (m *MockClient) DeleteVolume(
	ctx context.Context, req *models.DeleteVolumeRequest,
) (*models.DeleteVolumeResponse, error) {
	return m.MockDeleteVolume(ctx, req)
}

func (m *MockClient) AttachVolume(
	ctx context.Context, req *models.AttachVolumeRequest,
) (*models.AttachVolumeResponse, error) {
	return m.MockAttachVolume(ctx, req)
}

func (m *MockClient) DetachVolume(
	ctx context.Context, req *models.DetachVolumeRequest,
) (*models.DetachVolumeResponse, error) {
	return m.MockDetachVolume(ctx, req)
}

func (m *MockClient) GetSnapshot(
	ctx context.Context, req *models.GetSnapshotRequest,
) (*models.GetSnapshotResponse, error) {
	return m.MockGetSnapshot(ctx, req)
}

func (m *MockClient) GetSnapshots(
	ctx context.Context, req *models.GetSnapshotsRequest,
) (*models.GetSnapshotsResponse, error) {
	return m.MockGetSnapshots(ctx, req)
}

func (m *MockClient) CreateSnapshot(
	ctx context.Context, req *models.CreateSnapshotRequest,
) (*models.CreateSnapshotResponse, error) {
	return m.MockCreateSnapshot(ctx, req)
}

func (m *MockClient) DeleteSnapshot(
	ctx context.Context, req *models.DeleteSnapshotRequest,
) (*models.DeleteSnapshotResponse, error) {
	return m.MockDeleteSnapshot(ctx, req)
}
