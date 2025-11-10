package configs

import (
	"fmt"
	"strings"

	"github.com/ory/viper"
	"github.com/spf13/cobra"
)

const (
	configFileFlag     = "config"
	configFileDefault  = "dev/storms.yaml"
	localIPFlag        = "local_ip"
	localIPDefault     = "127.0.0.1"
	grpcPortFlag       = "grpc_port"
	grpcPortDefault    = 55554
	clusterFileFlag    = "cluster_file"
	clusterFileDefault = "dev/clusters.yaml"
)

var appConfig *AppConfig //nolint:gochecknoglobals // using a global to avoid passing large config struct around

type AppConfig struct {
	// filepath for yaml config
	ConfigFile string `mapstructure:"config"`
	// local IP of host
	LocalIP string `mapstructure:"local_ip"`
	// port for listening gRPC request
	GrpcPort int `mapstructure:"grpc_port"`
	// // information needed for HTTP and gRPC authentication
	// AuthInfo *rpc.AuthInfo `mapstructure:"auth_info"`
	// cluster file
	ClusterFile string `mapstructure:"cluster_file"`
}

func Parse(cmd *cobra.Command) error {
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return fmt.Errorf("failed to bind CLI config flags: %w", err)
	}

	viper.SetEnvPrefix("storms")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	setViperDefaults()

	configFile := viper.GetString(configFileFlag)
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read in config: %w", err)
	}

	appConfig = &AppConfig{}
	if err := viper.Unmarshal(appConfig); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return nil
}

func Get() AppConfig {
	if appConfig == nil {
		return AppConfig{}
	}

	return *appConfig
}

func setViperDefaults() {
	mustBindEnv(configFileFlag)
	viper.SetDefault(configFileFlag, configFileDefault)
	mustBindEnv(localIPFlag)
	viper.SetDefault(localIPFlag, localIPDefault)
	mustBindEnv(grpcPortFlag)
	viper.SetDefault(grpcPortFlag, grpcPortDefault)
	mustBindEnv(clusterFileFlag)
	viper.SetDefault(clusterFileFlag, clusterFileDefault)

	// Bind more env vars here.
}

func mustBindEnv(flag string) {
	if err := viper.BindEnv(flag); err != nil {
		panic(err)
	}
}

// AddFlags attaches the CLI flags the Region Coordinator needs to the provided command.
func AddFlags(cmd *cobra.Command) {
	cmd.Flags().String(configFileFlag, configFileDefault,
		"Filepath of StorMS config (yaml) file")
}

func ApplyConfig(cmd *cobra.Command, _ []string) error {
	if err := Parse(cmd); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return nil
}
