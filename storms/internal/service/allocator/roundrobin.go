package allocator

import (
	"fmt"
	"sync"
)

type clusterManager interface {
	AllIDs() []string
}

// RoundRobinAllocator provides a simple round-robin implementation of Allocator.
type RoundRobinAllocator struct {
	mu             sync.Mutex
	clusterManager clusterManager
	nextIndex      int
}

// NewRoundRobinAllocator creates a new RoundRobinAllocator.
func NewRoundRobinAllocator(cm clusterManager) *RoundRobinAllocator {
	return &RoundRobinAllocator{
		clusterManager: cm,
	}
}

func (a *RoundRobinAllocator) SelectClusterForNewResource() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	clusterIDs := a.clusterManager.AllIDs()
	if len(clusterIDs) == 0 {
		return "", fmt.Errorf("no clusters available for allocation")
	}
	// Simple round-robin logic
	if a.nextIndex >= len(clusterIDs) {
		a.nextIndex = 0
	}
	selectedClusterID := clusterIDs[a.nextIndex]
	a.nextIndex++

	return selectedClusterID, nil
}
