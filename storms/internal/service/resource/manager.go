package resource

import (
	"errors"
	"fmt"
	"sync"

	"github.com/samber/lo"
)

var (
	errUnmappedResource = errors.New("unmapped resource")
)

type InMemoryManager struct {
	mu sync.RWMutex

	resourceIDToResourceMetadata map[string]*Resource
}

func NewInMemoryManager() *InMemoryManager {
	return &InMemoryManager{
		resourceIDToResourceMetadata: make(map[string]*Resource),
	}
}

func (m *InMemoryManager) Map(r *Resource) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.resourceIDToResourceMetadata[r.ID] = r

	return nil
}

func (m *InMemoryManager) Unmap(resourceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.resourceIDToResourceMetadata, resourceID)

	return nil
}

func (m *InMemoryManager) GetResourceCluster(resourceID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	r, ok := m.resourceIDToResourceMetadata[resourceID]
	if !ok {
		return "", fmt.Errorf("couldn't find cluster for resource %s: %w", resourceID, errUnmappedResource)
	}

	return r.ClusterID, nil
}

func (m *InMemoryManager) GetResourceCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.resourceIDToResourceMetadata)
}

func (m *InMemoryManager) GetResourcesOfCluster(clusterID string) []*Resource {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := lo.FilterMap(lo.Keys(m.resourceIDToResourceMetadata), func(id string, _ int) (*Resource, bool) {
		r := m.resourceIDToResourceMetadata[id]

		return r, r.ClusterID == clusterID
	})

	return out
}

func (m *InMemoryManager) GetResourcesOfAllClusters() map[string][]*Resource {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := lo.GroupByMap(lo.Keys(m.resourceIDToResourceMetadata), func(id string) (string, *Resource) {
		return m.resourceIDToResourceMetadata[id].ClusterID, m.resourceIDToResourceMetadata[id]
	})

	return out
}
