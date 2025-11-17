package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/storms/internal/service/resource"
)

var (
	errUnexpected          = errors.New("an unexpected error occurred")
	errUnspecifiedResource = errors.New("unspecified resource")
)

func (s *Service) GetVolume(ctx context.Context, req *storms.GetVolumeRequest,
) (*storms.GetVolumeResponse, error) {
	volID := req.GetUuid()

	clusterID, c, err := s.getClientForResource(volID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for resource: %w", err)
	}

	resp, err := s.clientTranslator.GetVolume(ctx, c, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get volume in translation layer: %w", err)
	}

	log.Info().Str("resource_id", volID).Str("cluster_id", clusterID).Msgf("fetched volume")

	return resp, nil
}

func (s *Service) GetVolumes(ctx context.Context, req *storms.GetVolumesRequest) (*storms.GetVolumesResponse, error) {
	out := []*storms.Volume{}
	clusterIDs := s.clusterManager.AllIDs()
	for _, clusterID := range clusterIDs {
		c, err := s.clusterManager.Get(clusterID)
		if err != nil {
			return nil, fmt.Errorf("failed to get client: %w", err)
		}

		resp, err := s.clientTranslator.GetVolumes(ctx, c, req)
		if err != nil {
			return nil, fmt.Errorf("failed to get volumes in translation layer: %w", err)
		}
		out = append(out, resp.Volumes...)

		log.Info().Str("cluster_id", clusterID).Msgf("fetched volumes")
	}

	return &storms.GetVolumesResponse{
		Volumes: out,
	}, nil
}

func (s *Service) CreateVolume(ctx context.Context, req *storms.CreateVolumeRequest,
) (*storms.CreateVolumeResponse, error) {
	var clusterID string
	var err error
	switch source := req.GetSource().(type) {
	case *storms.CreateVolumeRequest_FromSnapshot:
		snapshotID := source.FromSnapshot.SnapshotUuid
		clusterID, err = s.resourceManager.GetResourceCluster(snapshotID)
		if err != nil {
			return nil, fmt.Errorf("failed to get cluster for resource: %w", err)
		}
	case *storms.CreateVolumeRequest_FromNew:
		clusterID, err = s.allocator.SelectClusterForNewResource()
		if err != nil {
			return nil, fmt.Errorf("failed to get allocate cluster for resource: %w", err)
		}
	}

	c, err := s.clusterManager.Get(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for cluster: %w", err)
	}

	resp, err := s.clientTranslator.CreateVolume(ctx, c, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume in translation layer: %w", err)
	}
	r := &resource.Resource{ID: req.Uuid, ClusterID: clusterID, ResourceType: resource.TypeVolume}
	err = s.resourceManager.Map(r)
	if err != nil {
		log.Warn().Str("resource_id", req.Uuid).Interface("err", err).Msg("failed to map resource")
	}

	log.Info().Str("cluster_id", clusterID).Str("resource_id", req.Uuid).Msg("created volume")

	return resp, nil
}

func (s *Service) ResizeVolume(ctx context.Context, req *storms.ResizeVolumeRequest,
) (*storms.ResizeVolumeResponse, error) {
	volID := req.GetUuid()
	clusterID, c, err := s.getClientForResource(volID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for resource: %w", err)
	}
	resp, err := s.clientTranslator.ResizeVolume(ctx, c, req)
	if err != nil {
		return nil, fmt.Errorf("failed to resize volume in translation layer: %w", err)
	}
	log.Info().Str("cluster_id", clusterID).Str("resource_id", req.Uuid).Msg("resized volume")

	return resp, nil
}

func (s *Service) DeleteVolume(ctx context.Context, req *storms.DeleteVolumeRequest,
) (*storms.DeleteVolumeResponse, error) {
	volID := req.GetUuid()
	clusterID, c, err := s.getClientForResource(volID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for resource: %w", err)
	}

	resp, err := s.clientTranslator.DeleteVolume(ctx, c, req)
	if err != nil {
		return nil, fmt.Errorf("failed to delete volume in translation layer: %w", err)
	}
	err = s.resourceManager.Unmap(req.GetUuid())
	if err != nil {
		log.Warn().Str("resource_id", req.Uuid).Interface("err", err).Msg("failed to unmap resource")
	}

	log.Info().Str("cluster_id", clusterID).Str("resource_id", req.Uuid).Msg("deleted volume")

	return resp, nil
}

func (s *Service) AttachVolume(ctx context.Context, req *storms.AttachVolumeRequest,
) (*storms.AttachVolumeResponse, error) {
	volID := req.GetUuid()
	clusterID, c, err := s.getClientForResource(volID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for resource: %w", err)
	}

	resp, err := s.clientTranslator.AttachVolume(ctx, c, req)
	if err != nil {
		return nil, fmt.Errorf("failed to attach volume in translation layer: %w", err)
	}

	log.Info().Str("cluster_id", clusterID).Str("resource_id", req.Uuid).Msg("attached volume")

	return resp, nil
}

func (s *Service) DetachVolume(ctx context.Context, req *storms.DetachVolumeRequest,
) (*storms.DetachVolumeResponse, error) {
	volID := req.GetUuid()
	clusterID, c, err := s.getClientForResource(volID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for resource: %w", err)
	}

	resp, err := s.clientTranslator.DetachVolume(ctx, c, req)
	if err != nil {
		return nil, fmt.Errorf("failed to detach volume in translation layer: %w", err)
	}

	log.Info().Str("cluster_id", clusterID).Str("resource_id", req.Uuid).Msg("detached volume")

	return resp, nil
}

