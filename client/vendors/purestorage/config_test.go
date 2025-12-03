package purestorage

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

	require.Equal(t, "test-api-token-12345", clientConfig.AuthToken)
	require.Equal(t, "testuser", clientConfig.Username)
	require.Equal(t, "testpass", clientConfig.Password)
	require.Equal(t, "2.45", clientConfig.APIVersion)
	require.Len(t, clientConfig.Endpoints, 2)
	require.Equal(t, "10.0.0.1", clientConfig.Endpoints[0])
	require.Equal(t, "10.0.0.2", clientConfig.Endpoints[1])
}

func Test_ParseConfig_InvalidYAML(t *testing.T) {
	invalidYAML := []byte(`
endpoints: ['10.0.0.1'
auth_token: "missing-bracket"
`)

	var clientConfig ClientConfig
	err := ParseConfig(invalidYAML, &clientConfig)
	require.Error(t, err)
}

func Test_ParseConfig_EmptyData(t *testing.T) {
	var clientConfig ClientConfig
	err := ParseConfig([]byte{}, &clientConfig)
	require.NoError(t, err) // Empty YAML should parse successfully with zero values
	require.Empty(t, clientConfig.Endpoints)
	require.Empty(t, clientConfig.AuthToken)
	require.Empty(t, clientConfig.Username)
	require.Empty(t, clientConfig.Password)
	require.Empty(t, clientConfig.APIVersion)
}

func Test_ParseConfig_NilPointer(t *testing.T) {
	data := []byte(`endpoints: ['10.0.0.1']`)
	// This should panic, so we expect a panic recovery
	defer func() {
		if r := recover(); r != nil {
			// Expected panic due to nil pointer
			require.NotNil(t, r)
		}
	}()

	// This will panic, which is expected behavior
	_ = ParseConfig(data, nil)
	// If we get here without panic, the test should fail
	t.Fatal("Expected panic when passing nil pointer")
}
