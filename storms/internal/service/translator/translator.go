package translator

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"

	"gitlab.com/crusoeenergy/island/storage/storms/client"
	"gitlab.com/crusoeenergy/island/storage/storms/client/models"
	storms "gitlab.com/crusoeenergy/island/storage/storms/pkg/api/gen/go/storms/v1"
)

var (
	errUnsupportedSectorSize        = errors.New("unsupported sector size")
	errNilNewVolumeSpecs            = errors.New("nil new volume specs")
	errNilSnapshotSourceVolumeSpecs = errors.New("nil snapshot-source volume specs")
	errUnsupportVolumeSource        = errors.New("unsupported or missing volume source in request")
)

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
		ACL:  req.GetAcl(),
	}

	_, err := c.AttachVolume(ctx, translatedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to attach volume: %w", err)
	}

	return &storms.AttachVolumeResponse{
		// Empty; ACK
	}, nil
}

func (ct *ClientTranslator) CreateSnapshot(ctx context.Context, c client.Client, req *storms.CreateSnapshotRequest,
) (*storms.CreateSnapshotResponse, error) {
	translatedReq := &models.CreateSnapshotRequest{
		UUID:             req.GetUuid(),
		SourceVolumeUUID: req.SrcVolumeUuid,
	}

	_, err := c.CreateSnapshot(ctx, translatedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	return &storms.CreateSnapshotResponse{
		// Empty; ACK
	}, nil
}

func (ct *ClientTranslator) CreateVolume(ctx context.Context, c client.Client, req *storms.CreateVolumeRequest,
) (*storms.CreateVolumeResponse, error) {
	translatedReq := &models.CreateVolumeRequest{
		UUID: req.GetUuid(),
	}

	switch source := req.GetSource().(type) {
	case *storms.CreateVolumeRequest_FromNew:
		spec := source.FromNew
		if spec == nil {
			return nil, errNilNewVolumeSpecs
		}

		sectorSize, err := translateSectorSizeEnumToUint32(spec.GetSectorSize())
		if err != nil {
			return nil, fmt.Errorf("failed to translate sector size: %w", err)
		}

		translatedReq.Source = &models.NewVolumeSpec{
			Size:       spec.GetSize(),
			SectorSize: sectorSize,
		}

	case *storms.CreateVolumeRequest_FromSnapshot:
		spec := source.FromSnapshot
		if spec == nil {
			return nil, errNilSnapshotSourceVolumeSpecs
		}

		translatedReq.Source = &models.SnapshotSource{
			SnapshotUUID: spec.GetSnapshotUuid(),
		}

	default:
		// This handles the case where the 'source' oneof is not set.
		return nil, errUnsupportVolumeSource
	}

	_, err := c.CreateVolume(ctx, translatedReq)
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
		ACL:  req.GetAcl(),
	}

	_, err := c.DetachVolume(ctx, translatedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to detach volume: %w", err)
	}

	return &storms.DetachVolumeResponse{
		// Empty; ACK
	}, nil
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
	snapshot := s.Snapshot
	if snapshot == nil {
		return &storms.GetSnapshotResponse{Snapshot: nil}, nil
	}

	sectorSizeEnum := translateUint32ToSectorSizeEnum(snapshot.SectorSize)

	return &storms.GetSnapshotResponse{
		Snapshot: &storms.Snapshot{
			Uuid:             snapshot.UUID,
			VendorSnapshotId: snapshot.VendorSnapshotID,
			Size:             snapshot.Size,
			SectorSize:       sectorSizeEnum,
			IsAvailable:      snapshot.IsAvailable,
			SourceVolumeUuid: snapshot.SourceVolumeUUID,
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
			sectorSizeEnum := translateUint32ToSectorSizeEnum(s.SectorSize)

			return &storms.Snapshot{
				Uuid:             s.UUID,
				VendorSnapshotId: s.VendorSnapshotID,
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
	if vol == nil {
		return &storms.GetVolumeResponse{Volume: nil}, nil
	}

	sectorSizeEnum := translateUint32ToSectorSizeEnum(vol.SectorSize)

	resp := &storms.GetVolumeResponse{
		Volume: &storms.Volume{
			Uuid:               vol.UUID,
			VendorVolumeId:     vol.VendorVolumeID,
			Size:               vol.Size,
			SectorSize:         sectorSizeEnum,
			Acl:                vol.ACL,
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
		sectorSizeEnum := translateUint32ToSectorSizeEnum(v.SectorSize)

		return &storms.Volume{
			Uuid:               v.UUID,
			VendorVolumeId:     v.VendorVolumeID,
			Size:               v.Size,
			SectorSize:         sectorSizeEnum,
			IsAvailable:        v.IsAvailable,
			SourceSnapshotUuid: v.SourceSnapshotUUID,
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

// Begin -- Helper

func translateUint32ToSectorSizeEnum(u uint32) storms.SectorSizeEnum {
	switch u {
	case 4096:
		return storms.SectorSizeEnum_SECTOR_SIZE_ENUM_4096
	case 512:
		return storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512

	default:
		return storms.SectorSizeEnum_SECTOR_SIZE_ENUM_UNSPECIFIED
	}
}

func translateSectorSizeEnumToUint32(e storms.SectorSizeEnum) (uint32, error) {
	switch e {
	case storms.SectorSizeEnum_SECTOR_SIZE_ENUM_512:
		return 512, nil
	case storms.SectorSizeEnum_SECTOR_SIZE_ENUM_4096:
		return 4096, nil
	case storms.SectorSizeEnum_SECTOR_SIZE_ENUM_UNSPECIFIED:
		return 0, nil
	default:
		return 0, errUnsupportedSectorSize
	}
}
