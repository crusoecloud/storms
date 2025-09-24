package lightbits

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"gitlab.com/crusoeenergy/island/storage/storms/client/models"
	"gitlab.com/crusoeenergy/island/storage/storms/client/vendors/lightbits/dms"
)

var errMustHaveOneACL = errors.New("must have exactly 1 ACL")

// The Lightbits adapter is a wrapper around the Lightbits client that translate generic federation-level
// requests to Lightbits-specific API requests.
type ClientAdapter struct {
	client    *Client
	dmsClient *dms.Client
	// DMS with loadbalancer -> x2 cients
}

func NewClientAdapter(cfg ClientConfig) (*ClientAdapter, error) {
	c, err := NewClientV2(cfg)
	if err != nil {
		return nil, fmt.Errorf("faied to create new lightbits client: %w", err)
	}

	dmsClient, err := dms.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create new lightbits dms client: %w", err)
	}

	return &ClientAdapter{
		client:    c,
		dmsClient: dmsClient,
	}, nil
}

func (a *ClientAdapter) GetVolume(_ context.Context, req *models.GetVolumeRequest,
) (*models.GetVolumeResponse, error) {
	lbResp, err := a.client.GetVolume(req.UUID) // Note: volume UUID is lightbits volume name
	if err != nil {
		return nil, fmt.Errorf("failed to get volume: %w", err)
	}

	sz, err := stringToUint64(lbResp.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume size '%s': %w", lbResp.Size, err)
	}

	sectorSz, err := intToUint32Checked(lbResp.SectorSize)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sector size: %w", err)
	}

	return &models.GetVolumeResponse{
		Volume: &models.Volume{
			UUID:               lbResp.Name,
			Size:               sz,
			SectorSize:         sectorSz,
			Acls:               lbResp.ACL.Values,
			IsAvailable:        volumeStateToIsAvail(lbResp.State),
			SourceSnapshotUUID: lbResp.SourceSnapshotName,
		},
	}, nil
}

func (a *ClientAdapter) GetVolumes(_ context.Context, _ *models.GetVolumesRequest,
) (*models.GetVolumesResponse, error) {
	lbResp, err := a.client.GetVolumes() // Note: volume UUID is lightbits volume name
	if err != nil {
		return nil, fmt.Errorf("failed to get volume: %w", err)
	}

	volumes := lo.Map[*Volume, *models.Volume](lbResp, func(v *Volume, _ int) *models.Volume {
		sz, err := stringToUint64(v.Size)
		if err != nil {
			log.Err(err).Msgf("error parsing volume size '%s'", v.Size)

			return nil
		}

		sectorSz, err := intToUint32Checked(v.SectorSize)
		if err != nil {
			log.Err(err).Msgf("error parsing sector size: '%d'", v.SectorSize)

			return nil
		}

		return &models.Volume{
			UUID:               v.Name,
			Size:               sz,
			SectorSize:         sectorSz,
			Acls:               v.ACL.Values,
			IsAvailable:        volumeStateToIsAvail(v.State),
			SourceSnapshotUUID: v.SourceSnapshotName,
		}
	})

	return &models.GetVolumesResponse{
		Volumes: volumes,
	}, nil
}

func (a *ClientAdapter) CreateVolume(_ context.Context, req *models.CreateVolumeRequest,
) (*models.CreateVolumeResponse, error) {
	// Required args: name, acl, repica count, size
	v := &Volume{
		Name: req.UUID,
		ACL: ACL{
			Values: req.Acls,
		},
		ReplicaCount: a.client.replicationFactor,
		Size:         bytesToGiBString(req.Size),
	}

	_, err := a.client.CreateVolume(v)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	return &models.CreateVolumeResponse{
		// Empty; ACK
	}, nil
}

func (a *ClientAdapter) ResizeVolume(_ context.Context, req *models.ResizeVolumeRequest,
) (*models.ResizeVolumeResponse, error) {
	name := req.UUID
	getVolResp, err := a.client.GetVolume(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get volume for resize: %w", err)
	}

	id := getVolResp.UUID // Note: volume UUID is lightbits volume name
	size := bytesToGiBString(req.Size)
	err = a.client.UpdateVolume(id, &UpdateVolumeRequest{Size: size})
	if err != nil {
		return nil, fmt.Errorf("failed to update volume: %w", err)
	}

	return &models.ResizeVolumeResponse{
		// Empty; ACK
	}, nil
}

func (a *ClientAdapter) DeleteVolume(_ context.Context, req *models.DeleteVolumeRequest,
) (*models.DeleteVolumeResponse, error) {
	name := req.UUID // Note: volume UUID is lightbits volume name
	err := a.client.DeleteVolume(name)
	if err != nil {
		return nil, fmt.Errorf("failed to delete volume: %w", err)
	}

	return &models.DeleteVolumeResponse{
		// Empty; ACK
	}, nil
}

