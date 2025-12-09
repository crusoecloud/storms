package krusoe

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
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

func (b *backend) createNewVolume(apiKey, name string, size, sectorSize uint) (*Volume, error) {
	if apiKey != secretAPIKey {
		return nil, errAuth
	}

	v := &Volume{
		name:       name,
		id:         uuid.NewString(),
		size:       size,
		sectorSize: sectorSize,
		acl:        []string{},
		CreatedAt:  time.Now(),
	}

	b.volumes[v.name] = v

	return v, nil
}

func (b *backend) createVolumeFromSnapshot(apiKey, name, srcSnapshotName string) (*Volume, error) {
	if apiKey != secretAPIKey {
		return nil, errAuth
	}

	s, err := b.getSnapshot(apiKey, srcSnapshotName)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	v := &Volume{
		name:       name,
		id:         uuid.NewString(),
		size:       s.size,
		sectorSize: s.sectorSize,
		acl:        []string{},
		CreatedAt:  time.Now(),
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

func (b *backend) attachVolume(apiKey, id string, acl []string) (*Volume, error) {
	if apiKey != secretAPIKey {
		return nil, errAuth
	}

	v, ok := b.volumes[id]
	if !ok {
		return nil, errResourceNotFound
	}

	if len(v.acl) != 0 {
		return nil, errVolAttached
	}

	v.acl = acl

	return v, nil
}

func (b *backend) detachVolume(apiKey, id string) (*Volume, error) {
	if apiKey != secretAPIKey {
		return nil, errAuth
	}

	v, ok := b.volumes[id]
	if !ok {
		return nil, errResourceNotFound
	}

	if len(v.acl) == 0 {
		return nil, errVolDetached
	}

	v.acl = []string{}

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
		createdAt:      time.Now(),
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
