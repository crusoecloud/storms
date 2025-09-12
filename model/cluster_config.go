package model

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type ClusterConfig struct {
	Clusters []Cluster `yaml:"clusters"`
}

type Cluster struct {
	Vendor   string `yaml:"vendor"`
	Endpoint string `yaml:"endpoint"`
}

func LoadClusterConfig(filename string) (*ClusterConfig, error) {
	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML into struct
	var clusterConfig ClusterConfig
	if err := yaml.Unmarshal(data, &clusterConfig); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &clusterConfig, nil
}
