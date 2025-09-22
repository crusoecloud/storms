//nolint:err113,revive,lll,golines // skeleton code
package purestorage

import (
	"context"
	"fmt"

	"gitlab.com/crusoeenergy/island/storage/storms/client/models"
)

type Client struct {
	/*
		TODO - this can be a http.Client
	*/
}

func NewClient(cfg ClientConfig) (*Client, error) {
	return &Client{
		// TODO - add stuff here
	}, nil
}

func (c *Client) GetVolume(ctx context.Context, req *models.GetVolumeRequest) (*models.GetVolumeResponse, error) {
	// TODO: make API call to cluster

	return nil, fmt.Errorf("not implemented")
}

func (c *Client) GetVolumes(ctx context.Context, req *models.GetVolumesRequest) (*models.GetVolumesResponse, error) {
	// TODO: make API call to cluster

	return nil, fmt.Errorf("not implemented")
}

func (c *Client) CreateVolume(ctx context.Context, req *models.CreateVolumeRequest) (*models.CreateVolumeResponse, error) {
	// TODO: make API call to cluster

	return nil, fmt.Errorf("not implemented")
}

func (c *Client) ResizeVolume(ctx context.Context, req *models.ResizeVolumeRequest) (*models.ResizeVolumeResponse, error) {
	// TODO: make API call to cluster

	return nil, fmt.Errorf("not implemented")
}

func (c *Client) DeleteVolume(ctx context.Context, req *models.DeleteVolumeRequest) (*models.DeleteVolumeResponse, error) {
	// TODO: make API call to cluster

	return nil, fmt.Errorf("not implemented")
}

func (c *Client) AttachVolume(ctx context.Context, req *models.AttachVolumeRequest) (*models.AttachVolumeResponse, error) {
	// TODO: make API call to cluster

	return nil, fmt.Errorf("not implemented")
}

func (c *Client) DetachVolume(ctx context.Context, req *models.DetachVolumeRequest) (*models.DetachVolumeResponse, error) {
	// TODO: make API call to cluster

	return nil, fmt.Errorf("not implemented")
}

func (c *Client) GetSnapshot(ctx context.Context, req *models.GetSnapshotRequest) (*models.GetSnapshotResponse, error) {
	// TODO: make API call to cluster

	return nil, fmt.Errorf("not implemented")
}

func (c *Client) GetSnapshots(ctx context.Context, req *models.GetSnapshotsRequest) (*models.GetSnapshotsResponse, error) {
	// TODO: make API call to cluster

	return nil, fmt.Errorf("not implemented")
}

func (c *Client) CreateSnapshot(ctx context.Context, req *models.CreateSnapshotRequest) (*models.CreateSnapshotResponse, error) {
	// TODO: make API call to cluster

	return nil, fmt.Errorf("not implemented")
}

func (c *Client) DeleteSnapshot(ctx context.Context, req *models.DeleteSnapshotRequest) (*models.DeleteSnapshotResponse, error) {
	// TODO: make API call to cluster

	return nil, fmt.Errorf("not implemented")
}

func (c *Client) CloneVolume(ctx context.Context, req *models.CloneVolumeRequest) (*models.CloneVolumeResponse, error) {
	// TODO: make API call to cluster

	return nil, fmt.Errorf("not implemented")
}

func (c *Client) CloneSnapshot(ctx context.Context, req *models.CloneSnapshotRequest) (*models.CloneSnapshotResponse, error) {
	// TODO: make API call to cluster

	return nil, fmt.Errorf("not implemented")
}

func (c *Client) GetCloneStatus(ctx context.Context, req *models.GetCloneStatusRequest) (*models.GetCloneStatusResponse, error) {
	// TODO: make API call to cluster

	return nil, fmt.Errorf("not implemented")
}
