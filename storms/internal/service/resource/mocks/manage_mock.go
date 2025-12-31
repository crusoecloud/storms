package mocks

import "gitlab.com/crusoeenergy/island/storage/storms/storms/internal/service/resource"

type MockResourceManager struct {
	MockMap                       func(r *resource.Resource) error
	MockUnmap                     func(resourceID string) error
	MockGetResourceCluster        func(resourceID string) (string, error)
	MockGetResourceCount          func() int
	MockGetResourcesOfCluster     func(clusterID string) []*resource.Resource
	MockGetResourcesOfAllClusters func() map[string][]*resource.Resource
}

func (m *MockResourceManager) Map(r *resource.Resource) error {
	return m.MockMap(r)
}

func (m *MockResourceManager) Unmap(resourceID string) error {
	return m.MockUnmap(resourceID)
}

func (m *MockResourceManager) GetResourceCluster(resourceID string) (string, error) {
	return m.MockGetResourceCluster(resourceID)
}

func (m *MockResourceManager) GetResourceCount() int {
	return m.MockGetResourceCount()
}

func (m *MockResourceManager) GetResourcesOfCluster(clusterID string) []*resource.Resource {
	return m.MockGetResourcesOfCluster(clusterID)
}

func (m *MockResourceManager) GetResourcesOfAllClusters() map[string][]*resource.Resource {
	return m.MockGetResourcesOfAllClusters()
}