func (s *Service) GetSnapshot(ctx context.Context, req *storms.GetSnapshotRequest,
) (*storms.GetSnapshotResponse, error) {
	snapshotID := req.GetUuid()
	clusterID, c, err := s.getClientForResource(snapshotID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for resource: %w", err)
	}

	resp, err := s.clientTranslator.GetSnapshot(ctx, c, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot translation layer: %w", err)
	}

	log.Info().Str("cluster_id", clusterID).Str("resource_id", req.Uuid).Msg("fetched snapshot")

	return resp, nil
}

func (s *Service) GetSnapshots(ctx context.Context, req *storms.GetSnapshotsRequest,
) (*storms.GetSnapshotsResponse, error) {
	out := []*storms.Snapshot{}

	clusterIDs := s.clusterManager.AllIDs()
	for _, clusterID := range clusterIDs {
		c, err := s.clusterManager.Get(clusterID)
		if err != nil {
			return nil, fmt.Errorf("failed to get client: %w", err)
		}

		snapshots, err := s.clientTranslator.GetSnapshots(ctx, c, req)
		if err != nil {
			return nil, fmt.Errorf("faild to get snapshots in translation layer: %w", err)
		}
		out = append(out, snapshots.Snapshots...)

		log.Info().Str("cluster_id", clusterID).Msgf("fetched snapshots")
	}

	return &storms.GetSnapshotsResponse{
		Snapshots: out,
	}, nil
}

func (s *Service) CreateSnapshot(ctx context.Context, req *storms.CreateSnapshotRequest,
) (*storms.CreateSnapshotResponse, error) {
	clusterID, err := s.resourceManager.GetResourceCluster(req.GetSrcVolumeUuid())
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster for resource: %w", err)
	}
	c, err := s.clusterManager.Get(clusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for cluster: %w", err)
	}

	resp, err := s.clientTranslator.CreateSnapshot(ctx, c, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot in translation layer: %w", err)
	}
	r := &resource.Resource{ID: req.Uuid, ClusterID: clusterID, ResourceType: resource.TypeSnapshot}
	err = s.resourceManager.Map(r)
	if err != nil {
		log.Warn().Str("resource_id", req.Uuid).Interface("err", err).Msg("failed to map resource")
	}

	log.Info().Str("cluster_id", clusterID).Str("resource_id", req.Uuid).Msg("created snapshot")

	return resp, nil
}

func (s *Service) DeleteSnapshot(ctx context.Context, req *storms.DeleteSnapshotRequest,
) (*storms.DeleteSnapshotResponse, error) {
	snapshotID := req.GetUuid()
	log.Info().Msgf("Mapping snapshot [id=%s] to client", snapshotID)
	clusterID, c, err := s.getClientForResource(snapshotID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for resource: %w", err)
	}

	resp, err := s.clientTranslator.DeleteSnapshot(ctx, c, req)
	if err != nil {
		return nil, fmt.Errorf("failed to delete volume in translation layer: %w", err)
	}
	err = s.resourceManager.Unmap(req.GetUuid())
	if err != nil {
		log.Warn().Str("resource_id", req.Uuid).Interface("err", err).Msg("failed to unmap resource")
	}

	log.Info().Str("cluster_id", clusterID).Str("resource_id", req.Uuid).Msg("deleted snapshot")

	return resp, nil
}

func (s *Service) SyncResource(ctx context.Context, req *storms.SyncResourceRequest,
) (*storms.SyncResourceResponse, error) {
	r := &resource.Resource{ID: req.Uuid, ClusterID: req.ClusterUuid}
	if req.ResourceType == storms.ResourceType_RESOURCE_TYPE_SNAPSHOT {
		r.ResourceType = resource.TypeSnapshot
	} else if req.ResourceType == storms.ResourceType_RESOURCE_TYPE_VOLUME {
		r.ResourceType = resource.TypeVolume
	}
	err := s.resourceManager.Map(r)
	if err != nil {
		log.Warn().Str("resource_id", req.Uuid).Interface("err", err).Msg("failed to map resource")
	}
	rm, err := s.syncResourceHelper(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to sync resource: %w", err)
	}

	if rm {
		err := s.resourceManager.Unmap(req.Uuid)
		if err != nil {
			log.Warn().Str("resource_id", req.Uuid).Interface("err", err).Msg("failed to unmap resource")
		}
	}

	return &storms.SyncResourceResponse{}, nil
}

func (s *Service) syncResourceHelper(ctx context.Context, req *storms.SyncResourceRequest,
) (bool, error) {
	switch req.ResourceType {
	case storms.ResourceType_RESOURCE_TYPE_VOLUME:
		getVolReq := &storms.GetVolumeRequest{
			Uuid: req.Uuid,
		}
		getVolResp, err := s.GetVolume(ctx, getVolReq)
		if err != nil {
			return true, fmt.Errorf("failed to get volume: %w", err)
		}
		if getVolResp.Volume != nil {
			return false, nil
		}

	case storms.ResourceType_RESOURCE_TYPE_SNAPSHOT:
		getSnapshotReq := &storms.GetSnapshotRequest{
			Uuid: req.Uuid,
		}
		getSnapshotResp, err := s.GetSnapshot(ctx, getSnapshotReq)
		if err != nil {
			return true, fmt.Errorf("failed to get volume: %w", err)
		}
		if getSnapshotResp.Snapshot != nil {
			return false, nil
		}

	case storms.ResourceType_RESOURCE_TYPE_UNSPECIFIED:
		return true, errUnspecifiedResource
	}

	return true, errUnexpected
}

func (s *Service) SyncAllResources(_ context.Context, _ *storms.SyncAllResourcesRequest,
) (*storms.SyncAllResourcesResponse, error) {
	s.syncResourceManager()
	log.Info().Msgf("Syncing metadata of clusters")

	return &storms.SyncAllResourcesResponse{}, nil
}
