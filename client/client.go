package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"

	"gitlab.com/crusoeenergy/island/storage/storms/client/models"
	"gitlab.com/crusoeenergy/island/storage/storms/client/vendors/krusoe"
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
}

//nolint:cyclop // multipliex function
func NewClient(vendor string, cfg map[string]interface{}) (Client, error) {
	cfgBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed ot marshal client config: %w", err)
	}

	switch vendor {
	case "lightbits":
		var cfg lightbits.ClientConfig
		if err := lightbits.ParseConfig(cfgBytes, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse Krusoe config: %w", err)
		}

		log.Info().Msgf("Creating new Lightbits client: %#v", cfg)

		clientAdapter, err := lightbits.NewClientAdapter(&cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create new lightbits client adapter: %w", err)
		}

		return clientAdapter, nil

	case "purestorage":
		var cfg purestorage.ClientConfig
		if err := purestorage.ParseConfig(cfgBytes, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse PureStorage config: %w", err)
		}
		log.Info().Msgf("Creating new PureStorage client: %#v", cfg)

		client, err := purestorage.NewClient(&cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create new purestorage client adapter: %w", err)
		}

		return client, nil

	case "krusoe":
		var cfg krusoe.Config
		if err := krusoe.ParseConfig(cfgBytes, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse Krusoe config: %w", err)
		}

		client := krusoe.NewClient(cfg)

		return client, nil

	default:
		return nil, errUnsupportedVendor
	}
}
