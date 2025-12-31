package configs

import (
	"fmt"
	"os"

	"gitlab.com/crusoeenergy/island/storage/storms/storms/internal/service/cluster"
	"gopkg.in/yaml.v2"
)

type ClustersConfig struct {
	Clusters []*cluster.Config `yaml:"clusters"`
}

func LoadClusterConfig(filename string) (*ClustersConfig, error) {
	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML into struct
	var clustersConfig ClustersConfig
	if err := yaml.Unmarshal(data, &clustersConfig); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &clustersConfig, nil
}
