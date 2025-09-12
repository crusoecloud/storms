package configs

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func Test_Default(t *testing.T) {
	args := []string{"--config", "testdata/emptyConfig.yaml"}
	mockCmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := Parse(cmd); err != nil {
				t.Error(err)
			}

			require.Equal(t, grpcPortDefault, Get().GrpcPort)
			require.Equal(t, localIPDefault, Get().LocalIP)
			require.Nil(t, Get().AuthInfo)
			require.Equal(t, clusterFileDefault, Get().ClusterFile)
			require.Equal(t, resourceFileDefault, Get().ResourceFile)

			return nil
		},
	}

	AddFlags(mockCmd)
	mockCmd.SetArgs(args)
	if err := mockCmd.Execute(); err != nil {
		t.Error(err)
	}
}

func Test_Invalid(t *testing.T) {
	args := []string{"--config", "testdata/invalidConfig.yaml"}
	mockCmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			return Parse(cmd)
		},
	}

	AddFlags(mockCmd)
	mockCmd.SetArgs(args)
	if err := mockCmd.Execute(); err == nil {
		t.Error("expected failure with invalid config file")
	}
}

// Currently broken
func Test_Parse(t *testing.T) {
	args := []string{"--config", "testdata/validConfig.yaml"}
	mockCmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := Parse(cmd); err != nil {
				t.Error(err)
			}

			require.Equal(t, 8888, Get().GrpcPort)
			require.Equal(t, "127.127.127.127", Get().LocalIP)
			require.Nil(t, Get().AuthInfo)
			require.Equal(t, "/some_dir/cluster_file.yaml", Get().ClusterFile)
			require.Equal(t, "/some_dir/resource_file.yaml", Get().ResourceFile)

			return nil
		},
	}

	AddFlags(mockCmd)
	mockCmd.SetArgs(args)
	if err := mockCmd.Execute(); err != nil {
		t.Error(err)
	}
}
