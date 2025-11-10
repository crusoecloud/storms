package testutil

import (
	"context"

	"gitlab.com/crusoeenergy/island/storage/storms/client"
	"gitlab.com/crusoeenergy/island/storage/storms/client/models"
	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/storms/internal/service/resource"
)

// BEGIN MockClient.
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

// END MockClient.

// BEGIN MockClientTranslator.

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

// END ClientTranslator.

// BEGIN MockClusterManager.

type MockClusterManager struct {
	MockSet    func(clusterID string, client client.Client) error
	MockRemove func(clusterID string) error
	MockGet    func(clusterID string) (client.Client, error)
	MockAllIDs func() []string
	MockCount  func() int
}

func (m *MockClusterManager) Set(clusterID string, c client.Client) error {
	return m.MockSet(clusterID, c)
}

func (m *MockClusterManager) Remove(clusterID string) error {
	return m.MockRemove(clusterID)
}

func (m *MockClusterManager) Get(clusterID string) (client.Client, error) {
	return m.MockGet(clusterID)
}

func (m *MockClusterManager) AllIDs() []string {
	return m.MockAllIDs()
}

func (m *MockClusterManager) Count() int {
	return m.MockCount()
}

// END ClusterManager.

// BEGIN MockResourceManager

type MockResourceManager struct {
	MockMap                       func(r *resource.Resource) error
	MockUnmap                     func(resourceID string) error
	MockGetResourceCluster        func(resourceID string) (string, error)
	MockGetResourceCount          func() int
	MockGetResourcesOfCluster     func(clusterID string) []*resource.Resource
	MockGetResourcesOfAllClusters func() map[string][]*resource.Resource
}

func (m *MockResourceManager) Map(r *resource.Resource) error {
	return m.MockMap(r)
}

func (m *MockResourceManager) Unmap(resourceID string) error {
	return m.MockUnmap(resourceID)
}

func (m *MockResourceManager) GetResourceCluster(resourceID string) (string, error) {
	return m.MockGetResourceCluster(resourceID)
}

func (m *MockResourceManager) GetResourceCount() int {
	return m.MockGetResourceCount()
}

func (m *MockResourceManager) GetResourcesOfCluster(clusterID string) []*resource.Resource {
	return m.MockGetResourcesOfCluster(clusterID)
}

func (m *MockResourceManager) GetResourcesOfAllClusters() map[string][]*resource.Resource {
	return m.MockGetResourcesOfAllClusters()
}

// END MockResourceManager

// BEGIN MockResourceMapper.

// type MockResourceMapper struct {
// 	MockMap                    func(resourceID, clusterID string) error
// 	MockUnmap                  func(resourceID string) error
// 	MockOwnerCluster           func(resourceID string) (string, error)
// 	MockResourceCount          func() int
// 	MockResourceCountByCluster func(clusterID string) int
// 	MockGetAllClusterResources func() map[string][]string
// }

// func (m *MockResourceMapper) Map(resourceID, clusterID string) error {
// 	return m.MocMockkMap(resourceID, clusterID)
// }

// func (m *MockResourceMapper) Unmap(resourceID string) error {
// 	return m.MocMockkUnmap(resourceID)
// }

// func (m *MockResourceMapper) OwnerCluster(resourceID string) (string, error) {
// 	return m.MocMockkOwnerCluster(resourceID)
// }

// func (m *MockResourceMapper) ResourceCount() int {
// 	return m.MocMockkResourceCount()
// }

// func (m *MockResourceMapper) ResourceCountByCluster(clusterID string) int {
// 	return m.MocMockkResourceCountByCluster(clusterID)
// }

// func (m *MockResourceMapper) GetAllClusterResources() map[string][]string {
// 	return m.MocMockkGetAllClusterResources()
// }

// END ResourceMapper.

// BEGIN MockAllocator.

type MockAllocator struct {
	MockSelectClusterForNewResource func() (string, error)
}

func (m *MockAllocator) SelectClusterForNewResource() (string, error) {
	return m.MockSelectClusterForNewResource()
}

// END Allocator.
