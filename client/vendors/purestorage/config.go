package purestorage

//nolint:tagliatelle // using snake case for YAML
type ClientConfig struct {
	Endpoints []string `yaml:"endpoints"`
	AuthToken string   `yaml:"auth_token"`
	Password  string   `yaml:"password"`
}
