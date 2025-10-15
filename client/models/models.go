package models

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
	UUID   string
	Source CreateVolumeSource
}

type CreateVolumeSource interface {
	isCreateVolumeSource()
}

type NewVolumeSpec struct {
	Size       uint64
	SectorSize uint32
}

// isCreateVolumeSource implements the CreateVolumeSource interface for NewVolumeSpec.
func (NewVolumeSpec) isCreateVolumeSource() {}

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
