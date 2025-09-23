package service

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	storms "gitlab.com/crusoeenergy/island/storage/storms/api/gen/go/storms/v1"
)

func (s *Service) GetVolume(ctx context.Context, req *storms.GetVolumeRequest,
) (*storms.GetVolumeResponse, error) {
	volID := req.GetUuid()
	log.Info().Msgf("Mapping volume [id=%s] to client", volID)
	clientID, c, err := s.ResourceManager.getClientForResource(volID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for resource: %w", err)
	}

	log.Info().Msgf("GetVolume request [id=%s] routed to client [id=%s]", volID, clientID)
	resp, err := s.ClientTranslator.GetVolume(ctx, c, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get volume in translation layer: %w", err)
	}

	return resp, nil
}

func (s *Service) GetVolumes(ctx context.Context, req *storms.GetVolumesRequest) (*storms.GetVolumesResponse, error) {
	out := []*storms.Volume{}

	log.Info().Msgf("GetVolumes request routed to all clients")
	clientIDs := s.ResourceManager.getAllClientIDs()
	for _, clientID := range clientIDs {
		c, err := s.ResourceManager.getClient(clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to get client: %w", err)
		}

		resp, err := s.ClientTranslator.GetVolumes(ctx, c, req)
		if err != nil {
			return nil, fmt.Errorf("failed to get volumes in translation layer: %w", err)
		}
		out = append(out, resp.Volumes...)
	}

	return &storms.GetVolumesResponse{
		Volumes: out,
	}, nil
}

func (s *Service) CreateVolume(ctx context.Context, req *storms.CreateVolumeRequest,
) (*storms.CreateVolumeResponse, error) {
	clientID, c, err := s.ResourceManager.allocateClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	log.Info().Msgf("create volume grpc endpoint req: %v", req)

	// TODO - we may want to block creation of volume with the same UUID

	log.Info().Msgf("CreateVolume request routed to client [id=%s]", clientID)
	resp, err := s.ClientTranslator.CreateVolume(ctx, c, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume in translation layer: %w", err)
	}

	log.Info().Msgf("Resource added to StorMS [id=%s, clientID=%s]", req.Volume.Uuid, clientID)
	err = s.ResourceManager.addResource(req.Volume.Uuid, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to add resource: %w", err)
	}

	return resp, nil
}

func (s *Service) ResizeVolume(ctx context.Context, req *storms.ResizeVolumeRequest,
) (*storms.ResizeVolumeResponse, error) {
	volID := req.GetUuid()
	log.Info().Msgf("Mapping volume [id=%s] to client", volID)
	clientID, c, err := s.ResourceManager.getClientForResource(volID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for resource: %w", err)
	}

	log.Info().Msgf("ResizeVolume request routed to client [id=%s]", clientID)
	resp, err := s.ClientTranslator.ResizeVolume(ctx, c, req)
	if err != nil {
		return nil, fmt.Errorf("failed to resize volume in translation layer: %w", err)
	}

	return resp, nil
}

func (s *Service) DeleteVolume(ctx context.Context, req *storms.DeleteVolumeRequest,
) (*storms.DeleteVolumeResponse, error) {
	volID := req.GetUuid()
	log.Info().Msgf("Mapping volume [id=%s] to client", volID)
	clientID, c, err := s.ResourceManager.getClientForResource(volID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for resource: %w", err)
	}

	log.Info().Msgf("DeleteVolume request routed to client [id=%s]", clientID)
	resp, err := s.ClientTranslator.DeleteVolume(ctx, c, req)
	if err != nil {
		return nil, fmt.Errorf("failed to delete volume in translation layer: %w", err)
	}

	log.Info().Msgf("Resource removed from StorMS [id=%s]", req.GetUuid())
	err = s.ResourceManager.removeResource(req.GetUuid())
	if err != nil {
		return nil, fmt.Errorf("failed to remove resource: %w", err)
	}

	return resp, nil
}

func (s *Service) AttachVolume(ctx context.Context, req *storms.AttachVolumeRequest,
) (*storms.AttachVolumeResponse, error) {
	volID := req.GetUuid()
	log.Info().Msgf("Mapping volume [id=%s] to client", volID)
	clientID, c, err := s.ResourceManager.getClientForResource(volID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for resource: %w", err)
	}

	log.Info().Msgf("AttachVolume request routed to client [id=%s]", clientID)
	resp, err := s.ClientTranslator.AttachVolume(ctx, c, req)
	if err != nil {
		return nil, fmt.Errorf("failed to attach volume in translation layer: %w", err)
	}

	return resp, nil
}

