package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	storms "gitlab.com/crusoeenergy/island/storage/storms/api/gen/go/storms/v1"
	"gitlab.com/crusoeenergy/island/storage/storms/client"
	"gitlab.com/crusoeenergy/island/storage/storms/client/models"
)

var errUnsupportedSectorSize = errors.New("unsupported sector size")

// The ClientTranslator translates federation service requests/responses to and from generic client request/responses.
//
//	Its purpose is to decouple client implementation from StorMS.
//
// The federation service (StorMS) relies on ClientTranslator to be able to communicate with the downstream clients and
// vice versa. This design allows changes to the gRPC service of StorMS to not directly affect the client
// implementations.
type ClientTranslator struct{}

func NewClientTranslator() *ClientTranslator {
	return &ClientTranslator{}
}

func (ct *ClientTranslator) AttachVolume(ctx context.Context, c client.Client, req *storms.AttachVolumeRequest,
) (*storms.AttachVolumeResponse, error) {
	translatedReq := &models.AttachVolumeRequest{
		UUID: req.GetUuid(),
		Acls: req.GetAcls(),
	}

	_, err := c.AttachVolume(ctx, translatedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to attach volume: %w", err)
	}

	return &storms.AttachVolumeResponse{
		// Empty; ACK
	}, nil
}

func (ct *ClientTranslator) CloneSnapshot(_ context.Context, _ client.Client, _ *storms.CloneSnapshotRequest,
) (*storms.CloneSnapshotResponse, error) {
	return nil, fmt.Errorf("not implemented") //nolint:err113 // wip
}

func (ct *ClientTranslator) CloneVolume(_ context.Context, _ client.Client, _ *storms.CloneVolumeRequest,
) (*storms.CloneVolumeResponse, error) {
	return nil, fmt.Errorf("not implemented") //nolint:err113 // wip
}

