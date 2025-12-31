package mocks

import (
	"context"

	"gitlab.com/crusoeenergy/island/storage/storms/client"
	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
)

type MockClientTranslator struct {
	MockAttachVolume func(ctx context.Context, c client.Client, req *storms.AttachVolumeRequest,
	) (*storms.AttachVolumeResponse, error)
	MockCreateSnapshot func(ctx context.Context, c client.Client, req *storms.CreateSnapshotRequest,
	) (*storms.CreateSnapshotResponse, error)
	MockCreateVolume func(ctx context.Context, c client.Client, req *storms.CreateVolumeRequest,
	) (*storms.CreateVolumeResponse, error)
	MockDeleteSnapshot func(ctx context.Context, c client.Client, req *storms.DeleteSnapshotRequest,
	) (*storms.DeleteSnapshotResponse, error)
	MockDeleteVolume func(ctx context.Context, c client.Client, req *storms.DeleteVolumeRequest,
	) (*storms.DeleteVolumeResponse, error)
	MockDetachVolume func(ctx context.Context, c client.Client, req *storms.DetachVolumeRequest,
	) (*storms.DetachVolumeResponse, error)
	MockGetSnapshot func(ctx context.Context, c client.Client, req *storms.GetSnapshotRequest,
	) (*storms.GetSnapshotResponse, error)
	MockGetSnapshots func(ctx context.Context, c client.Client, _ *storms.GetSnapshotsRequest,
	) (*storms.GetSnapshotsResponse, error)
	MockGetVolume func(ctx context.Context, c client.Client, req *storms.GetVolumeRequest,
	) (*storms.GetVolumeResponse, error)
	MockGetVolumes func(ctx context.Context, c client.Client, _ *storms.GetVolumesRequest,
	) (*storms.GetVolumesResponse, error)
	MockResizeVolume func(ctx context.Context, c client.Client, req *storms.ResizeVolumeRequest,
	) (*storms.ResizeVolumeResponse, error)
}

func (m *MockClientTranslator) AttachVolume(ctx context.Context, c client.Client, req *storms.AttachVolumeRequest,
) (*storms.AttachVolumeResponse, error) {
	return m.MockAttachVolume(ctx, c, req)
}

func (m *MockClientTranslator) CreateSnapshot(ctx context.Context, c client.Client, req *storms.CreateSnapshotRequest,
) (*storms.CreateSnapshotResponse, error) {
	return m.MockCreateSnapshot(ctx, c, req)
}

func (m *MockClientTranslator) CreateVolume(ctx context.Context, c client.Client, req *storms.CreateVolumeRequest,
) (*storms.CreateVolumeResponse, error) {
	return m.MockCreateVolume(ctx, c, req)
}

func (m *MockClientTranslator) DeleteSnapshot(ctx context.Context, c client.Client, req *storms.DeleteSnapshotRequest,
) (*storms.DeleteSnapshotResponse, error) {
	return m.MockDeleteSnapshot(ctx, c, req)
}

func (m *MockClientTranslator) DeleteVolume(ctx context.Context, c client.Client, req *storms.DeleteVolumeRequest,
) (*storms.DeleteVolumeResponse, error) {
	return m.MockDeleteVolume(ctx, c, req)
}

func (m *MockClientTranslator) DetachVolume(ctx context.Context, c client.Client, req *storms.DetachVolumeRequest,
) (*storms.DetachVolumeResponse, error) {
	return m.MockDetachVolume(ctx, c, req)
}

func (m *MockClientTranslator) GetSnapshot(ctx context.Context, c client.Client, req *storms.GetSnapshotRequest,
) (*storms.GetSnapshotResponse, error) {
	return m.MockGetSnapshot(ctx, c, req)
}

func (m *MockClientTranslator) GetSnapshots(ctx context.Context, c client.Client, req *storms.GetSnapshotsRequest,
) (*storms.GetSnapshotsResponse, error) {
	return m.MockGetSnapshots(ctx, c, req)
}

func (m *MockClientTranslator) GetVolume(ctx context.Context, c client.Client, req *storms.GetVolumeRequest,
) (*storms.GetVolumeResponse, error) {
	return m.MockGetVolume(ctx, c, req)
}

func (m *MockClientTranslator) GetVolumes(ctx context.Context, c client.Client, req *storms.GetVolumesRequest,
) (*storms.GetVolumesResponse, error) {
	return m.MockGetVolumes(ctx, c, req)
}

func (m *MockClientTranslator) ResizeVolume(ctx context.Context, c client.Client, req *storms.ResizeVolumeRequest,
) (*storms.ResizeVolumeResponse, error) {
	return m.MockResizeVolume(ctx, c, req)
}
