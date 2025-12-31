package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gitlab.com/crusoeenergy/island/storage/storms/client"
	clientmocks "gitlab.com/crusoeenergy/island/storage/storms/client/mocks"
	"gitlab.com/crusoeenergy/island/storage/storms/client/models"
	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
	allocatormocks "gitlab.com/crusoeenergy/island/storage/storms/storms/internal/service/allocator/mocks"
	"gitlab.com/crusoeenergy/island/storage/storms/storms/internal/service/cluster"
	clustermocks "gitlab.com/crusoeenergy/island/storage/storms/storms/internal/service/cluster/mocks"
	resource "gitlab.com/crusoeenergy/island/storage/storms/storms/internal/service/resource"
	resourcemocks "gitlab.com/crusoeenergy/island/storage/storms/storms/internal/service/resource/mocks"
	translatormocks "gitlab.com/crusoeenergy/island/storage/storms/storms/internal/service/translator/mocks"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	defaultOSDiskSizeBytes = 137438953472
	sectorSize512          = 512
	sectorSize4096         = 4096
)

var (
	expectedTimestamp = timestamppb.New(time.Date(2025, 11, 15, 10, 30, 0, 0, time.UTC))

	clusterID1  = uuid.NewString()
	mockClient1 = &clientmocks.MockClient{
		MockGetVolumes: func(ctx context.Context, req *models.GetVolumesRequest) (*models.GetVolumesResponse, error) {
			return &models.GetVolumesResponse{
				Volumes: []*models.Volume{
					{
						UUID:               resourceID1,
						VendorVolumeID:     resourceID1,
						Size:               defaultOSDiskSizeBytes,
						SectorSize:         sectorSize512,
						ACL:                []string{},
						IsAvailable:        true,
						SourceSnapshotUUID: "",
						CreatedAt:          expectedTimestamp.AsTime(),
					},
				},
			}, nil
		},
		MockGetSnapshots: func(ctx context.Context, req *models.GetSnapshotsRequest) (*models.GetSnapshotsResponse, error) {
			return &models.GetSnapshotsResponse{
				Snapshots: []*models.Snapshot{},
			}, nil
		},
	}
	resourceID1  = uuid.NewString()
	vendor1      = "vendor-a"
	mockCluster1 = &cluster.Cluster{
		Config: &cluster.Config{
			Vendor:       vendor1,
			ClusterID:    clusterID1,
			AffinityTags: map[string]string{},
			VendorConfig: nil,
		},
		Client: mockClient1,
	}

	vendor2     = "vendor-b"
	clusterID2  = uuid.NewString()
	mockClient2 = &clientmocks.MockClient{
		MockGetVolumes: func(ctx context.Context, req *models.GetVolumesRequest) (*models.GetVolumesResponse, error) {
			return &models.GetVolumesResponse{
				Volumes: []*models.Volume{},
			}, nil
		},
		MockGetSnapshots: func(ctx context.Context, req *models.GetSnapshotsRequest) (*models.GetSnapshotsResponse, error) {
			return &models.GetSnapshotsResponse{
				Snapshots: []*models.Snapshot{
					{
						UUID:             resourceID2,
						VendorSnapshotID: resourceID2,
						Size:             defaultOSDiskSizeBytes,
						SectorSize:       sectorSize4096,
						IsAvailable:      true,
						SourceVolumeUUID: uuid.NewString(),
						CreatedAt:        expectedTimestamp.AsTime(),
					},
				},
			}, nil
		},
	}
	resourceID2  = uuid.NewString()
	mockCluster2 = &cluster.Cluster{
		Config: &cluster.Config{
			Vendor:       vendor2,
			ClusterID:    clusterID2,
			AffinityTags: map[string]string{},
			VendorConfig: nil,
		},
		Client: mockClient2,
	}
)

