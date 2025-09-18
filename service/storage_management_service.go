package service

import (
	"context"

	"github.com/rs/zerolog/log"

	storms "gitlab.com/crusoeenergy/island/storage/storms/api/gen/go/storms/v1"
)

func (s *Service) GetVolume(_ context.Context, _ *storms.GetVolumeRequest,
) (*storms.GetVolumeResponse, error) {
	log.Info().Msg("called GetVolume endpoint")

	return &storms.GetVolumeResponse{}, nil
}

func (s *Service) CreateVolume(_ context.Context, _ *storms.CreateVolumeRequest,
) (*storms.CreateVolumeResponse, error) {
	log.Info().Msg("called CreateVolume endpoint")

	return &storms.CreateVolumeResponse{}, nil
}

func (s *Service) ResizeVolume(_ context.Context, _ *storms.ResizeVolumeRequest,
) (*storms.ResizeVolumeResponse, error) {
	log.Info().Msg("called ResizeVolume endpoint")

	return &storms.ResizeVolumeResponse{}, nil
}

func (s *Service) DeleteVolume(_ context.Context, _ *storms.DeleteVolumeRequest,
) (*storms.DeleteVolumeResponse, error) {
	log.Info().Msg("called DeleteVolume endpoint")

	return &storms.DeleteVolumeResponse{}, nil
}

func (s *Service) AttachVolume(_ context.Context, _ *storms.AttachVolumeRequest,
) (*storms.AttachVolumeResponse, error) {
	log.Info().Msg("called AttachVolume endpoint")

	return &storms.AttachVolumeResponse{}, nil
}

func (s *Service) DetachVolume(_ context.Context, _ *storms.DetachVolumeRequest,
) (*storms.DetachVolumeResponse, error) {
	log.Info().Msg("called DetachVolume endpoint")

	return &storms.DetachVolumeResponse{}, nil
}

func (s *Service) GetSnapshot(_ context.Context, _ *storms.GetSnapshotRequest,
) (*storms.GetSnapshotResponse, error) {
	log.Info().Msg("called GetSnapshot endpoint")

	return &storms.GetSnapshotResponse{}, nil
}
