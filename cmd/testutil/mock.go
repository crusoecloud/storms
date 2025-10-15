package testutil

import (
	"context"

	storms "gitlab.com/crusoeenergy/island/storage/storms/api/gen/go/storms/v1"
	"google.golang.org/grpc"
)

type MockStorMSClient struct {
	MockGetVolume func(ctx context.Context, in *storms.GetVolumeRequest, opts ...grpc.CallOption,
	) (*storms.GetVolumeResponse, error)
	MockGetVolumes func(ctx context.Context, in *storms.GetVolumesRequest, opts ...grpc.CallOption,
	) (*storms.GetVolumesResponse, error)
	MockCreateVolume func(ctx context.Context, in *storms.CreateVolumeRequest, opts ...grpc.CallOption,
	) (*storms.CreateVolumeResponse, error)
	MockResizeVolume func(ctx context.Context, in *storms.ResizeVolumeRequest, opts ...grpc.CallOption,
	) (*storms.ResizeVolumeResponse, error)
	MockDeleteVolume func(ctx context.Context, in *storms.DeleteVolumeRequest, opts ...grpc.CallOption,
	) (*storms.DeleteVolumeResponse, error)
	MockAttachVolume func(ctx context.Context, in *storms.AttachVolumeRequest, opts ...grpc.CallOption,
	) (*storms.AttachVolumeResponse, error)
	MockDetachVolume func(ctx context.Context, in *storms.DetachVolumeRequest, opts ...grpc.CallOption,
	) (*storms.DetachVolumeResponse, error)
	MockGetSnapshot func(ctx context.Context, in *storms.GetSnapshotRequest, opts ...grpc.CallOption,
	) (*storms.GetSnapshotResponse, error)
	MockGetSnapshots func(ctx context.Context, in *storms.GetSnapshotsRequest, opts ...grpc.CallOption,
	) (*storms.GetSnapshotsResponse, error)
	MockCreateSnapshot func(ctx context.Context, in *storms.CreateSnapshotRequest, opts ...grpc.CallOption,
	) (*storms.CreateSnapshotResponse, error)
	MockDeleteSnapshot func(ctx context.Context, in *storms.DeleteSnapshotRequest, opts ...grpc.CallOption,
	) (*storms.DeleteSnapshotResponse, error)
	MockSyncResource func(ctx context.Context, in *storms.SyncResourceRequest, opts ...grpc.CallOption,
	) (*storms.SyncResourceResponse, error)
	MockSyncAllResources func(ctx context.Context, in *storms.SyncAllResourcesRequest, opts ...grpc.CallOption,
	) (*storms.SyncAllResourcesResponse, error)
}

func (m *MockStorMSClient) GetVolume(
	ctx context.Context, in *storms.GetVolumeRequest, opts ...grpc.CallOption,
) (*storms.GetVolumeResponse, error) {
	return m.MockGetVolume(ctx, in, opts...)
}

func (m *MockStorMSClient) GetVolumes(
	ctx context.Context, in *storms.GetVolumesRequest, opts ...grpc.CallOption,
) (*storms.GetVolumesResponse, error) {
	return m.MockGetVolumes(ctx, in, opts...)
}

func (m *MockStorMSClient) CreateVolume(
	ctx context.Context, in *storms.CreateVolumeRequest, opts ...grpc.CallOption,
) (*storms.CreateVolumeResponse, error) {
	return m.MockCreateVolume(ctx, in, opts...)
}

func (m *MockStorMSClient) ResizeVolume(
	ctx context.Context, in *storms.ResizeVolumeRequest, opts ...grpc.CallOption,
) (*storms.ResizeVolumeResponse, error) {
	return m.MockResizeVolume(ctx, in, opts...)
}

func (m *MockStorMSClient) DeleteVolume(
	ctx context.Context, in *storms.DeleteVolumeRequest, opts ...grpc.CallOption,
) (*storms.DeleteVolumeResponse, error) {
	return m.MockDeleteVolume(ctx, in, opts...)
}

func (m *MockStorMSClient) AttachVolume(
	ctx context.Context, in *storms.AttachVolumeRequest, opts ...grpc.CallOption,
) (*storms.AttachVolumeResponse, error) {
	return m.MockAttachVolume(ctx, in, opts...)
}

func (m *MockStorMSClient) DetachVolume(
	ctx context.Context, in *storms.DetachVolumeRequest, opts ...grpc.CallOption,
) (*storms.DetachVolumeResponse, error) {
	return m.MockDetachVolume(ctx, in, opts...)
}

func (m *MockStorMSClient) GetSnapshot(
	ctx context.Context, in *storms.GetSnapshotRequest, opts ...grpc.CallOption,
) (*storms.GetSnapshotResponse, error) {
	return m.MockGetSnapshot(ctx, in, opts...)
}

func (m *MockStorMSClient) GetSnapshots(
	ctx context.Context, in *storms.GetSnapshotsRequest, opts ...grpc.CallOption,
) (*storms.GetSnapshotsResponse, error) {
	return m.MockGetSnapshots(ctx, in, opts...)
}

func (m *MockStorMSClient) CreateSnapshot(
	ctx context.Context, in *storms.CreateSnapshotRequest, opts ...grpc.CallOption,
) (*storms.CreateSnapshotResponse, error) {
	return m.MockCreateSnapshot(ctx, in, opts...)
}

func (m *MockStorMSClient) DeleteSnapshot(
	ctx context.Context, in *storms.DeleteSnapshotRequest, opts ...grpc.CallOption,
) (*storms.DeleteSnapshotResponse, error) {
	return m.MockDeleteSnapshot(ctx, in, opts...)
}

func (m *MockStorMSClient) SyncResource(
	ctx context.Context, in *storms.SyncResourceRequest, opts ...grpc.CallOption,
) (*storms.SyncResourceResponse, error) {
	return m.MockSyncResource(ctx, in, opts...)
}

func (m *MockStorMSClient) SyncAllResources(
	ctx context.Context, in *storms.SyncAllResourcesRequest, opts ...grpc.CallOption,
) (*storms.SyncAllResourcesResponse, error) {
	return m.MockSyncAllResources(ctx, in, opts...)
}

type MockCloser struct{}

func (m *MockCloser) Close() error {
	return nil
}
