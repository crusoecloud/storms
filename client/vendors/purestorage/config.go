package purestorage

import (
	"errors"

	"gopkg.in/yaml.v2"
)

var errConfigParse = errors.New("failed to parse config")

//nolint:tagliatelle // using snake case for YAML
type ClientConfig struct {
	Endpoints  []string `yaml:"endpoints"`
	AuthToken  string   `yaml:"auth_token"`
	Username   string   `yaml:"username"`
	Password   string   `yaml:"password"`
	APIVersion string   `yaml:"api_version"`
}

func ParseConfig(bytes []byte, cfg *ClientConfig) error {
	if err := yaml.Unmarshal(bytes, cfg); err != nil {
		return errConfigParse
	}

	return nil
}
