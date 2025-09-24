package models

import timestamppb "google.golang.org/protobuf/types/known/timestamppb"

// --- Begin resources

type Volume struct {
	UUID               string
	Size               uint64
	SectorSize         uint32
	Acls               []string
	IsAvailable        bool
	SourceSnapshotUUID string
}

type Snapshot struct {
	UUID             string
	Size             uint64
	SectorSize       uint32
	IsAvailable      bool
	SourceVolumeUUID string
}

// --- Begin requests and responses

type GetVolumeRequest struct {
	UUID string
}

type GetVolumeResponse struct {
	Volume *Volume
}

type GetVolumesRequest struct {
	// Empty
}

type GetVolumesResponse struct {
	Volumes []*Volume
}

type CreateVolumeRequest struct {
	UUID       string
	Size       uint64
	SectorSize uint32
	Acls       []string
}

type CreateVolumeResponse struct {
	Volume *Volume
}

type ResizeVolumeRequest struct {
	UUID string
	Size uint64
}

type ResizeVolumeResponse struct {
	// Empty; ACK
}

type DeleteVolumeRequest struct {
	UUID string
}

type DeleteVolumeResponse struct {
	// Empty; ACK
}

type AttachVolumeRequest struct {
	UUID string
	Acls []string
}

type AttachVolumeResponse struct {
	// Empty; ACK
}

type DetachVolumeRequest struct {
	UUID string
	Acls []string
}

type DetachVolumeResponse struct {
	// Empty; ACK
}

type GetSnapshotRequest struct {
	UUID string
}

type GetSnapshotResponse struct {
	Snapshot *Snapshot
}

type GetSnapshotsRequest struct {
	// Empty
}

type GetSnapshotsResponse struct {
	Snapshots []*Snapshot
}

type CreateSnapshotRequest struct {
	UUID             string
	Size             uint64
	SectorSize       uint32
	SourceVolumeUUID string
}

type CreateSnapshotResponse struct {
	Snapshot *Snapshot
}

type DeleteSnapshotRequest struct {
	UUID string
}

type DeleteSnapshotResponse struct {
	// Empty; ACK
}

type CloneVolumeRequest struct {
	SrcSnapshotUUID string
	DstVolumeUUID   string
}

type CloneVolumeResponse struct {
	OperationID string
}

type CloneSnapshotRequest struct {
	SrcSnaphotUUID  string
	DstSnapshotUUID string
}

type CloneSnapshotResponse struct {
	OperationID string
}

type GetCloneStatusRequest struct {
	OperationID string
}

type GetCloneStatusResponse struct {
	OperationID string
	CreatedAt   *timestamppb.Timestamp
	EndedAt     *timestamppb.Timestamp
}
