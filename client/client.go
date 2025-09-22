package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"

	"gitlab.com/crusoeenergy/island/storage/storms/client/models"
	"gitlab.com/crusoeenergy/island/storage/storms/client/vendors/lightbits"
	"gitlab.com/crusoeenergy/island/storage/storms/client/vendors/purestorage"
)

var errUnsupportedVendor = errors.New("unsupported vendor")

type Client interface {
	// Volume operations
	GetVolume(ctx context.Context, req *models.GetVolumeRequest) (*models.GetVolumeResponse, error)
	GetVolumes(ctx context.Context, req *models.GetVolumesRequest) (*models.GetVolumesResponse, error)
	CreateVolume(ctx context.Context, req *models.CreateVolumeRequest) (*models.CreateVolumeResponse, error)
	ResizeVolume(ctx context.Context, req *models.ResizeVolumeRequest) (*models.ResizeVolumeResponse, error)
	DeleteVolume(ctx context.Context, req *models.DeleteVolumeRequest) (*models.DeleteVolumeResponse, error)
	AttachVolume(ctx context.Context, req *models.AttachVolumeRequest) (*models.AttachVolumeResponse, error)
	DetachVolume(ctx context.Context, req *models.DetachVolumeRequest) (*models.DetachVolumeResponse, error)

	// Snapshot operations
	GetSnapshot(ctx context.Context, req *models.GetSnapshotRequest) (*models.GetSnapshotResponse, error)
	GetSnapshots(ctx context.Context, req *models.GetSnapshotsRequest) (*models.GetSnapshotsResponse, error)
	CreateSnapshot(ctx context.Context, req *models.CreateSnapshotRequest) (*models.CreateSnapshotResponse, error)
	DeleteSnapshot(ctx context.Context, req *models.DeleteSnapshotRequest) (*models.DeleteSnapshotResponse, error)

	// Cloning operations
	CloneVolume(ctx context.Context, req *models.CloneVolumeRequest) (*models.CloneVolumeResponse, error)
	CloneSnapshot(ctx context.Context, req *models.CloneSnapshotRequest) (*models.CloneSnapshotResponse, error)
	GetCloneStatus(ctx context.Context, req *models.GetCloneStatusRequest) (*models.GetCloneStatusResponse, error)
}

//nolint:ireturn // need to return interface to support generic type
func NewClient(vendor string, cfg map[string]interface{}) (Client, error) {
	cfgBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed ot marshal client config: %w", err)
	}

	switch vendor {
	case "lightbits":
		var cfg lightbits.ClientConfig
		if err := yaml.Unmarshal(cfgBytes, &cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal lightbits cluster config: %w", err)
		}
		log.Info().Msgf("Creating new Lightbits client: %#v", cfg)

		clientAdapter, err := lightbits.NewClientAdapter(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create new lightbits client adapter: %w", err)
		}

		return clientAdapter, nil
	case "purestorage":
		var cfg purestorage.ClientConfig
		if err := yaml.Unmarshal(cfgBytes, &cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal purestorage cluster config: %w", err)
		}
		log.Info().Msgf("Creating new PureStorage client: %#v", cfg)

		client, err := purestorage.NewClient(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create new purestorage client adapter: %w", err)
		}

		return client, nil
	default:
		return nil, errUnsupportedVendor
	}
}
