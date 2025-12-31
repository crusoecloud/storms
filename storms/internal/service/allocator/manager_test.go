package allocator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	clientmocks "gitlab.com/crusoeenergy/island/storage/storms/client/mocks"
	clustermocks "gitlab.com/crusoeenergy/island/storage/storms/storms/internal/service/cluster/mocks"

	"gitlab.com/crusoeenergy/island/storage/storms/storms/internal/service/cluster"
)

var (
	vendor1       = "vendor-a"
	clusterID1    = "b22de56f-600f-4fa1-99ba-8f045e905f8e"
	client1       = &clientmocks.MockClient{}
	affinityTags1 = map[string]string{
		"region": "us-east-1",
		"tier":   "tier-1",
		"type":   "type-a",
	}
	cluster1 = &cluster.Cluster{
		Config: &cluster.Config{
			Vendor:       vendor1,
			ClusterID:    clusterID1,
			AffinityTags: affinityTags1,
			VendorConfig: nil,
		},
		Client: client1,
	}

	vendor2       = "vendor-b"
	clusterID2    = "2ecbb15a-3a55-4085-856a-296565f34f40"
	client2       = &clientmocks.MockClient{}
	affinityTags2 = map[string]string{
		"region": "us-south-1",
		"tier":   "tier-2",
		"type":   "type-a",
	}
	cluster2 = &cluster.Cluster{
		Config: &cluster.Config{
			Vendor:       vendor2,
			ClusterID:    clusterID2,
			AffinityTags: affinityTags2,
			VendorConfig: nil,
		},
		Client: client2,
	}

	vendor3       = "vendor-c"
	clusterID3    = "33d46525-72d9-4427-b5ac-6254fe081da3"
	client3       = &clientmocks.MockClient{}
	affinityTags3 = map[string]string{
		"region": "us-east-1",
		"tier":   "tier-1",
		"type":   "type-a",
	}
	cluster3 = &cluster.Cluster{
		Config: &cluster.Config{
			Vendor:       vendor3,
			ClusterID:    clusterID3,
			AffinityTags: affinityTags3,
			VendorConfig: nil,
		},
		Client: client3,
	}

	vendor4       = "vendor-d"
	clusterID4    = "c7228eca-81a1-4d62-93b6-165162094ec3"
	client4       = &clientmocks.MockClient{}
	affinityTags4 = map[string]string{
		"region": "us-west-1",
		"tier":   "tier-2",
		"type":   "type-b",
	}
	cluster4 = &cluster.Cluster{
		Config: &cluster.Config{
			Vendor:       vendor4,
			ClusterID:    clusterID4,
			AffinityTags: affinityTags4,
			VendorConfig: nil,
		},
		Client: client4,
	}

	mockClusterManager = &clustermocks.MockClusterManager{
		MockGet: func(clusterID string) (*cluster.Cluster, error) {
			switch clusterID {
			case clusterID1:
				return cluster1, nil
			case clusterID2:
				return cluster2, nil
			case clusterID3:
				return cluster3, nil
			case clusterID4:
				return cluster4, nil
			default:
				return nil, fmt.Errorf("error")
			}

		},
		MockAllIDs: func() []string {
			return []string{clusterID1, clusterID2, clusterID3, clusterID4}
		},
	}

	setupAllocator = func() *Manager {
		return NewManager(mockClusterManager)
	}
)

func Test_AllocateCluster(t *testing.T) {
	tests := []struct {
		name            string
		input           map[string]string
		possibleExpects []string
		expectErr       bool
	}{
		{
			name: "exactly 1 match",
			input: map[string]string{
				"region": "us-south-1",
				"tier":   "tier-2",
				"type":   "type-a",
			},
			possibleExpects: []string{clusterID2},
			expectErr:       false,
		},
		{
			name: "0 matches",
			input: map[string]string{
				"region": "vietnam",
			},
			possibleExpects: []string{},
			expectErr:       true,
		},
		{
			name: "multiple matches",
			input: map[string]string{
				"type": "type-a",
			},
			possibleExpects: []string{clusterID1, clusterID2, clusterID3},
			expectErr:       false,
		},
	}

	allocationManager := setupAllocator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := allocationManager.AllocateCluster(tt.input)
			if tt.expectErr {
				require.NotNil(t, err)
				require.Equal(t, "", actual)
			} else {
				require.Nil(t, err)
				require.Contains(t, tt.possibleExpects, actual)
			}
		})
	}
}

func Test_tagMatch(t *testing.T) {
	type input struct {
		a map[string]string
		b map[string]string
	}

	tests := []struct {
		name   string
		input  input
		expect bool
	}{
		{
			name: "equal sets",
			input: input{
				a: map[string]string{
					"region": "us-east-1",
					"type":   "block-storage",
				},
				b: map[string]string{
					"region": "us-east-1",
					"type":   "block-storage",
				},
			},
			expect: true,
		},
		{
			name: "a is strict subset of b",
			input: input{
				a: map[string]string{
					"region": "us-east-1",
					"type":   "block-storage",
				},
				b: map[string]string{
					"region": "us-east-1",
					"type":   "block-storage",
					"vendor": "aws",
				},
			},
			expect: true,
		},
		{
			name: "b is strict subset of a",
			input: input{
				a: map[string]string{
					"region": "us-east-1",
					"type":   "block-storage",
					"vendor": "aws",
				},
				b: map[string]string{
					"region": "us-east-1",
					"type":   "block-storage",
				},
			},
			expect: false,
		},
		{
			name: "a and b are disjoint sets",
			input: input{
				a: map[string]string{
					"region": "us-east-1",
					"type":   "block-storage",
				},
				b: map[string]string{
					"region": "us-south-1",
					"type":   "shared-nfs",
				},
			},
			expect: false,
		},
		{
			name: "a and b are empty sets",
			input: input{
				a: map[string]string{},
				b: map[string]string{},
			},
			expect: true,
		},
		{
			name: "a is empty set, b is non-empty set",
			input: input{
				a: map[string]string{},
				b: map[string]string{
					"region": "us-south-1",
					"type":   "shared-nfs",
				},
			},
			expect: true,
		},
		{
			name: "a is non-empty set, b is empty set",
			input: input{
				a: map[string]string{
					"region": "us-east-1",
					"type":   "block-storage",
				},
				b: map[string]string{},
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tagMatch(tt.input.a, tt.input.b)
			require.Equal(t, tt.expect, actual)
		})
	}
}
