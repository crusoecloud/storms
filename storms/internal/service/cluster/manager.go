package cluster

import (
	"fmt"
	"sync"

	"github.com/samber/lo"
	"gitlab.com/crusoeenergy/island/storage/storms/client"
)

// InMemoryManager provides a thread-safe, in-memory implementation of ClusterManager.
type InMemoryManager struct {
	mu      sync.RWMutex
	clients map[string]client.Client
}

// NewInMemoryManager creates a new InMemoryManager.
func NewInMemoryManager() *InMemoryManager {
	return &InMemoryManager{
		clients: make(map[string]client.Client),
	}
}

func (m *InMemoryManager) Set(clusterID string, c client.Client) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.clients[clusterID]; exists {
		// This can be idempotent by just returning nil if the client is the same,
		// or replacing it. For now, we'll error.
		return fmt.Errorf("cluster with ID '%s' already exists", clusterID)
	}
	m.clients[clusterID] = c

	return nil
}

func (m *InMemoryManager) Remove(clusterID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.clients[clusterID]; !exists {
		return fmt.Errorf("cluster with ID '%s' not found", clusterID)
	}
	delete(m.clients, clusterID)

	return nil
}

func (m *InMemoryManager) Get(clusterID string) (client.Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, exists := m.clients[clusterID]
	if !exists {
		return nil, fmt.Errorf("client for cluster ID '%s' not found", clusterID)
	}

	return c, nil
}

func (m *InMemoryManager) AllIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return lo.Keys(m.clients)
}

func (m *InMemoryManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.clients)
}