func Test_GetVolume(t *testing.T) {
	s := &Service{
		clusterManager: &clustermocks.MockClusterManager{
			MockGet: func(clusterID string) (*cluster.Cluster, error) {
				return mockCluster1, nil
			},
		},
		resourceManager: &resourcemocks.MockResourceManager{
			MockGetResourceCluster: func(resourceID string) (string, error) {
				return clusterID1, nil
			},
		},
		clientTranslator: &translatormocks.MockClientTranslator{
			MockGetVolume: func(ctx context.Context, c client.Client, req *storms.GetVolumeRequest,
			) (*storms.GetVolumeResponse, error) {
				return &storms.GetVolumeResponse{
					Volume: &storms.Volume{
						Uuid:               req.GetUuid(),
						Size:               defaultOSDiskSizeBytes,
						SectorSize:         storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512,
						Acl:                []string{},
						IsAvailable:        true,
						SourceSnapshotUuid: "",
						CreatedAt:          expectedTimestamp,
					},
				}, nil
			},
		},
	}

	volID := uuid.NewString()
	req := &storms.GetVolumeRequest{
		Uuid: volID,
	}

	resp, err := s.GetVolume(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, resp.Volume.Uuid, volID)
	require.NotNil(t, resp.Volume.CreatedAt)
	require.Equal(t, expectedTimestamp.AsTime(), resp.Volume.CreatedAt.AsTime())
}

func Test_GetVolumes(t *testing.T) {
	s := &Service{
		clusterManager: &clustermocks.MockClusterManager{
			MockAllIDs: func() []string {
				return []string{clusterID1, clusterID2}
			},
			MockGet: func(clusterID string) (*cluster.Cluster, error) {
				switch clusterID {
				case clusterID1:
					return mockCluster1, nil

				case clusterID2:
					return mockCluster2, nil
				}
				return nil, fmt.Errorf("error")
			},
		},
		resourceManager: &resourcemocks.MockResourceManager{
			MockGetResourceCluster: func(resourceID string) (string, error) {
				switch resourceID {
				case resourceID1:
					return clusterID1, nil

				case resourceID2:
					return clusterID2, nil
				}
				return "", fmt.Errorf("error")
			},
		},
		clientTranslator: &translatormocks.MockClientTranslator{
			MockGetVolumes: func(ctx context.Context, c client.Client, _ *storms.GetVolumesRequest,
			) (*storms.GetVolumesResponse, error) {
				switch c {
				case mockClient1:
					return &storms.GetVolumesResponse{
						Volumes: []*storms.Volume{
							{
								Uuid:               resourceID1,
								Size:               defaultOSDiskSizeBytes,
								SectorSize:         storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512,
								Acl:                []string{},
								IsAvailable:        true,
								SourceSnapshotUuid: "",
								CreatedAt:          expectedTimestamp,
							},
						},
					}, nil
				case mockClient2:
					return &storms.GetVolumesResponse{
						Volumes: []*storms.Volume{
							{
								Uuid:               resourceID2,
								Size:               defaultOSDiskSizeBytes,
								SectorSize:         storms.SectorSizeEnum_SECTOR_SIZE_ENUM_4096,
								Acl:                []string{},
								IsAvailable:        true,
								SourceSnapshotUuid: "",
								CreatedAt:          expectedTimestamp,
							},
						},
					}, nil
				}
				return nil, nil
			},
		},
	}

	req := &storms.GetVolumesRequest{}
	resp, err := s.GetVolumes(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Volumes, 2)
	for _, v := range resp.Volumes {
		require.NotNil(t, v.CreatedAt)
		require.Equal(t, expectedTimestamp.AsTime(), v.CreatedAt.AsTime())
	}
}

