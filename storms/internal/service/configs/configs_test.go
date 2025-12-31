package configs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_LoadClusterConfig_Invalid(t *testing.T) {
	filename := "testdata/invalidClusters.yaml"
	_, err := LoadClusterConfig(filename)
	require.Error(t, err)
}

func Test_LoadClusterConfig_Valid(t *testing.T) {
	filename := "testdata/clusters.yaml"
	cfg, err := LoadClusterConfig(filename)
	require.NoError(t, err)

	require.Equal(t, len(cfg.Clusters), 2)

	require.Equal(t, cfg.Clusters[0].Vendor, "krusoe")
	require.Equal(t, cfg.Clusters[0].ClusterID, "b5cc813f-3a10-46ea-be93-b573e4a05ea1")
	require.Equal(t, cfg.Clusters[0].AffinityTags,
		map[string]string{
			"region": "us-east-1",
			"type":   "totally-not-nvme",
		},
	)
	require.NotNil(t, cfg.Clusters[0].VendorConfig)

	require.Equal(t, cfg.Clusters[1].Vendor, "lightbits")
	require.Equal(t, cfg.Clusters[1].ClusterID, "eee4019a-3d0f-4fea-8142-a5a8d0a3c20a")
	require.Equal(t, cfg.Clusters[1].AffinityTags,
		map[string]string{
			"region": "us-south-1",
			"type":   "not-nvme",
		},
	)
	require.NotNil(t, cfg.Clusters[1].VendorConfig)
}
