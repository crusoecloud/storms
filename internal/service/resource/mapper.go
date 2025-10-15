package resource

// import (
// 	"errors"
// 	"fmt"
// 	"sync"

// 	"github.com/samber/lo"
// )

// var (
// 	errUnmappedResource = errors.New("unmapped resource")
// )

// // InMemoryMapper provides a thread-safe, in-memory implementation of ResourceMapper.
// type InMemoryMapper struct {
// 	mu                    sync.RWMutex
// 	resourceIDToClusterID map[string]string
// 	clusterIDToResourceID map[string]map[string]struct{}
// }

// // NewInMemoryMapper creates a new InMemoryMapper.
// func NewInMemoryMapper() *InMemoryMapper {
// 	return &InMemoryMapper{
// 		resourceIDToClusterID: make(map[string]string),
// 		clusterIDToResourceID: make(map[string]map[string]struct{}), // Effectively, string:set
// 	}
// }

// func (m *InMemoryMapper) Map(resourceID, clusterID string) error {
// 	m.mu.Lock()
// 	defer m.mu.Unlock()
// 	m.resourceIDToClusterID[resourceID] = clusterID
// 	if _, ok := m.clusterIDToResourceID[clusterID]; !ok {
// 		m.clusterIDToResourceID[clusterID] = make(map[string]struct{})
// 	}
// 	m.clusterIDToResourceID[clusterID][resourceID] = struct{}{} // Create struct to represent "exists".

// 	return nil
// }

// func (m *InMemoryMapper) Unmap(resourceID string) error {
// 	m.mu.Lock()
// 	defer m.mu.Unlock()
// 	clusterID, ok := m.resourceIDToClusterID[resourceID]
// 	if !ok {
// 		return nil
// 	}
// 	delete(m.resourceIDToClusterID, resourceID)
// 	delete(m.clusterIDToResourceID, clusterID)

// 	return nil
// }

// func (m *InMemoryMapper) OwnerCluster(resourceID string) (string, error) {
// 	m.mu.RLock()
// 	defer m.mu.RUnlock()
// 	clusterID, ok := m.resourceIDToClusterID[resourceID]
// 	if !ok {
// 		return "", fmt.Errorf("failed to find resource's cluster: %w", errUnmappedResource)
// 	}

// 	return clusterID, nil
// }

// func (m *InMemoryMapper) ResourceCount() int {
// 	m.mu.RLock()
// 	defer m.mu.RUnlock()

// 	return len(m.resourceIDToClusterID)
// }

// func (m *InMemoryMapper) ResourceCountByCluster(clusterID string) int {
// 	m.mu.RLock()
// 	defer m.mu.RUnlock()
// 	if resources, ok := m.clusterIDToResourceID[clusterID]; ok {
// 		return len(resources)
// 	}

// 	return 0
// }

// func (m *InMemoryMapper) GetAllClusterResources() map[string][]string {
// 	out := make(map[string][]string)
// 	for clusterID, resourceIDSet := range m.clusterIDToResourceID {
// 		out[clusterID] = lo.Keys(resourceIDSet)
// 	}

// 	return out
// }