func (ct *ClientTranslator) CreateSnapshot(ctx context.Context, c client.Client, req *storms.CreateSnapshotRequest,
) (*storms.CreateSnapshotResponse, error) {
	sectorSize, err := int32ToUint32Checked(int32(req.Snapshot.SectorSize))
	if err != nil {
		return nil, fmt.Errorf("failed to convert int32 to uint32: %w", err)
	}
	translatedReq := &models.CreateSnapshotRequest{
		UUID:             req.Snapshot.GetUuid(),
		Size:             req.Snapshot.GetSize(),
		SectorSize:       sectorSize,
		SourceVolumeUUID: req.Snapshot.GetSourceVolumeUuid(),
	}

	_, err = c.CreateSnapshot(ctx, translatedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	return &storms.CreateSnapshotResponse{
		// Empty; ACK
	}, nil
}

func (ct *ClientTranslator) CreateVolume(ctx context.Context, c client.Client, req *storms.CreateVolumeRequest,
) (*storms.CreateVolumeResponse, error) {
	sectorSize, err := int32ToUint32Checked(int32(req.Volume.SectorSize))
	if err != nil {
		return nil, fmt.Errorf("failed to convert int32 to uint32: %w", err)
	}
	translatedReq := &models.CreateVolumeRequest{
		UUID:       req.Volume.GetUuid(),
		Size:       req.Volume.GetSize(),
		SectorSize: sectorSize,
		Acls:       req.Volume.GetAcls(),
	}

	_, err = c.CreateVolume(ctx, translatedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	return &storms.CreateVolumeResponse{
		// Empty; ACK
	}, nil
}

func (ct *ClientTranslator) DeleteSnapshot(ctx context.Context, c client.Client, req *storms.DeleteSnapshotRequest,
) (*storms.DeleteSnapshotResponse, error) {
	translatedReq := &models.DeleteSnapshotRequest{
		UUID: req.GetUuid(),
	}

	_, err := c.DeleteSnapshot(ctx, translatedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to delete snapshot: %w", err)
	}

	return &storms.DeleteSnapshotResponse{
		// Empty; ACK
	}, nil
}

func (ct *ClientTranslator) DeleteVolume(ctx context.Context, c client.Client, req *storms.DeleteVolumeRequest,
) (*storms.DeleteVolumeResponse, error) {
	translatedReq := &models.DeleteVolumeRequest{
		UUID: req.GetUuid(),
	}

	_, err := c.DeleteVolume(ctx, translatedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to delete volume: %w", err)
	}

	return &storms.DeleteVolumeResponse{
		// Empty; ACK
	}, nil
}

func (ct *ClientTranslator) DetachVolume(ctx context.Context, c client.Client, req *storms.DetachVolumeRequest,
) (*storms.DetachVolumeResponse, error) {
	translatedReq := &models.DetachVolumeRequest{
		UUID: req.GetUuid(),
	}

	_, err := c.DetachVolume(ctx, translatedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to detach volume: %w", err)
	}

	return &storms.DetachVolumeResponse{
		// Empty; ACK
	}, nil
}

func (ct *ClientTranslator) GetCloneStatus(_ context.Context, _ client.Client, _ *storms.GetCloneStatusRequest,
) (*storms.GetCloneStatusResponse, error) {
	return nil, fmt.Errorf("not implemented") //nolint:err113 // wip
}

func (ct *ClientTranslator) GetSnapshot(ctx context.Context, c client.Client, req *storms.GetSnapshotRequest,
) (*storms.GetSnapshotResponse, error) {
	translatedReq := &models.GetSnapshotRequest{
		UUID: req.GetUuid(),
	}

	s, err := c.GetSnapshot(ctx, translatedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	sectorSizeEnum, err := uint32ToSectorSizeEnumChecked(s.Snapshot.SectorSize)
	if err != nil {
		return nil, fmt.Errorf("failed to translate sector size: %w", err)
	}

	return &storms.GetSnapshotResponse{
		Snapshot: &storms.Snapshot{
			Uuid:             s.Snapshot.UUID,
			Size:             s.Snapshot.Size,
			SectorSize:       sectorSizeEnum,
			IsAvailable:      s.Snapshot.IsAvailable,
			SourceVolumeUuid: s.Snapshot.SourceVolumeUUID,
		},
	}, nil
}

func (ct *ClientTranslator) GetSnapshots(ctx context.Context, c client.Client, _ *storms.GetSnapshotsRequest,
) (*storms.GetSnapshotsResponse, error) {
	translatedReq := &models.GetSnapshotsRequest{}
	snapshots, err := c.GetSnapshots(ctx, translatedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots: %w", err)
	}

	ss := lo.Map[*models.Snapshot, *storms.Snapshot](
		snapshots.Snapshots,
		func(s *models.Snapshot, _ int) *storms.Snapshot {
			sectorSizeEnum, err := uint32ToSectorSizeEnumChecked(s.SectorSize)
			if err != nil {
				log.Err(err)

				return nil
			}

			return &storms.Snapshot{
				Uuid:             s.UUID,
				Size:             s.Size,
				SectorSize:       sectorSizeEnum,
				IsAvailable:      s.IsAvailable,
				SourceVolumeUuid: s.SourceVolumeUUID,
			}
		})

	return &storms.GetSnapshotsResponse{
		Snapshots: ss,
	}, nil
}

func (ct *ClientTranslator) GetVolume(ctx context.Context, c client.Client, req *storms.GetVolumeRequest,
) (*storms.GetVolumeResponse, error) {
	translatedRequest := &models.GetVolumeRequest{
		UUID: req.GetUuid(),
	}

	clientResp, err := c.GetVolume(ctx, translatedRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to get volume: %w", err)
	}
	vol := clientResp.Volume

	sectorSizeEnum, err := uint32ToSectorSizeEnumChecked(vol.SectorSize)
	if err != nil {
		return nil, fmt.Errorf("failed translate sector size: %w", err)
	}

	resp := &storms.GetVolumeResponse{
		Volume: &storms.Volume{
			Uuid:               vol.UUID,
			Size:               vol.Size,
			SectorSize:         sectorSizeEnum,
			Acls:               vol.Acls,
			IsAvailable:        vol.IsAvailable,
			SourceSnapshotUuid: vol.SourceSnapshotUUID,
		},
	}

	return resp, nil
}

func (ct *ClientTranslator) GetVolumes(ctx context.Context, c client.Client, _ *storms.GetVolumesRequest,
) (*storms.GetVolumesResponse, error) {
	translatedReq := &models.GetVolumesRequest{}
	volumes, err := c.GetVolumes(ctx, translatedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots: %w", err)
	}

	vs := lo.Map[*models.Volume, *storms.Volume](volumes.Volumes, func(v *models.Volume, _ int) *storms.Volume {
		sectorSizeEnum, err := uint32ToSectorSizeEnumChecked(v.SectorSize)
		if err != nil {
			log.Err(err)

			return nil
		}

		return &storms.Volume{
			Uuid:               v.UUID,
			Size:               v.Size,
			SectorSize:         sectorSizeEnum,
			IsAvailable:        v.IsAvailable,
			SourceSnapshotUuid: v.SourceSnapshotUUID, // TODO - fix
		}
	})

	return &storms.GetVolumesResponse{
		Volumes: vs,
	}, nil
}

func (ct *ClientTranslator) ResizeVolume(ctx context.Context, c client.Client, req *storms.ResizeVolumeRequest,
) (*storms.ResizeVolumeResponse, error) {
	translatedReq := &models.ResizeVolumeRequest{
		UUID: req.GetUuid(),
		Size: req.GetSize(),
	}

	_, err := c.ResizeVolume(ctx, translatedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to resize volume: %w", err)
	}

	return &storms.ResizeVolumeResponse{
		// Empty; ACK
	}, nil
}

// Begin -- Helper function(s)

func uint32ToSectorSizeEnumChecked(u uint32) (storms.SectorSizeEnum, error) {
	sectorSize, err := uint32ToInt32Checked(u)
	if err != nil {
		return storms.SectorSizeEnum_SECTOR_SIZE_ENUM_UNSPECIFIED,
			fmt.Errorf("failed to convert uint32 to int32: %w", err)
	}
	_, ok := storms.SectorSizeEnum_name[sectorSize]
	if !ok {
		return storms.SectorSizeEnum_SECTOR_SIZE_ENUM_UNSPECIFIED, errUnsupportedSectorSize
	}

	return storms.SectorSizeEnum(sectorSize), nil
}

var (
	errInt32OutOfRange  = errors.New("int32 out of range")
	errUint32OutOfRange = errors.New("uint32 out of range")
)

func int32ToUint32Checked(i int32) (uint32, error) {
	if i < 0 {
		return 0, errInt32OutOfRange
	}

	return uint32(i), nil
}

func uint32ToInt32Checked(u uint32) (int32, error) {
	if u > math.MaxInt32 {
		return 0, errUint32OutOfRange
	}

	return int32(u), nil
}
