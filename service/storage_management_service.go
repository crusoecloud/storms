package service

import (
	"context"

	"github.com/rs/zerolog/log"

	"gitlab.com/crusoeenergy/schemas/api/island/v2/storagemulti"
)

func (s *Service) GetVolume(_ context.Context, _ *storagemulti.GetVolumeRequest,
) (*storagemulti.GetVolumeResponse, error) {
	log.Info().Msg("called GetVolume endpoint")

	return &storagemulti.GetVolumeResponse{}, nil
}

func (s *Service) CreateVolume(_ context.Context, _ *storagemulti.CreateVolumeRequest,
) (*storagemulti.CreateVolumeResponse, error) {
	log.Info().Msg("called CreateVolume endpoint")

	return &storagemulti.CreateVolumeResponse{}, nil
}

func (s *Service) ResizeVolume(_ context.Context, _ *storagemulti.ResizeVolumeRequest,
) (*storagemulti.ResizeVolumeResponse, error) {
	log.Info().Msg("called ResizeVolume endpoint")

	return &storagemulti.ResizeVolumeResponse{}, nil
}

func (s *Service) DeleteVolume(_ context.Context, _ *storagemulti.DeleteVolumeRequest,
) (*storagemulti.DeleteVolumeResponse, error) {
	log.Info().Msg("called DeleteVolume endpoint")

	return &storagemulti.DeleteVolumeResponse{}, nil
}

func (s *Service) AttachVolume(_ context.Context, _ *storagemulti.AttachVolumeRequest,
) (*storagemulti.AttachVolumeResponse, error) {
	log.Info().Msg("called AttachVolume endpoint")

	return &storagemulti.AttachVolumeResponse{}, nil
}

func (s *Service) DetachVolume(_ context.Context, _ *storagemulti.DetachVolumeRequest,
) (*storagemulti.DetachVolumeResponse, error) {
	log.Info().Msg("called DetachVolume endpoint")

	return &storagemulti.DetachVolumeResponse{}, nil
}

func (s *Service) GetSnapshot(_ context.Context, _ *storagemulti.GetSnapshotRequest,
) (*storagemulti.GetSnapshotResponse, error) {
	log.Info().Msg("called GetSnapshot endpoint")

	return &storagemulti.GetSnapshotResponse{}, nil
}
