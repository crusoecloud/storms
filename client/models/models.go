package models

// --- Begin resources

type Volume struct {
	UUID               string
	VendorVolumeID     string
	Size               uint64
	SectorSize         uint32
	ACL                []string
	IsAvailable        bool
	SourceSnapshotUUID string
}

type Snapshot struct {
	UUID             string
	VendorSnapshotID string
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
	UUID   string
	Source CreateVolumeSource
}

type CreateVolumeSource interface {
	isCreateVolumeSource()
}

// For creating a volume from scratch.
type NewVolumeSpec struct {
	Size       uint64 // Size of volume (unit: bytes)
	SectorSize uint32 // Size of sector (unit: bytes)
}

// isCreateVolumeSource implements the CreateVolumeSource interface for NewVolumeSpec.
func (NewVolumeSpec) isCreateVolumeSource() {}

// For creating a volume from snapshot. The volume should inherit relevant properties of the snapshot such as size.
type SnapshotSource struct {
	SnapshotUUID string
}

// isCreateVolumeSource implements the CreateVolumeSource interface for SnapshotSource.
func (SnapshotSource) isCreateVolumeSource() {}

type CreateVolumeResponse struct {
	Volume *Volume
}

type ResizeVolumeRequest struct {
	UUID string
	Size uint64 // Size of volume (unit: bytes)
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
	ACL  []string // Unique NQN of host to attach the volume to
}

type AttachVolumeResponse struct {
	// Empty; ACK
}

type DetachVolumeRequest struct {
	UUID string
	ACL  []string // Unique NQN of host to deatch the volume from
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
