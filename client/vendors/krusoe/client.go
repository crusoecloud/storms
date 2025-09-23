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
			SourceSnapshotUUID: "",
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
			SourceSnapshotUUID: "",
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
	v, err := c.backend.createVolume(c.apiKey, req.UUID, uint(req.Size), uint(req.SectorSize))
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
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
		SourceSnapshotUUID: "",
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
		return nil, fmt.Errorf("failed to resize volume: %w", err)
	}

	return &models.AttachVolumeResponse{}, nil
}

func (c *Client) DetachVolume(_ context.Context, req *models.DetachVolumeRequest,
) (*models.DetachVolumeResponse, error) {
	_, err := c.backend.detachVolume(c.apiKey, req.UUID)
	if err != nil {
		return nil, fmt.Errorf("failed to resize volume: %w", err)
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
		SourceVolumeUUID: "",
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
	out := lo.Map[*Snapshot, *models.Snapshot](ss, func(v *Snapshot, _ int) *models.Snapshot {
		sectorSize, err := uintToUint32Checked(v.sectorSize)
		if err != nil {
			mapErr = multierr.Append(mapErr, err)

			return nil
		}

		return &models.Snapshot{
			UUID:             v.id,
			Size:             uint64(v.size),
			SectorSize:       sectorSize,
			IsAvailable:      true,
			SourceVolumeUUID: "",
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

func (c *Client) CloneVolume(_ context.Context, req *models.CloneVolumeRequest) (*models.CloneVolumeResponse, error) {
	srcSnapshotID := req.SrcSnapshotUUID

	srcSnapshot, err := c.backend.getSnapshot(c.apiKey, srcSnapshotID)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	v, err := c.backend.cloneVolume(c.apiKey, srcSnapshot.sourceVolumeID, req.DstVolumeUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to clone volume: %w", err)
	}

	return &models.CloneVolumeResponse{
		OperationID: v.id,
	}, nil
}

func (c *Client) CloneSnapshot(_ context.Context, req *models.CloneSnapshotRequest,
) (*models.CloneSnapshotResponse, error) {
	s, err := c.backend.cloneSnapshot(c.apiKey, req.SrcSnaphotUUID, req.DstSnapshotUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to clone snapshot: %w", err)
	}

	return &models.CloneSnapshotResponse{
		OperationID: s.id,
	}, nil
}

func (c *Client) GetCloneStatus(_ context.Context, _ *models.GetCloneStatusRequest,
) (*models.GetCloneStatusResponse, error) {
	err := c.backend.getCloneStatus(c.apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get clone status: %w", err)
	}

	return &models.GetCloneStatusResponse{
		OperationID: "",
	}, nil
}

var errUint32OutOfRange = errors.New("uint32 out of range")

func uintToUint32Checked(u uint) (uint32, error) {
	if u > math.MaxUint32 {
		return 0, errUint32OutOfRange
	}

	return uint32(u), nil
}
