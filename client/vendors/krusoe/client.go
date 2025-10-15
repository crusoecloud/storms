package krusoe

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/samber/lo"
	"go.uber.org/multierr"

	"gitlab.com/crusoeenergy/island/storage/storms/client/models"
)

var errUnsupportVolumeSource = errors.New("unsupport volume source")

type Client struct {
	apiKey  string
	backend *backend
}

func NewClient(cfg Config) *Client {
	return &Client{
		apiKey:  cfg.APIKey,
		backend: newBackend(),
	}
}

func (c *Client) GetVolume(_ context.Context, req *models.GetVolumeRequest) (*models.GetVolumeResponse, error) {
	v, err := c.backend.getVolume(c.apiKey, req.UUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get volume: %w", err)
	}

	sectorSize, err := uintToUint32Checked(v.sectorSize)
	if err != nil {
		return nil, fmt.Errorf("failed to convert uint to uint32: %w", err)
	}

	return &models.GetVolumeResponse{
		Volume: &models.Volume{
			UUID:               v.id,
			SectorSize:         sectorSize,
			Size:               uint64(v.size),
			Acls:               v.acls,
			IsAvailable:        true,
			SourceSnapshotUUID: v.srcSnapshotID,
		},
	}, nil
}

func (c *Client) GetVolumes(_ context.Context, _ *models.GetVolumesRequest) (*models.GetVolumesResponse, error) {
	vs, err := c.backend.getVolumes(c.apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get volumes: %w", err)
	}

	var mapErr error
	out := lo.Map[*Volume, *models.Volume](vs, func(v *Volume, _ int) *models.Volume {
		sectorSize, err := uintToUint32Checked(v.sectorSize)
		if err != nil {
			mapErr = multierr.Append(mapErr, err)

			return nil
		}

		return &models.Volume{
			UUID:               v.id,
			SectorSize:         sectorSize,
			Size:               uint64(v.size),
			Acls:               v.acls,
			IsAvailable:        true,
			SourceSnapshotUUID: v.srcSnapshotID,
		}
	})

	if mapErr != nil {
		return nil, fmt.Errorf("failed to translate volume attributes: %w", mapErr)
	}

	return &models.GetVolumesResponse{
		Volumes: out,
	}, nil
}

func (c *Client) CreateVolume(_ context.Context, req *models.CreateVolumeRequest,
) (*models.CreateVolumeResponse, error) {
	var v *Volume
	var err error

	switch source := req.Source.(type) {
	case models.NewVolumeSpec:
		v, err = c.backend.createNewVolume(c.apiKey, req.UUID, uint(source.Size), uint(source.SectorSize))
		if err != nil {
			return nil, fmt.Errorf("failed to create new volume: %w", err)
		}
	case models.SnapshotSource:
		v, err = c.backend.createVolumeFromSnapshot(c.apiKey, req.UUID, source.SnapshotUUID)
		if err != nil {
			return nil, fmt.Errorf("failed to create volume from snapshot: %w", err)
		}
	default:
		return nil, errUnsupportVolumeSource
	}

	sectorSize, err := uintToUint32Checked(v.sectorSize)
	if err != nil {
		return nil, fmt.Errorf("failed to convert uint to uint32: %w", err)
	}

	out := &models.Volume{
		UUID:               v.name,
		Size:               uint64(v.size),
		SectorSize:         sectorSize,
		Acls:               v.acls,
		IsAvailable:        true,
		SourceSnapshotUUID: v.srcSnapshotID,
	}

	return &models.CreateVolumeResponse{
		Volume: out,
	}, nil
}

func (c *Client) ResizeVolume(_ context.Context, req *models.ResizeVolumeRequest,
) (*models.ResizeVolumeResponse, error) {
	_, err := c.backend.resizeVolume(c.apiKey, req.UUID, uint(req.Size))
	if err != nil {
		return nil, fmt.Errorf("failed to resize volume: %w", err)
	}

	return &models.ResizeVolumeResponse{}, nil
}

