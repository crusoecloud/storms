//nolint:tagliatelle // tag names are defined by lightbits api
package lightbits

import (
	"time"

	"github.com/google/uuid"
)

type VolumeState string

const (
	VolumeStateUnknown   VolumeState = "Unknown"
	VolumeStateCreating  VolumeState = "Creating"
	VolumeStateAvailable VolumeState = "Available"
	VolumeStateDeleting  VolumeState = "Deleting"
	VolumeStateDeleted   VolumeState = "Deleted"
	VolumeStateFailed    VolumeState = "Failed"
	VolumeStateUpdating  VolumeState = "Updating"
	VolumeStateRollback  VolumeState = "Rollback"
	VolumeStateMigrating VolumeState = "Migrating"
)

// These are constants expected by lightbits which mean the access control list either accepts everything or nothing.
const (
	ACLNone = "ALLOW_NONE"
	ACLAll  = "ALLOW_ALL"
)

// ProtectionState represents the difference between the desired and current state of
// replication of a volume in a Lightbits Cluster.
// From the documentation:
// "If a node fails, volumes that have data stored on that node can be affected.
// For a volume with a replication factor of 3, a single node failure may cause the volume
// protection state to become Degraded. If another node fails, the volume’s state may become
// ReadOnly.".
// See README.md for a link to documentation.
type ProtectionState string

const (
	ProtectionStateUnknown        ProtectionState = "Unknown"
	ProtectionStateFullyProtected ProtectionState = "FullyProtected"
	ProtectionStateDegraded       ProtectionState = "Degraded"
	ProtectionStateReadOnly       ProtectionState = "ReadOnly"
	ProtectionStateNotAvailable   ProtectionState = "NotAvailable"
)

type ACL struct {
	Values []string `json:"values"`
}

type Volume struct {
	State           VolumeState     `json:"state"`
	ProtectionState ProtectionState `json:"protectionState"`
	ReplicaCount    int             `json:"replicaCount"`
	NodeList        []uuid.UUID     `json:"nodeList"`
	UUID            uuid.UUID       `json:"uuid"`
	NamespaceID     int             `json:"nsid"`
	ACL             ACL             `json:"acl"`
	// Valid values are true/enable/enabled or false/disable/disabled
	Compression        string `json:"compression"`
	Size               string `json:"size"`
	Name               string `json:"name"`
	RebuildProgress    string `json:"rebuildProgress"`
	SectorSize         int    `json:"sectorSize"`
	ProjectName        string `json:"projectName"`
	SourceSnapshotUUID string `json:"sourceSnapshotUUID"`
	SourceSnapshotName string `json:"sourceSnapshotName"`
	CreationTime       time.Time `json:"creationTime"`
}

type SnapshotState string

const (
	SnapshotStateUnknown   SnapshotState = "Unknown"
	SnapshotStateCreating  SnapshotState = "Creating"
	SnapshotStateAvailable SnapshotState = "Available"
	SnapshotStateDeleting  SnapshotState = "Deleting"
	SnapshotStateDeleted   SnapshotState = "Deleted"
	SnapshotStateFailed    SnapshotState = "Failed"
)

type SnapshotStatistics struct {
	PhysicalCapacity      string `json:"physicalCapacity"`
	PhysicalOwnedCapacity string `json:"physicalOwnedCapacity"`
	PhysicalOwnedMemory   string `json:"physicalOwnedMemory"`
	PhysicalMemory        string `json:"physicalMemory"`
	UserWritten           string `json:"userWritten"`
}

type Snapshot struct {
	State            SnapshotState      `json:"state"`
	UUID             uuid.UUID          `json:"uuid"`
	Name             string             `json:"name"`
	Description      string             `json:"description"`
	CreationTime     time.Time          `json:"creationTime"`
	SourceVolumeUUID uuid.UUID          `json:"sourceVolumeUUID"`
	SourceVolumeName string             `json:"sourceVolumeName"`
	ReplicaCount     int                `json:"replicaCount"`
	NodeList         []string           `json:"nodeList"`
	Size             string             `json:"size"`
	SectorSize       int                `json:"sectorSize"`
	ProjectName      string             `json:"projectName"`
	Statistics       SnapshotStatistics `json:"statistics"`
}

type CreateSnapshotRequest struct {
	Name             string `json:"name"`
	SourceVolumeUUID string `json:"sourceVolumeUUID,omitempty"`
	SourceVolumeName string `json:"sourceVolumeName,omitempty"`
	RetentionTime    string `json:"retentionTime,omitempty"`
	Description      string `json:"description,omitempty"`
	ProjectName      string `json:"projectName"`
}

type GetVolumeResponse struct {
	Volumes []*Volume `json:"volumes"`
}

type GetSnapshotResponse struct {
	Snapshots []*Snapshot `json:"snapshots"`
}

type UpdateVolumeRequest struct {
	Size string `json:"size,omitempty"`
	ACL  *ACL   `json:"acl,omitempty"`
}
