package resource

// import (
// 	"testing"

// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/require"
// )

// var (
// 	clusterID1  = "686802c1-f03c-40bf-91aa-2e7f639e1002"
// 	clusterID2  = "b0147f5d-97b6-438c-8214-3c77f0419db8"
// 	resourceID1 = "1c16841a-e3b2-48de-87d6-8292bca5b246"
// 	resourceID2 = "9628b6ac-58aa-43e9-9601-3e8536b60bb4"
// 	resourceID3 = "f0477318-65e2-4391-9048-6b81682e76a1"

// 	setupMapper = func() *InMemoryMapper {
// 		mapper := &InMemoryMapper{
// 			resourceIDToClusterID: map[string]string{
// 				resourceID1: clusterID1,
// 				resourceID2: clusterID2,
// 				resourceID3: clusterID2,
// 			},
// 			clusterIDToResourceID: map[string]map[string]struct{}{
// 				clusterID1: {resourceID1: struct{}{}},
// 				clusterID2: {
// 					resourceID2: struct{}{},
// 					resourceID3: struct{}{},
// 				},
// 			},
// 		}
// 		mapper.Map(resourceID1, clusterID1)

// 		return mapper
// 	}
// )

// func Test_Map(t *testing.T) {

// 	type input struct {
// 		resourceID string
// 		clusterID  string
// 	}
// 	tests := []struct {
// 		name   string
// 		inputs []input
// 	}{
// 		{
// 			name: "map single",
// 			inputs: []input{
// 				{
// 					resourceID: resourceID1,
// 					clusterID:  clusterID1,
// 				},
// 			},
// 		},
// 		{
// 			name: "map multiple unique",
// 			inputs: []input{
// 				{
// 					resourceID: resourceID1,
// 					clusterID:  clusterID1,
// 				},
// 				{
// 					resourceID: resourceID2,
// 					clusterID:  clusterID2,
// 				},
// 			},
// 		},
// 		{
// 			name: "map multiple duplicate",
// 			inputs: []input{
// 				{
// 					resourceID: resourceID1,
// 					clusterID:  clusterID1,
// 				},
// 				{
// 					resourceID: resourceID1,
// 					clusterID:  clusterID1,
// 				},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mapper := NewInMemoryMapper()
// 			for _, input := range tt.inputs {
// 				err := mapper.Map(input.resourceID, input.clusterID)
// 				require.NoError(t, err)
// 			}
// 		})
// 	}
// }

// func Test_Unmap(t *testing.T) {
// 	type input struct {
// 		resourceID string
// 		clusterID  string
// 	}
// 	tests := []struct {
// 		name   string
// 		inputs []input
// 	}{
// 		{
// 			name: "unmap single",
// 			inputs: []input{
// 				{
// 					resourceID: resourceID1,
// 					clusterID:  clusterID1,
// 				},
// 			},
// 		},
// 		{
// 			name: "unmap multiple unique",
// 			inputs: []input{
// 				{
// 					resourceID: resourceID1,
// 					clusterID:  clusterID1,
// 				},
// 				{
// 					resourceID: resourceID2,
// 					clusterID:  clusterID2,
// 				},
// 			},
// 		},
// 		{
// 			name: "unmap multiple duplicate",
// 			inputs: []input{
// 				{
// 					resourceID: resourceID1,
// 					clusterID:  clusterID1,
// 				},
// 				{
// 					resourceID: resourceID1,
// 					clusterID:  clusterID1,
// 				},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mapper := setupMapper()
// 			for _, input := range tt.inputs {
// 				err := mapper.Unmap(input.resourceID)
// 				require.NoError(t, err)
// 			}
// 		})
// 	}
// }

// func Test_OwnerCluster(t *testing.T) {
// 	tests := []struct {
// 		name      string
// 		input     string
// 		expect    string
// 		expectErr bool
// 	}{
// 		{
// 			name:      "valid",
// 			input:     resourceID1,
// 			expect:    clusterID1,
// 			expectErr: false,
// 		},
// 		{
// 			name:      "unmapped resource",
// 			input:     uuid.NewString(),
// 			expect:    "",
// 			expectErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mapper := setupMapper()
// 			actual, err := mapper.OwnerCluster(tt.input)
// 			if tt.expectErr {
// 				require.Error(t, err)

// 				return
// 			}
// 			require.Equal(t, tt.expect, actual)
// 		})
// 	}
// }

// func Test_ResourceCount(t *testing.T) {
// 	tests := []struct {
// 		name   string
// 		expect int
// 	}{
// 		{
// 			name:   "valid",
// 			expect: 3,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mapper := setupMapper()
// 			actual := mapper.ResourceCount()
// 			require.Equal(t, tt.expect, actual)
// 		})
// 	}
// }

// func Test_ResourceCountByCluster(t *testing.T) {
// 	tests := []struct {
// 		name   string
// 		input  string
// 		expect int
// 	}{
// 		{
// 			name:   "valid #1",
// 			input:  clusterID1,
// 			expect: 1,
// 		},
// 		{
// 			name:   "valid #2",
// 			input:  clusterID2,
// 			expect: 2,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mapper := setupMapper()
// 			actual := mapper.ResourceCountByCluster(tt.input)
// 			require.Equal(t, tt.expect, actual)
// 		})
// 	}
// }

// func Test_GetAllClusterResources(t *testing.T) {
// 	tests := []struct {
// 		name   string
// 		expect map[string][]string
// 	}{
// 		{
// 			name: "valid",
// 			expect: map[string][]string{
// 				clusterID1: {resourceID1},
// 				clusterID2: {resourceID2, resourceID3},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mapper := setupMapper()
// 			actual := mapper.GetAllClusterResources()
// 			require.ElementsMatch(t, tt.expect[clusterID1], actual[clusterID1])
// 			require.ElementsMatch(t, tt.expect[clusterID2], actual[clusterID2])
// 		})
// 	}
// }