func Test_CreateVolume(t *testing.T) {
	tests := []struct {
		name      string
		input     *storms.CreateVolumeRequest
		expectErr bool
	}{
		{
			name: "create new empty volume",
			input: &storms.CreateVolumeRequest{
				Uuid: uuid.NewString(),
				Source: &storms.CreateVolumeRequest_FromNew{
					FromNew: &storms.NewVolumeSpec{
						Size:       defaultOSDiskSizeBytes,
						SectorSize: sectorSize512,
					},
				},
			},
			expectErr: false,
		},
		{
			name: "create volume from snapshot source",
			input: &storms.CreateVolumeRequest{
				Uuid: uuid.NewString(),
				Source: &storms.CreateVolumeRequest_FromSnapshot{
					FromSnapshot: &storms.SnapshotSourceVolumeSpec{
						SnapshotUuid: uuid.NewString(),
					},
				},
			}, expectErr: false,
		},
	}

	s := &Service{
		clusterManager: &clustermocks.MockClusterManager{
			MockGet: func(clusterID string) (*cluster.Cluster, error) {
				return mockCluster1, nil
			},
		},
		resourceManager: &resourcemocks.MockResourceManager{
			MockGetResourceCluster: func(resourceID string) (string, error) {
				return clusterID1, nil
			},
			MockMap: func(r *resource.Resource) error { return nil },
		},
		allocator: &allocatormocks.MockAllocator{
			MockAllocateCluster: func(affinityTags map[string]string) (string, error) {
				return clusterID1, nil
			},
		},
		clientTranslator: &translatormocks.MockClientTranslator{
			MockCreateVolume: func(ctx context.Context, c client.Client, req *storms.CreateVolumeRequest,
			) (*storms.CreateVolumeResponse, error) {
				return &storms.CreateVolumeResponse{}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.CreateVolume(context.Background(), tt.input)
			if tt.expectErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}

func Test_ResizeVolume(t *testing.T) {
	s := &Service{
		clusterManager: &clustermocks.MockClusterManager{
			MockGet: func(clusterID string) (*cluster.Cluster, error) {
				return mockCluster1, nil
			},
		},
		resourceManager: &resourcemocks.MockResourceManager{
			MockGetResourceCluster: func(resourceID string) (string, error) {
				return clusterID1, nil
			},
		},
		allocator: &allocatormocks.MockAllocator{},
		clientTranslator: &translatormocks.MockClientTranslator{
			MockResizeVolume: func(ctx context.Context, c client.Client, req *storms.ResizeVolumeRequest) (*storms.ResizeVolumeResponse, error) {
				return &storms.ResizeVolumeResponse{}, nil
			},
		},
	}
	req := &storms.ResizeVolumeRequest{}

	resp, err := s.ResizeVolume(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func Test_DeleteVolume(t *testing.T) {
	s := &Service{
		clusterManager: &clustermocks.MockClusterManager{
			MockGet: func(clusterID string) (*cluster.Cluster, error) {
				return mockCluster1, nil
			},
		},
		resourceManager: &resourcemocks.MockResourceManager{
			MockGetResourceCluster: func(resourceID string) (string, error) {
				return clusterID1, nil
			},
			MockUnmap: func(resourceID string) error { return nil },
		},
		allocator: &allocatormocks.MockAllocator{},
		clientTranslator: &translatormocks.MockClientTranslator{
			MockDeleteVolume: func(ctx context.Context, c client.Client, req *storms.DeleteVolumeRequest,
			) (*storms.DeleteVolumeResponse, error) {
				return &storms.DeleteVolumeResponse{}, nil
			},
		},
	}
	req := &storms.DeleteVolumeRequest{
		Uuid: uuid.NewString(),
	}

	resp, err := s.DeleteVolume(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func Test_AttachVolume(t *testing.T) {
	s := &Service{
		clusterManager: &clustermocks.MockClusterManager{
			MockGet: func(clusterID string) (*cluster.Cluster, error) {
				return mockCluster1, nil
			},
		},
		resourceManager: &resourcemocks.MockResourceManager{
			MockGetResourceCluster: func(resourceID string) (string, error) {
				return clusterID1, nil
			},
		},
		allocator: &allocatormocks.MockAllocator{},
		clientTranslator: &translatormocks.MockClientTranslator{
			MockAttachVolume: func(ctx context.Context, c client.Client, req *storms.AttachVolumeRequest,
			) (*storms.AttachVolumeResponse, error) {
				return &storms.AttachVolumeResponse{}, nil
			},
		},
	}
	req := &storms.AttachVolumeRequest{
		Uuid: uuid.NewString(),
		Acl:  []string{uuid.NewString()},
	}

	resp, err := s.AttachVolume(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

}

func Test_DetachVolume(t *testing.T) {
	s := &Service{
		clusterManager: &clustermocks.MockClusterManager{
			MockGet: func(clusterID string) (*cluster.Cluster, error) {
				return mockCluster1, nil
			},
		},
		resourceManager: &resourcemocks.MockResourceManager{
			MockGetResourceCluster: func(resourceID string) (string, error) {
				return clusterID1, nil
			},
		},
		allocator: &allocatormocks.MockAllocator{},
		clientTranslator: &translatormocks.MockClientTranslator{
			MockDetachVolume: func(ctx context.Context, c client.Client, req *storms.DetachVolumeRequest,
			) (*storms.DetachVolumeResponse, error) {
				return &storms.DetachVolumeResponse{}, nil
			},
		},
	}
	req := &storms.DetachVolumeRequest{
		Uuid: uuid.NewString(),
	}

	resp, err := s.DetachVolume(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func Test_GetSnapshot(t *testing.T) {
	s := &Service{
		clusterManager: &clustermocks.MockClusterManager{
			MockGet: func(clusterID string) (*cluster.Cluster, error) {
				return mockCluster1, nil
			},
		},
		resourceManager: &resourcemocks.MockResourceManager{
			MockGetResourceCluster: func(resourceID string) (string, error) {
				return clusterID1, nil
			},
		},
		clientTranslator: &translatormocks.MockClientTranslator{
			MockGetSnapshot: func(ctx context.Context, c client.Client, req *storms.GetSnapshotRequest,
			) (*storms.GetSnapshotResponse, error) {
				return &storms.GetSnapshotResponse{
					Snapshot: &storms.Snapshot{
						Uuid:             req.GetUuid(),
						Size:             defaultOSDiskSizeBytes,
						SectorSize:       storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512,
						IsAvailable:      true,
						SourceVolumeUuid: uuid.NewString(),
						CreatedAt:        expectedTimestamp,
					},
				}, nil
			},
		},
	}

	snapshotID := uuid.NewString()
	req := &storms.GetSnapshotRequest{
		Uuid: snapshotID,
	}

	resp, err := s.GetSnapshot(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, resp.Snapshot.Uuid, snapshotID)
	require.NotNil(t, resp.Snapshot.CreatedAt)
	require.Equal(t, expectedTimestamp.AsTime(), resp.Snapshot.CreatedAt.AsTime())

}

func Test_GetSnapshots(t *testing.T) {
	s := &Service{
		clusterManager: &clustermocks.MockClusterManager{
			MockAllIDs: func() []string {
				return []string{clusterID1, clusterID2}
			},
			MockGet: func(clusterID string) (*cluster.Cluster, error) {
				switch clusterID {
				case clusterID1:
					return mockCluster1, nil

				case clusterID2:
					return mockCluster2, nil
				}
				return nil, fmt.Errorf("error")
			},
		},
		resourceManager: &resourcemocks.MockResourceManager{
			MockGetResourceCluster: func(resourceID string) (string, error) {
				switch resourceID {
				case resourceID1:
					return clusterID1, nil

				case resourceID2:
					return clusterID2, nil
				}
				return "", fmt.Errorf("error")
			},
		},
		clientTranslator: &translatormocks.MockClientTranslator{
			MockGetSnapshots: func(ctx context.Context, c client.Client, _ *storms.GetSnapshotsRequest) (*storms.GetSnapshotsResponse, error) {
				switch c {
				case mockClient1:
					return &storms.GetSnapshotsResponse{
						Snapshots: []*storms.Snapshot{
							{
								Uuid:             resourceID1,
								Size:             defaultOSDiskSizeBytes,
								SectorSize:       storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512,
								IsAvailable:      true,
								SourceVolumeUuid: uuid.NewString(),
								CreatedAt:        expectedTimestamp,
							},
						},
					}, nil
				case mockClient2:
					return &storms.GetSnapshotsResponse{
						Snapshots: []*storms.Snapshot{
							{
								Uuid:             resourceID2,
								Size:             defaultOSDiskSizeBytes,
								SectorSize:       storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512,
								IsAvailable:      true,
								SourceVolumeUuid: uuid.NewString(),
								CreatedAt:        expectedTimestamp,
							},
						},
					}, nil
				}
				return nil, fmt.Errorf("error")
			},
		},
	}

	req := &storms.GetSnapshotsRequest{}

	resp, err := s.GetSnapshots(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Snapshots, 2)
	for _, s := range resp.Snapshots {
		require.NotNil(t, s.CreatedAt)
		require.Equal(t, expectedTimestamp.AsTime(), s.CreatedAt.AsTime())
	}

}

func Test_CreateSnapshot(t *testing.T) {
	s := &Service{
		clusterManager: &clustermocks.MockClusterManager{
			MockGet: func(clusterID string) (*cluster.Cluster, error) {
				return mockCluster1, nil
			},
		},
		resourceManager: &resourcemocks.MockResourceManager{
			MockGetResourceCluster: func(resourceID string) (string, error) {
				return clusterID1, nil
			},
			MockMap: func(r *resource.Resource) error { return nil },
		},
		allocator: &allocatormocks.MockAllocator{
			MockAllocateCluster: func(affinityTags map[string]string) (string, error) {
				return clusterID1, nil
			},
		},
		clientTranslator: &translatormocks.MockClientTranslator{
			MockCreateSnapshot: func(ctx context.Context, c client.Client, req *storms.CreateSnapshotRequest) (*storms.CreateSnapshotResponse, error) {
				return &storms.CreateSnapshotResponse{}, nil
			},
		},
	}

	req := &storms.CreateSnapshotRequest{
		Uuid:          uuid.NewString(),
		SrcVolumeUuid: uuid.NewString(),
	}
	resp, err := s.CreateSnapshot(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func Test_DeleteSnapshot(t *testing.T) {
	s := &Service{
		clusterManager: &clustermocks.MockClusterManager{
			MockGet: func(clusterID string) (*cluster.Cluster, error) {
				return mockCluster1, nil
			},
		},
		resourceManager: &resourcemocks.MockResourceManager{
			MockGetResourceCluster: func(resourceID string) (string, error) {
				return clusterID1, nil
			},
			MockUnmap: func(resourceID string) error { return nil },
		},
		allocator: &allocatormocks.MockAllocator{},
		clientTranslator: &translatormocks.MockClientTranslator{
			MockDeleteSnapshot: func(ctx context.Context, c client.Client, req *storms.DeleteSnapshotRequest,
			) (*storms.DeleteSnapshotResponse, error) {
				return &storms.DeleteSnapshotResponse{}, nil
			},
		},
	}
	req := &storms.DeleteSnapshotRequest{}

	resp, err := s.DeleteSnapshot(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func Test_SyncResource(t *testing.T) {
	s := &Service{
		clusterManager: &clustermocks.MockClusterManager{
			MockGet: func(clusterID string) (*cluster.Cluster, error) {
				switch clusterID {
				case clusterID1:
					return mockCluster1, nil
				case clusterID2:
					return mockCluster2, nil
				}
				return nil, fmt.Errorf("error")
			},
		},
		resourceManager: &resourcemocks.MockResourceManager{
			MockMap:   func(r *resource.Resource) error { return nil },
			MockUnmap: func(resourceID string) error { return nil },
			MockGetResourceCluster: func(resourceID string) (string, error) {
				switch resourceID {
				case resourceID1:
					return clusterID1, nil
				case resourceID2:
					return clusterID2, nil
				}
				return "", fmt.Errorf("error")
			},
		},
		allocator: &allocatormocks.MockAllocator{},
		clientTranslator: &translatormocks.MockClientTranslator{
			MockGetVolume: func(ctx context.Context, c client.Client, req *storms.GetVolumeRequest,
			) (*storms.GetVolumeResponse, error) {
				return &storms.GetVolumeResponse{
					Volume: &storms.Volume{
						Uuid:               req.GetUuid(),
						Size:               defaultOSDiskSizeBytes,
						SectorSize:         storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512,
						Acl:                []string{},
						IsAvailable:        true,
						SourceSnapshotUuid: "",
						CreatedAt:          expectedTimestamp,
					},
				}, nil
			},
			MockGetSnapshot: func(ctx context.Context, c client.Client, req *storms.GetSnapshotRequest,
			) (*storms.GetSnapshotResponse, error) {
				return &storms.GetSnapshotResponse{
					Snapshot: &storms.Snapshot{
						Uuid:             req.GetUuid(),
						Size:             defaultOSDiskSizeBytes,
						SectorSize:       storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512,
						IsAvailable:      true,
						SourceVolumeUuid: uuid.NewString(),
						CreatedAt:        expectedTimestamp,
					},
				}, nil
			},
		},
	}

	tests := []struct {
		name         string
		resourceType storms.ResourceType
		resourceID   string
		clusterID    string
		expectErr    bool
	}{
		{
			name:         "sync volume",
			resourceType: storms.ResourceType_RESOURCE_TYPE_VOLUME,
			resourceID:   resourceID1,
			clusterID:    clusterID1,
			expectErr:    false,
		},
		{
			name:         "sync snapshot",
			resourceType: storms.ResourceType_RESOURCE_TYPE_SNAPSHOT,
			resourceID:   resourceID2,
			clusterID:    clusterID2,
			expectErr:    false,
		},
		{
			name:         "sync unspecified resource type ",
			resourceType: storms.ResourceType_RESOURCE_TYPE_UNSPECIFIED,
			resourceID:   resourceID1,
			clusterID:    clusterID1,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := &storms.SyncResourceRequest{
				ResourceType: tt.resourceType,
				Uuid:         tt.resourceID,
				ClusterUuid:  tt.clusterID,
			}

			resp, err := s.SyncResource(context.Background(), req)
			if tt.expectErr {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}

func Test_SyncAllResources(t *testing.T) {
	s := &Service{
		clusterManager: &clustermocks.MockClusterManager{
			MockGet: func(clusterID string) (*cluster.Cluster, error) {
				switch clusterID {
				case clusterID1:
					return mockCluster1, nil
				case clusterID2:
					return mockCluster2, nil
				}
				return nil, fmt.Errorf("error")
			},
			MockAllIDs: func() []string {
				return []string{
					clusterID1, clusterID2,
				}
			},
		},
		resourceManager: &resourcemocks.MockResourceManager{
			MockMap:   func(r *resource.Resource) error { return nil },
			MockUnmap: func(resourceID string) error { return nil },
			MockGetResourceCluster: func(resourceID string) (string, error) {
				switch resourceID {
				case resourceID1:
					return clusterID1, nil
				case resourceID2:
					return clusterID2, nil
				}
				return "", fmt.Errorf("error")
			},
			MockGetResourcesOfAllClusters: func() map[string][]*resource.Resource {
				return map[string][]*resource.Resource{
					clusterID1: {
						&resource.Resource{
							ID:           resourceID1,
							ClusterID:    clusterID1,
							ResourceType: resource.TypeVolume,
						},
					},
					clusterID2: {
						&resource.Resource{
							ID:           resourceID1,
							ClusterID:    clusterID2,
							ResourceType: resource.TypeSnapshot,
						}},
				}
			},
		},
		allocator:        &allocatormocks.MockAllocator{},
		clientTranslator: &translatormocks.MockClientTranslator{},
	}

	resp, err := s.SyncAllResources(context.Background(), &storms.SyncAllResourcesRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}
