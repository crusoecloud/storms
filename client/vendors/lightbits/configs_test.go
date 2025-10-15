package lightbits

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ParseConfig(t *testing.T) {
	filename := "testdata/valid.yaml"
	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		t.FailNow()
	}

	// Parse YAML into struct
	var clientConfig ClientConfig
	if err := ParseConfig(data, &clientConfig); err != nil {
		t.FailNow()
	}

	require.Equal(t, "this_is_an_auth_token", clientConfig.AuthToken)
	require.Len(t, clientConfig.AddrsStrs, 2)
	require.Equal(t, "1.1.1.1:1", clientConfig.AddrsStrs[0])
	require.Equal(t, "2.2.2.2:2", clientConfig.AddrsStrs[1])

	require.Equal(t, 3, clientConfig.ReplicationFactor)
	require.Equal(t, "unit-test", clientConfig.ProjectName)
}
