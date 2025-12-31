package cluster

import (
	"fmt"
	"sync"

	"github.com/samber/lo"
)

// InMemoryManager provides a thread-safe, in-memory implementation of ClusterManager.
type InMemoryManager struct {
	mu       sync.RWMutex
	clusters map[string]*Cluster
}

// NewInMemoryManager creates a new InMemoryManager.
func NewInMemoryManager() *InMemoryManager {
	return &InMemoryManager{
		clusters: make(map[string]*Cluster),
	}
}

func (m *InMemoryManager) Set(clusterID string, c *Cluster) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clusters[clusterID] = c

	return nil
}

func (m *InMemoryManager) Remove(clusterID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.clusters[clusterID]; !exists {
		return fmt.Errorf("cluster with ID '%s' not found", clusterID)
	}
	delete(m.clusters, clusterID)

	return nil
}

func (m *InMemoryManager) Get(clusterID string) (*Cluster, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, exists := m.clusters[clusterID]
	if !exists {
		return nil, fmt.Errorf("cluster ID '%s' not found", clusterID)
	}

	return c, nil
}

func (m *InMemoryManager) AllIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return lo.Keys(m.clusters)
}

func (m *InMemoryManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.clusters)
}
