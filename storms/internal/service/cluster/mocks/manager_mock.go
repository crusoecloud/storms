package mocks

import "gitlab.com/crusoeenergy/island/storage/storms/storms/internal/service/cluster"

type MockClusterManager struct {
	MockSet    func(clusterID string, cluster *cluster.Cluster) error
	MockRemove func(clusterID string) error
	MockGet    func(clusterID string) (*cluster.Cluster, error)
	MockAllIDs func() []string
	MockCount  func() int
}

func (m *MockClusterManager) Set(clusterID string, c *cluster.Cluster) error {
	return m.MockSet(clusterID, c)
}

func (m *MockClusterManager) Remove(clusterID string) error {
	return m.MockRemove(clusterID)
}

func (m *MockClusterManager) Get(clusterID string) (*cluster.Cluster, error) {
	return m.MockGet(clusterID)
}

func (m *MockClusterManager) AllIDs() []string {
	return m.MockAllIDs()
}

func (m *MockClusterManager) Count() int {
	return m.MockCount()
}
