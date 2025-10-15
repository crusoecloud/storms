package cluster

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"gitlab.com/crusoeenergy/island/storage/storms/client"
	testutil "gitlab.com/crusoeenergy/island/storage/storms/internal/service/testutil"
)

var (
	clusterID1   = "b22de56f-600f-4fa1-99ba-8f045e905f8e"
	client1      = &testutil.MockClient{}
	clusterID2   = "2ecbb15a-3a55-4085-856a-296565f34f40"
	client2      = &testutil.MockClient{}
	setupManager = func() *InMemoryManager {
		manager := &InMemoryManager{
			clients: map[string]client.Client{
				clusterID1: client1,
				clusterID2: client2,
			},
		}

		return manager
	}
)

func Test_Set(t *testing.T) {
	type input struct {
		clusterID string
		c         client.Client
	}

	tests := []struct {
		name      string
		input     input
		expectErr bool
	}{
		{
			name: "valid",
			input: input{
				clusterID: uuid.NewString(),
				c:         &testutil.MockClient{},
			},
			expectErr: false,
		},
		{
			name: "duplicate",
			input: input{
				clusterID: clusterID1,
				c:         client1,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := setupManager()
			err := manager.Set(tt.input.clusterID, tt.input.c)
			if tt.expectErr {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
		})
	}
}

func Test_Remove(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "valid",
			input:     clusterID1,
			expectErr: false,
		},
		{
			name:      "invalid cluster id",
			input:     uuid.NewString(),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		manager := setupManager()
		t.Run(tt.name, func(t *testing.T) {
			err := manager.Remove(tt.input)
			if tt.expectErr {
				require.Error(t, err)

				return
			}
			require.NoError(t, err)
		})
	}
}

func Test_Get(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expect    client.Client
		expectErr bool
	}{
		{
			name:      "valid #1",
			input:     clusterID1,
			expect:    client1,
			expectErr: false,
		},
		{
			name:      "valid #2",
			input:     clusterID2,
			expect:    client2,
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
			switch tt.input {
			case clusterID1:
				require.Equal(t, client1, actual)
			case clusterID2:
				require.Equal(t, client2, actual)
			}
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
