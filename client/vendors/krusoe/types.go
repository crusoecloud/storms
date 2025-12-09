package krusoe

import (
	"time"
)

type Volume struct {
	name          string
	id            string
	size          uint
	sectorSize    uint
	acl           []string
	srcSnapshotID string
	CreatedAt     time.Time
}

type Snapshot struct {
	name           string
	id             string
	size           uint
	sectorSize     uint
	sourceVolumeID string
	createdAt      time.Time
}
