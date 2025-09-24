package krusoe

import (
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

const (
	secretAPIKey = "krusoe"
)

var (
	errAuth = errors.New("incorrect api key")

	errResourceNotFound = errors.New("resource not found")
	errCannotResizeDown = errors.New("cannot resize to a smaller size")
	errVolDetached      = errors.New("volume is detached")
	errVolAttached      = errors.New("volume is attached")
	errUnimplemented    = errors.New("unimplemented")
)

type backend struct {
	volumes   map[string]*Volume   // mapping of krusoe volume name to volume
	snapshots map[string]*Snapshot // mapping of krusoe snapshot name to snapshot
}

func newBackend() *backend {
	return &backend{
		volumes:   make(map[string]*Volume),
		snapshots: make(map[string]*Snapshot),
	}
}

func (b *backend) getVolume(apiKey, name string) (*Volume, error) {
	if apiKey != secretAPIKey {
		return nil, errAuth
	}

	v, ok := b.volumes[name]
	if !ok {
		return nil, errResourceNotFound
	}

	return v, nil
}

func (b *backend) getVolumes(apiKey string) ([]*Volume, error) {
	if apiKey != secretAPIKey {
		return nil, errAuth
	}

	return lo.Values(b.volumes), nil
}

func (b *backend) createVolume(apiKey, name string, size, sectorSize uint) (*Volume, error) {
	if apiKey != secretAPIKey {
		return nil, errAuth
	}

	log.Info().Msgf("crusoe backend - create volume: sector size: %d", sectorSize)

	v := &Volume{
		name:       name,
		id:         uuid.NewString(),
		size:       size,
		sectorSize: sectorSize,
		acls:       []string{},
	}

	b.volumes[v.name] = v

	return v, nil
}

func (b *backend) resizeVolume(apiKey, id string, size uint) (*Volume, error) {
	if apiKey != secretAPIKey {
		return nil, errAuth
	}

	v, err := b.getVolume(apiKey, id)
	if err != nil {
		return nil, errResourceNotFound
	}
	if size <= v.size {
		return nil, errCannotResizeDown
	}

	v.size = size

	return v, nil
}

func (b *backend) deleteVolume(apiKey, id string) error {
	if apiKey != secretAPIKey {
		return errAuth
	}

	_, ok := b.volumes[id]
	if !ok {
		return errResourceNotFound
	}

	delete(b.volumes, id)

	return nil
}

func (b *backend) attachVolume(apiKey, id string, acls []string) (*Volume, error) {
	if apiKey != secretAPIKey {
		return nil, errAuth
	}

	v, ok := b.volumes[id]
	if !ok {
		return nil, errResourceNotFound
	}

	if len(v.acls) != 0 {
		return nil, errVolAttached
	}

	v.acls = acls

	return v, nil
}

//nolint:revive,unparam // will need to update when supporting multi-attach, detach acls
func (b *backend) detachVolume(apiKey, id string, acls []string) (*Volume, error) {
	if apiKey != secretAPIKey {
		return nil, errAuth
	}

	v, ok := b.volumes[id]
	if !ok {
		return nil, errResourceNotFound
	}

	if len(v.acls) == 0 {
		return nil, errVolDetached
	}

	v.acls = []string{}

	return v, nil
}

func (b *backend) getSnapshot(apiKey, name string) (*Snapshot, error) {
	if apiKey != secretAPIKey {
		return nil, errAuth
	}

	s, ok := b.snapshots[name]
	if !ok {
		return nil, errResourceNotFound
	}

	return s, nil
}

func (b *backend) getSnapshots(apiKey string) ([]*Snapshot, error) {
	if apiKey != secretAPIKey {
		return nil, errAuth
	}

	return lo.Values(b.snapshots), nil
}

func (b *backend) createSnapshot(apiKey, name, sourceVolumeID string) (*Snapshot, error) {
	if apiKey != secretAPIKey {
		return nil, errAuth
	}

	v, err := b.getVolume(apiKey, sourceVolumeID)
	if err != nil {
		return nil, errResourceNotFound
	}

	s := &Snapshot{
		name:           name,
		id:             uuid.NewString(),
		size:           v.size,
		sectorSize:     v.sectorSize,
		sourceVolumeID: sourceVolumeID,
	}

	b.snapshots[s.name] = s

	return s, nil
}

func (b *backend) deleteSnapshot(apiKey, id string) error {
	if apiKey != secretAPIKey {
		return errAuth
	}

	_, ok := b.snapshots[id]
	if !ok {
		return errResourceNotFound
	}

	delete(b.snapshots, id)

	return nil
}

func (b *backend) cloneVolume(apiKey, srcVolID, dstVolName string) (*Volume, error) {
	if apiKey != secretAPIKey {
		return nil, errAuth
	}

	srcVol, ok := b.volumes[srcVolID]
	if !ok {
		return nil, errResourceNotFound
	}

	dstVol := &Volume{
		name:       dstVolName,
		id:         uuid.NewString(),
		size:       srcVol.size,
		sectorSize: srcVol.sectorSize,
		acls:       srcVol.acls,
	}

	return dstVol, nil
}

func (b *backend) cloneSnapshot(apiKey, srcSnapshotID, dstSnapshotsName string) (*Snapshot, error) {
	if apiKey != secretAPIKey {
		return nil, errAuth
	}

	srcSnapshot, ok := b.snapshots[srcSnapshotID]
	if !ok {
		return nil, errResourceNotFound
	}

	dstSnapshot := &Snapshot{
		name:           dstSnapshotsName,
		id:             uuid.NewString(),
		size:           srcSnapshot.size,
		sectorSize:     srcSnapshot.sectorSize,
		sourceVolumeID: srcSnapshot.sourceVolumeID,
	}

	return dstSnapshot, nil
}

func (b *backend) getCloneStatus(apiKey string) error {
	if apiKey != secretAPIKey {
		return errAuth
	}

	return errUnimplemented
}