func (a *ClientAdapter) AttachVolume(_ context.Context, req *models.AttachVolumeRequest,
) (*models.AttachVolumeResponse, error) {
	name := req.UUID
	getVolResp, err := a.client.GetVolume(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get volume for attachment: %w", err)
	}

	if len(req.Acls) != 1 {
		return nil, errMustHaveOneACL
	}

	addNodes := req.Acls
	removeNodes := []string{}
	acl := constructACLSet(getVolResp.ACL.Values, addNodes, removeNodes)
	err = a.client.UpdateVolume(getVolResp.UUID, &UpdateVolumeRequest{
		ACL: &ACL{
			Values: acl,	
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach volume: %w", err)
	}

	return &models.AttachVolumeResponse{
		// Empty; ACK
	}, nil
}

func (a *ClientAdapter) DetachVolume(_ context.Context, req *models.DetachVolumeRequest,
) (*models.DetachVolumeResponse, error) {
	name := req.UUID
	getVolResp, err := a.client.GetVolume(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get volume for detachment: %w", err)
	}

	if len(req.Acls) != 1 {
		return nil, errMustHaveOneACL
	}

	addNodes := []string{}
	removeNodes := req.Acls
	acl := constructACLSet(getVolResp.ACL.Values, addNodes, removeNodes)
	err = a.client.UpdateVolume(getVolResp.UUID, &UpdateVolumeRequest{
		ACL: &ACL{
			Values: acl,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to detach volume: %w", err)
	}

	return &models.DetachVolumeResponse{
		// Empty; ACK
	}, nil
}

func (a *ClientAdapter) GetSnapshot(_ context.Context, req *models.GetSnapshotRequest,
) (*models.GetSnapshotResponse, error) {
	name := req.UUID
	lbResp, err := a.client.GetSnapshot(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	sz, err := stringToUint64(lbResp.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to parse snapshot size: %w", err)
	}

	sectorSz, err := intToUint32Checked(lbResp.SectorSize)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sector size: %w", err)
	}

	resp := &models.GetSnapshotResponse{
		Snapshot: &models.Snapshot{
			UUID:             lbResp.Name,
			Size:             sz,
			SectorSize:       sectorSz,
			IsAvailable:      snapshotStateToIsAvail(lbResp.State),
			SourceVolumeUUID: lbResp.SourceVolumeName, // TODO - fix
		},
	}

	return resp, nil
}

func (a *ClientAdapter) GetSnapshots(_ context.Context, _ *models.GetSnapshotsRequest,
) (*models.GetSnapshotsResponse, error) {
	lbResp, err := a.client.GetSnapshots()
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots: %w", err)
	}

	snapshots := lo.Map[*Snapshot, *models.Snapshot](lbResp, func(s *Snapshot, _ int) *models.Snapshot {
		sz, err := stringToUint64(s.Size)
		if err != nil {
			log.Err(err).Msgf("error parsing volume size '%s'", s.Size)

			return nil
		}

		sectorSz, err := intToUint32Checked(s.SectorSize)
		if err != nil {
			log.Err(err).Msgf("error parsing sector size: '%d'", s.SectorSize)

			return nil
		}

		return &models.Snapshot{
			UUID:             s.Name,
			Size:             sz,
			SectorSize:       sectorSz,
			IsAvailable:      snapshotStateToIsAvail(s.State),
			SourceVolumeUUID: s.SourceVolumeName, // TODO - fix
		}
	})

	resp := &models.GetSnapshotsResponse{
		Snapshots: snapshots,
	}

	return resp, nil
}

func (a *ClientAdapter) CreateSnapshot(_ context.Context, req *models.CreateSnapshotRequest,
) (*models.CreateSnapshotResponse, error) {
	lbReq := &CreateSnapshotRequest{
		Name:             req.UUID,
		SourceVolumeName: req.SourceVolumeUUID, // TODO - fix
		ProjectName:      a.client.projectName,
	}

	_, err := a.client.CreateSnapshot(lbReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	return &models.CreateSnapshotResponse{
		// Empty; ACK
	}, nil
}

func (a *ClientAdapter) DeleteSnapshot(_ context.Context, req *models.DeleteSnapshotRequest,
) (*models.DeleteSnapshotResponse, error) {
	name := req.UUID
	err := a.client.DeleteSnapshot(name)
	if err != nil {
		return nil, fmt.Errorf("failed to delete snapshot: %w", err)
	}

	return &models.DeleteSnapshotResponse{
		// Empty; ACK
	}, nil
}

func (a *ClientAdapter) CloneVolume(_ context.Context, _ *models.CloneVolumeRequest,
) (*models.CloneVolumeResponse, error) {
	err := a.dmsClient.CloneResource()
	if err != nil {
		return nil, fmt.Errorf("failed to resource: %w", err)
	}

	return &models.CloneVolumeResponse{
		// TODO - populate response fields
	}, nil
}

func (a *ClientAdapter) CloneSnapshot(_ context.Context, _ *models.CloneSnapshotRequest,
) (*models.CloneSnapshotResponse, error) {
	err := a.dmsClient.CloneResource()
	if err != nil {
		return nil, fmt.Errorf("failed to resource: %w", err)
	}

	return &models.CloneSnapshotResponse{
		// TODO - populate response fields
	}, nil
}

func (a *ClientAdapter) GetCloneStatus(_ context.Context, _ *models.GetCloneStatusRequest,
) (*models.GetCloneStatusResponse, error) {
	err := a.dmsClient.GetCloneStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to get clone status: %w", err)
	}

	return &models.GetCloneStatusResponse{
		// TODO - populate response fields
	}, nil
}
