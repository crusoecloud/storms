package resource

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var (
	clusterID1  = "686802c1-f03c-40bf-91aa-2e7f639e1002"
	clusterID2  = "b0147f5d-97b6-438c-8214-3c77f0419db8"
	resourceID1 = "1c16841a-e3b2-48de-87d6-8292bca5b246"
	resourceID2 = "9628b6ac-58aa-43e9-9601-3e8536b60bb4"
	resourceID3 = "f0477318-65e2-4391-9048-6b81682e76a1"

	setupManager = func() *InMemoryManager {
		manager := &InMemoryManager{
			resourceIDToResourceMetadata: map[string]*Resource{
				resourceID1: {
					ID:           resourceID1,
					ClusterID:    clusterID1,
					ResourceType: TypeVolume,
				},
				resourceID2: {
					ID:           resourceID2,
					ClusterID:    clusterID2,
					ResourceType: TypeVolume,
				},
				resourceID3: {
					ID:           resourceID3,
					ClusterID:    clusterID2,
					ResourceType: TypeSnapshot,
				},
			},
		}

		return manager
	}
)

func Test_Map(t *testing.T) {
	tests := []struct {
		name      string
		resources []*Resource
	}{
		{
			name: "map single",
			resources: []*Resource{
				{
					ID:           resourceID1,
					ResourceType: TypeVolume,
					ClusterID:    clusterID1,
				},
			},
		},
		{
			name: "map multiple unique",
			resources: []*Resource{
				{
					ID:           resourceID1,
					ResourceType: TypeVolume,
					ClusterID:    clusterID1,
				},
				{
					ID:           resourceID2,
					ResourceType: TypeVolume,
					ClusterID:    clusterID2,
				},
			},
		},
		{
			name: "map multiple with duplicate ids",
			resources: []*Resource{
				{
					ID:           resourceID1,
					ResourceType: TypeVolume,
					ClusterID:    clusterID1,
				},
				{
					ID:           resourceID1,
					ResourceType: TypeVolume,
					ClusterID:    clusterID2,
				},
			},
		},
	}

	for _, tt := range tests {
		manager := setupManager()
		for _, r := range tt.resources {
			err := manager.Map(r)
			require.NoError(t, err)
		}
	}
}

func Test_Unmap(t *testing.T) {
	type input struct {
		resourceID string
	}
	tests := []struct {
		name   string
		inputs []input
	}{
		{
			name: "unmap single",
			inputs: []input{
				{
					resourceID: resourceID1,
				},
			},
		},
		{
			name: "unmap multiple unique",
			inputs: []input{
				{
					resourceID: resourceID1,
				},
				{
					resourceID: resourceID2,
				},
			},
		},
		{
			name: "unmap multiple duplicate",
			inputs: []input{
				{
					resourceID: resourceID1,
				},
				{
					resourceID: resourceID1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := setupManager()
			for _, input := range tt.inputs {
				err := mapper.Unmap(input.resourceID)
				require.NoError(t, err)
			}
		})
	}
}

func Test_OwnerCluster(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expect    string
		expectErr bool
	}{
		{
			name:      "valid",
			input:     resourceID1,
			expect:    clusterID1,
			expectErr: false,
		},
		{
			name:      "valid #2",
			input:     resourceID3,
			expect:    clusterID2,
			expectErr: false,
		},
		{
			name:      "unmapped resource",
			input:     uuid.NewString(),
			expect:    "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := setupManager()
			actual, err := mapper.GetResourceCluster(tt.input)
			if tt.expectErr {
				require.Error(t, err)

				return
			}
			require.Equal(t, tt.expect, actual)
		})
	}
}