func (s *Service) DetachVolume(ctx context.Context, req *storms.DetachVolumeRequest,
) (*storms.DetachVolumeResponse, error) {
	volID := req.GetUuid()
	log.Info().Msgf("Mapping volume [id=%s] to client", volID)
	clientID, c, err := s.ResourceManager.getClientForResource(volID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for resource: %w", err)
	}

	log.Info().Msgf("DetachVolume request routed to client [id=%s]", clientID)
	resp, err := s.ClientTranslator.DetachVolume(ctx, c, req)
	if err != nil {
		return nil, fmt.Errorf("failed to detach volume in translation layer: %w", err)
	}

	return resp, nil
}

func (s *Service) GetSnapshot(ctx context.Context, req *storms.GetSnapshotRequest,
) (*storms.GetSnapshotResponse, error) {
	snapshotID := req.GetUuid()
	log.Info().Msgf("Mapping volume [id=%s] to client", snapshotID)
	clientID, c, err := s.ResourceManager.getClientForResource(snapshotID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for resource: %w", err)
	}

	log.Info().Msgf("GetSnapshot request routed to client [id=%s]", clientID)
	resp, err := s.ClientTranslator.GetSnapshot(ctx, c, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot translation layer: %w", err)
	}

	return resp, nil
}

func (s *Service) GetSnapshots(ctx context.Context, req *storms.GetSnapshotsRequest,
) (*storms.GetSnapshotsResponse, error) {
	out := []*storms.Snapshot{}

	log.Info().Msgf("GetVolumes request routed to all clients")
	clientIDs := s.ResourceManager.getAllClientIDs()
	for _, clientID := range clientIDs {
		c, err := s.ResourceManager.getClient(clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to get client: %w", err)
		}

		snapshots, err := s.ClientTranslator.GetSnapshots(ctx, c, req)
		if err != nil {
			return nil, fmt.Errorf("faield to get snapshots in translation layer: %w", err)
		}
		out = append(out, snapshots.Snapshots...)
	}

	return &storms.GetSnapshotsResponse{
		Snapshots: out,
	}, nil
}

func (s *Service) CreateSnapshot(ctx context.Context, req *storms.CreateSnapshotRequest,
) (*storms.CreateSnapshotResponse, error) {
	clientID, c, err := s.ResourceManager.allocateClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	log.Info().Msgf("CreateSnapshot request routed to client [id=%s]", clientID)
	resp, err := s.ClientTranslator.CreateSnapshot(ctx, c, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot in translation layer: %w", err)
	}

	log.Info().Msgf("Resource added to StorMS [id=%s, clientID=%s]", req.Snapshot.Uuid, clientID)
	err = s.ResourceManager.addResource(req.Snapshot.Uuid, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to add resource: %w", err)
	}

	return resp, nil
}

func (s *Service) DeleteSnapshot(ctx context.Context, req *storms.DeleteSnapshotRequest,
) (*storms.DeleteSnapshotResponse, error) {
	snapshotID := req.GetUuid()

	log.Info().Msgf("Mapping volume [id=%s] to client", snapshotID)
	clientID, c, err := s.ResourceManager.getClientForResource(snapshotID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client for resource: %w", err)
	}

	log.Info().Msgf("DeleteSnapshot request routed to client [id=%s]", clientID)
	resp, err := s.ClientTranslator.DeleteSnapshot(ctx, c, req)
	if err != nil {
		return nil, fmt.Errorf("failed to delete volume in translation layer: %w", err)
	}

	log.Info().Msgf("Resource removed from StorMS [id=%s]", req.GetUuid())
	err = s.ResourceManager.removeResource(req.GetUuid())
	if err != nil {
		return nil, fmt.Errorf("failed to remove snapshot: %w", err)
	}

	return resp, nil
}

func (s *Service) CloneVolume(_ context.Context, _ *storms.CloneVolumeRequest) (*storms.CloneVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CloneVolume not implemented")
}

func (s *Service) CloneSnapshot(_ context.Context, _ *storms.CloneSnapshotRequest,
) (*storms.CloneSnapshotResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CloneSnapshot not implemented")
}

func (s *Service) GetCloneStatus(_ context.Context, _ *storms.GetCloneStatusRequest,
) (*storms.GetCloneStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCloneStatus not implemented")
}
