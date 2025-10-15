package lightbits

import (
	"errors"

	"gopkg.in/yaml.v2"
)

var errConfigParse = errors.New("failed to parse config")

//nolint:tagliatelle // using snake case for YAML
type ClientConfig struct {
	AddrsStrs         []string `yaml:"addr_strs"`
	AuthToken         string   `yaml:"auth_token"`
	ProjectName       string   `yaml:"project_name"`
	ReplicationFactor int      `yaml:"replication_factor"`
}

func ParseConfig(bytes []byte, cfg *ClientConfig) error {
	if err := yaml.Unmarshal(bytes, cfg); err != nil {
		return errConfigParse
	}

	return nil
}
