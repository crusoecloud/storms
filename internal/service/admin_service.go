package service

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	admin "gitlab.com/crusoeenergy/island/storage/storms/api/gen/go/admin/v1"
	resource "gitlab.com/crusoeenergy/island/storage/storms/internal/service/resource"
)

func (s *Service) ReloadConfig(_ context.Context, _ *admin.ReloadConfigRequest,
) (*admin.ReloadConfigResponse, error) {
	err := s.loadClusterConfigs() // Effectively, reloads config file
	if err != nil {
		return nil, fmt.Errorf("failed to re-load cluster configuration: %w", err)
	}

	s.syncClusterManager()
	s.syncResourceManager()

	log.Info().Msgf("Config reloaded.")

	return &admin.ReloadConfigResponse{}, nil
}

func (s *Service) ShowClusters(context.Context, *admin.ShowClustersRequest) (*admin.ShowClustersResponse, error) {
	clusters := []*admin.Cluster{}
	clusterIDs := s.clusterManager.AllIDs()
	for _, clusterID := range clusterIDs {
		resources := s.resourceManager.GetResourcesOfCluster(clusterID)
		vendor := s.getClusterVendor(clusterID)
		cluster := &admin.Cluster{
			Id:     clusterID,
			Vendor: vendor,
			ResourceCount: map[string]int32{
				"volume": int32(lo.CountBy(resources, func(r *resource.Resource) bool {
					return r.ResourceType == resource.TypeVolume
				})),
				"snapshot": int32(lo.CountBy(resources, func(r *resource.Resource) bool {
					return r.ResourceType == resource.TypeSnapshot
				}))},
		}
		clusters = append(clusters, cluster)
	}

	return &admin.ShowClustersResponse{
		Clusters: clusters,
	}, nil
}

func (s *Service) getClusterVendor(clusterID string) string {
	for _, cluster := range s.clusterConfigs.Clusters {
		if cluster.ClusterID == clusterID {
			return cluster.Vendor
		}
	}

	return "UNKNOWN"
}
