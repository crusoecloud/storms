package krusoe

import (
	"errors"

	"gopkg.in/yaml.v2"
)

var errConfigParse = errors.New("failed to parse config")

//nolint:tagliatelle // using snake case for YAML
type Config struct {
	APIKey string `yaml:"api_key"`
}

func ParseConfig(bytes []byte, cfg *Config) error {
	if err := yaml.Unmarshal(bytes, cfg); err != nil {
		return errConfigParse
	}

	return nil
}
