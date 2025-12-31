package mocks

type MockAllocator struct {
	MockAllocateCluster func(affinityTags map[string]string) (string, error)
}

func (m *MockAllocator) AllocateCluster(affinityTags map[string]string) (string, error) {
	return m.MockAllocateCluster(affinityTags)
}
