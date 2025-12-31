package cluster

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"

	clientmocks "gitlab.com/crusoeenergy/island/storage/storms/client/mocks"
)

var (
	vendor1       = "vendor-a"
	clusterID1    = "b22de56f-600f-4fa1-99ba-8f045e905f8e"
	client1       = &clientmocks.MockClient{}
	affinityTags1 = map[string]string{"region": "us-east-1"}

	vendor2       = "vendor-b"
	clusterID2    = "2ecbb15a-3a55-4085-856a-296565f34f40"
	client2       = &clientmocks.MockClient{}
	affinityTags2 = map[string]string{"region": "us-south-1"}

	cluster1 = &Cluster{
		Config: &Config{
			Vendor:       vendor1,
			ClusterID:    clusterID1,
			AffinityTags: affinityTags1,
			VendorConfig: nil,
		},
		Client: client1,
	}
	cluster2 = &Cluster{
		Config: &Config{
			Vendor:       vendor2,
			ClusterID:    clusterID2,
			AffinityTags: affinityTags2,
			VendorConfig: nil,
		},
		Client: client2,
	}

	setupManager = func() *InMemoryManager {
		manager := &InMemoryManager{
			clusters: map[string]*Cluster{
				clusterID1: cluster1,
				clusterID2: cluster2,
			},
		}

		return manager
	}
)

func Test_Set(t *testing.T) {
	type input struct {
		clusterID string
		c         *Cluster
	}

	tests := []struct {
		name      string
		inputs    []input
		expectErr bool
	}{
		{
			name: "valid",
			inputs: []input{
				{
					clusterID: clusterID1,
					c:         cluster1,
				},
			},
			expectErr: false,
		},
		{
			name: "duplicate",
			inputs: []input{
				{
					clusterID: clusterID1,
					c:         cluster1,
				},
				{
					clusterID: clusterID1,
					c:         cluster1,
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := setupManager()
			var multiErr error
			for _, input := range tt.inputs {
				err := manager.Set(input.clusterID, input.c)
				if err != nil {
					multiErr = multierr.Append(multiErr, err)
				}
			}
			if tt.expectErr {
				require.Error(t, multiErr)
			} else {
				require.NoError(t, multiErr)
			}

		})
	}
}

func Test_Remove(t *testing.T) {
	tests := []struct {
		name      string
		inputs    []string
		expectErr bool
	}{
		{
			name:      "valid",
			inputs:    []string{clusterID1, clusterID2},
			expectErr: false,
		},
		{
			name:      "invalid cluster id",
			inputs:    []string{uuid.NewString()},
			expectErr: true,
		},
		{
			name:      "remove cluster twice",
			inputs:    []string{clusterID1, clusterID1},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := setupManager()
			var multiErr error
			for _, input := range tt.inputs {
				err := manager.Remove(input)
				if err != nil {
					multiErr = multierr.Append(multiErr, err)
				}
			}
			if tt.expectErr {
				require.Error(t, multiErr)
			} else {
				require.NoError(t, multiErr)
			}
		})
	}
}

func Test_Get(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expect    *Cluster
		expectErr bool
	}{
		{
			name:      "valid #1",
			input:     clusterID1,
			expect:    cluster1,
			expectErr: false,
		},
		{
			name:      "valid #2",
			input:     clusterID2,
			expect:    cluster2,
			expectErr: false,
		},
		{
			name:      "invalid cluster id",
			input:     uuid.NewString(),
			expect:    nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		manager := setupManager()
		t.Run(tt.name, func(t *testing.T) {
			actual, err := manager.Get(tt.input)
			if tt.expectErr {
				require.Error(t, err)
				require.Nil(t, actual)

				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.input, actual.Config.ClusterID)
		})
	}
}

func Test_AllIDs(t *testing.T) {
	tests := []struct {
		name   string
		expect []string
	}{
		{
			name:   "valid",
			expect: []string{clusterID1, clusterID2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := setupManager()
			actual := manager.AllIDs()
			require.ElementsMatch(t, tt.expect, actual)
		})
	}

}

func Test_Count(t *testing.T) {
	tests := []struct {
		name   string
		expect int
	}{
		{
			name:   "valid",
			expect: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := setupManager()
			actual := manager.Count()
			require.Equal(t, tt.expect, actual)
		})
	}

}