func (c *Client) DeleteVolume(_ context.Context, req *models.DeleteVolumeRequest,
) (*models.DeleteVolumeResponse, error) {
	err := c.backend.deleteVolume(c.apiKey, req.UUID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete volume: %w", err)
	}

	return &models.DeleteVolumeResponse{}, nil
}

func (c *Client) AttachVolume(_ context.Context, req *models.AttachVolumeRequest,
) (*models.AttachVolumeResponse, error) {
	_, err := c.backend.attachVolume(c.apiKey, req.UUID, req.Acls)
	if err != nil {
		return nil, fmt.Errorf("failed to attach volume: %w", err)
	}

	return &models.AttachVolumeResponse{}, nil
}

func (c *Client) DetachVolume(_ context.Context, req *models.DetachVolumeRequest,
) (*models.DetachVolumeResponse, error) {
	_, err := c.backend.detachVolume(c.apiKey, req.UUID)
	if err != nil {
		return nil, fmt.Errorf("failed to detach volume: %w", err)
	}

	return &models.DetachVolumeResponse{}, nil
}

func (c *Client) GetSnapshot(_ context.Context, req *models.GetSnapshotRequest) (*models.GetSnapshotResponse, error) {
	s, err := c.backend.getSnapshot(c.apiKey, req.UUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	sectorSize, err := uintToUint32Checked(s.sectorSize)
	if err != nil {
		return nil, fmt.Errorf("failed to convert uint to uint32: %w", err)
	}

	out := &models.Snapshot{
		UUID:             s.id,
		SectorSize:       sectorSize,
		Size:             uint64(s.size),
		IsAvailable:      true,
		SourceVolumeUUID: s.sourceVolumeID,
	}

	return &models.GetSnapshotResponse{
		Snapshot: out,
	}, nil
}

func (c *Client) GetSnapshots(_ context.Context, _ *models.GetSnapshotsRequest) (*models.GetSnapshotsResponse, error) {
	ss, err := c.backend.getSnapshots(c.apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots: %w", err)
	}

	var mapErr error
	out := lo.Map[*Snapshot, *models.Snapshot](ss, func(s *Snapshot, _ int) *models.Snapshot {
		sectorSize, err := uintToUint32Checked(s.sectorSize)
		if err != nil {
			mapErr = multierr.Append(mapErr, err)

			return nil
		}

		return &models.Snapshot{
			UUID:             s.id,
			Size:             uint64(s.size),
			SectorSize:       sectorSize,
			IsAvailable:      true,
			SourceVolumeUUID: s.sourceVolumeID,
		}
	})

	if mapErr != nil {
		return nil, fmt.Errorf("failed to translate snapshot attributes type: %w ", mapErr)
	}

	return &models.GetSnapshotsResponse{
		Snapshots: out,
	}, nil
}

func (c *Client) CreateSnapshot(_ context.Context, req *models.CreateSnapshotRequest,
) (*models.CreateSnapshotResponse, error) {
	v, err := c.backend.createSnapshot(c.apiKey, req.UUID, req.SourceVolumeUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	sectorSize, err := uintToUint32Checked(v.sectorSize)
	if err != nil {
		return nil, fmt.Errorf("failed to convert uint to uint32: %w", err)
	}

	out := &models.Snapshot{
		UUID:             v.name,
		Size:             uint64(v.size),
		SectorSize:       sectorSize,
		IsAvailable:      true,
		SourceVolumeUUID: req.SourceVolumeUUID,
	}

	return &models.CreateSnapshotResponse{
		Snapshot: out,
	}, nil
}

func (c *Client) DeleteSnapshot(_ context.Context, req *models.DeleteSnapshotRequest,
) (*models.DeleteSnapshotResponse, error) {
	err := c.backend.deleteSnapshot(c.apiKey, req.UUID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete snapshot: %w", err)
	}

	return &models.DeleteSnapshotResponse{}, nil
}

var errUint32OutOfRange = errors.New("uint32 out of range")

func uintToUint32Checked(u uint) (uint32, error) {
	if u > math.MaxUint32 {
		return 0, errUint32OutOfRange
	}

	return uint32(u), nil
}
