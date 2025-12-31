package cluster

import (
	"github.com/rs/zerolog/log"
	"gitlab.com/crusoeenergy/island/storage/storms/client"
)

//nolint:tagliatelle // using snake case for YAML
type Config struct {
	Vendor       string                 `yaml:"vendor"`
	ClusterID    string                 `yaml:"cluster_id"`
	AffinityTags map[string]string      `yaml:"affinity_tags"`
	VendorConfig map[string]interface{} `yaml:"vendor_config"` // This is for vendor-specific configuration.
}

type Cluster struct {
	Config *Config
	Client client.Client
}

func NewCluster(cfg *Config) (*Cluster, error) {
	c, err := client.NewClient(cfg.Vendor, cfg.VendorConfig)
	if err != nil {
		log.Err(err).Str("cluster_id", cfg.ClusterID).Msg("failed to create new client for cluster")
	}

	return &Cluster{
		Config: cfg,
		Client: c,
	}, nil
}
