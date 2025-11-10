package krusoe

type Volume struct {
	name          string
	id            string
	size          uint
	sectorSize    uint
	acl           []string
	srcSnapshotID string
}

type Snapshot struct {
	name           string
	id             string
	size           uint
	sectorSize     uint
	sourceVolumeID string
}
