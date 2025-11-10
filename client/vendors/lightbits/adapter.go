package lightbits

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"gitlab.com/crusoeenergy/island/storage/storms/client/models"
)

var (
	errMustHaveOneACL        = errors.New("must have exactly 1 ACL")
	errUnsupportVolumeSource = errors.New("unsupport volume source")
)

// The Lightbits adapter is a wrapper around the Lightbits client that translates generic federation-level
// requests to Lightbits-specific API requests.
type ClientAdapter struct {
	client *Client
}

func NewClientAdapter(cfg *ClientConfig) (*ClientAdapter, error) {
	c, err := NewClientV2(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create new lightbits client: %w", err)
	}

	return &ClientAdapter{
		client: c,
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
		return nil, fmt.Errorf("failed to cast sector size: %w", err)
	}

	return &models.GetVolumeResponse{
		Volume: &models.Volume{
			UUID:               lbResp.Name,
			VendorVolumeID:     lbResp.UUID.String(),
			Size:               sz,
			SectorSize:         sectorSz,
			ACL:                lbResp.ACL.Values,
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
			VendorVolumeID:     v.UUID.String(),
			Size:               sz,
			SectorSize:         sectorSz,
			ACL:                v.ACL.Values,
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
	var v *Volume
	var err error

	switch source := req.Source.(type) {
	case *models.NewVolumeSpec:
		v, err = createEmptyVolHelper(req.UUID, a.client.replicationFactor, source.Size, source.SectorSize)
		if err != nil {
			return nil, fmt.Errorf("failed to create request to create new empty volume: %w", err)
		}

	case *models.SnapshotSource:
		lbSnapshot, err1 := a.client.GetSnapshot(source.SnapshotUUID)
		if err1 != nil {
			return nil, fmt.Errorf("failed to get snapshot: %w", err1)
		}
		v = createVolFromSnapshotHelper(req.UUID, lbSnapshot)

	default:

		return nil, errUnsupportVolumeSource
	}

	lbVol, err := a.client.CreateVolume(v)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	genericVol, err := translateLBVolToGenericVolHelper(lbVol)
	if err != nil {
		return nil, fmt.Errorf("failed to translate lightbits volume to generic volume: %w", err)
	}

	return &models.CreateVolumeResponse{
		Volume: genericVol,
	}, nil
}

func createEmptyVolHelper(id string, rf int, size uint64, sectorSize uint32) (*Volume, error) {
	sectorSz, err := uint32ToIntChecked(sectorSize)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sector size: %w", err)
	}
	v := &Volume{
		Name:         id,
		ReplicaCount: rf,
		Size:         bytesToGiBString(size),
		SectorSize:   sectorSz,
		ACL: ACL{
			Values: []string{ACLNone},
		},
	}

	return v, nil
}

func createVolFromSnapshotHelper(id string, snapshot *Snapshot) *Volume {
	v := &Volume{
		Name:         id,
		ReplicaCount: snapshot.ReplicaCount,
		Size:         snapshot.Size,
		SectorSize:   snapshot.SectorSize,
		ACL: ACL{
			Values: []string{ACLNone},
		},
		SourceSnapshotUUID: snapshot.UUID.String(),
	}

	return v
}

func translateLBVolToGenericVolHelper(vol *Volume) (*models.Volume, error) {
	sizeUint64, err := strconv.Atoi(vol.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to convert size string to bytes: %w", err)
	}

	sectorSize, err := intToUint32Checked(vol.SectorSize)
	if err != nil {
		return nil, fmt.Errorf("failed to cast sector size: %w", err)
	}

	return &models.Volume{
		UUID:               vol.Name,
		VendorVolumeID:     vol.UUID.String(),
		Size:               uint64(sizeUint64),
		SectorSize:         sectorSize,
		ACL:                vol.ACL.Values,
		IsAvailable:        volumeStateToIsAvail(vol.State),
		SourceSnapshotUUID: vol.SourceSnapshotName,
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

	if len(req.ACL) != 1 {
		return nil, errMustHaveOneACL
	}

	addNodes := req.ACL
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

	if len(req.ACL) != 1 {
		return nil, errMustHaveOneACL
	}

	addNodes := []string{}
	removeNodes := req.ACL
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
			VendorSnapshotID: lbResp.UUID.String(),
			Size:             sz,
			SectorSize:       sectorSz,
			IsAvailable:      snapshotStateToIsAvail(lbResp.State),
			SourceVolumeUUID: lbResp.SourceVolumeName,
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
			VendorSnapshotID: s.UUID.String(),
			Size:             sz,
			SectorSize:       sectorSz,
			IsAvailable:      snapshotStateToIsAvail(s.State),
			SourceVolumeUUID: s.SourceVolumeName,
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
		SourceVolumeName: req.SourceVolumeUUID,
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
