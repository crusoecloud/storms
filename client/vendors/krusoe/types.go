package krusoe

type Volume struct {
	name       string
	id         string
	size       uint
	sectorSize uint
	acls       []string
	srcSnapshotID string
}

type Snapshot struct {
	name           string
	id             string
	size           uint
	sectorSize     uint
	sourceVolumeID string
}
