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

	require.NotNil(t, cfg.Clusters[0].VendorConfig)

	require.Equal(t, cfg.Clusters[1].Vendor, "lightbits")
	require.NotNil(t, cfg.Clusters[1].VendorConfig)

	
}
