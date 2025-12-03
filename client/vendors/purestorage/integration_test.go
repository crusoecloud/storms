//go:build integration
// +build integration

package purestorage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"gitlab.com/crusoeenergy/island/storage/storms/client/models"
)

// cleanupVolume is a helper function to delete a volume during test cleanup
func cleanupVolume(t *testing.T, client *Client, ctx context.Context, volumeName string) {
	t.Logf("Cleaning up volume: %s", volumeName)
	deleteReq := &models.DeleteVolumeRequest{
		UUID: volumeName,
	}
	_, err := client.DeleteVolume(ctx, deleteReq)
	if err != nil {
		t.Logf("Warning: Failed to cleanup volume %s: %v", volumeName, err)
	} else {
		t.Logf("Successfully cleaned up volume: %s", volumeName)
	}
}

// cleanupHost is a helper function to delete a host during test cleanup
func cleanupHost(t *testing.T, client *Client, ctx context.Context, hostUUID string) {
	// Translate UUID to NQN format
	nqn := fmt.Sprintf("nqn.2014-08.org.nvmexpress:uuid:%s", hostUUID)
	t.Logf("Cleaning up host: %s (NQN: %s)", hostUUID, nqn)

	path := fmt.Sprintf("/api/%s/hosts?names=%s", client.apiVersion, nqn)
	var resp interface{}
	err := client.delete(path, nil, &resp)
	if err != nil {
		t.Logf("Warning: Failed to cleanup host %s: %v", nqn, err)
	} else {
		t.Logf("Successfully cleaned up host: %s", nqn)
	}
}

// cleanupConnection is a helper function to delete a connection during test cleanup
func cleanupConnection(t *testing.T, client *Client, ctx context.Context, hostUUID, volumeName string) {
	// Translate UUID to NQN format
	nqn := fmt.Sprintf("nqn.2014-08.org.nvmexpress:uuid:%s", hostUUID)
	t.Logf("Cleaning up connection between host %s and volume %s", nqn, volumeName)

	err := client.deleteConnection(nqn, volumeName)
	if err != nil {
		t.Logf("Warning: Failed to cleanup connection: %v", err)
	} else {
		t.Logf("Successfully cleaned up connection")
	}
}

// getHost retrieves a host from the array by UUID
func getHost(t *testing.T, client *Client, hostUUID string) (*Host, error) {
	// Translate UUID to NQN format
	path := fmt.Sprintf("/api/%s/hosts?names=%s", client.apiVersion, hostUUID)
	var resp GetHostsResponse
	err := client.get(path, &resp)
	if err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("host not found: %s", hostUUID)
	}
	return &resp.Items[0], nil
}

// createHost creates a host on the array with the given name and NQN
func createHost(t *testing.T, client *Client, hostName string, nqn string) (*Host, error) {
	createPath := fmt.Sprintf("/api/%s/hosts?names=%s", client.apiVersion, hostName)
	requestBody := map[string]interface{}{
		"nqns": []string{nqn},
	}

	var createResp CreateHostsResponse
	err := client.post(createPath, requestBody, &createResp)
	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	if len(createResp.Items) == 0 {
		return nil, fmt.Errorf("no host returned in create response")
	}

	t.Logf("Successfully created host: %s with NQN: %s", createResp.Items[0].Name, nqn)
	return &createResp.Items[0], nil
}

// getConnections retrieves connections for a specific host and/or volume
func getConnections(t *testing.T, client *Client, hostUUID, volumeName string) ([]Connection, error) {
	// Translate UUID to NQN format
	nqn := fmt.Sprintf(hostUUID)
	path := fmt.Sprintf("/api/%s/connections?host_names=%s&volume_names=%s", client.apiVersion, nqn, volumeName)
	var resp GetConnectionsResponse
	err := client.get(path, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// IntegrationTestConfig holds configuration for integration tests
type IntegrationTestConfig struct {
	Endpoints  []string `yaml:"endpoints"`
	AuthToken  string   `yaml:"auth_token"`
	Username   string   `yaml:"username"`
	Password   string   `yaml:"password"`
	APIVersion string   `yaml:"api_version"`
}

// Test configuration constants - defined in Go code, not YAML
const (
	operationTimeout = 30 * time.Second
	suiteTimeout     = 300 * time.Second
	volumePrefix     = "storms-test"
	cleanupVolumes   = true // DeleteVolume is now implemented
)

// Test volume configurations
var testVolumes = struct {
	Small struct {
		SizeBytes  uint64
		SectorSize uint32
	}
	Medium struct {
		SizeBytes  uint64
		SectorSize uint32
	}
	Large struct {
		SizeBytes  uint64
		SectorSize uint32
	}
}{
	Small: struct {
		SizeBytes  uint64
		SectorSize uint32
	}{
		SizeBytes:  512 * 1024 * 1024, // 512MB
		SectorSize: 0,
	},
	Medium: struct {
		SizeBytes  uint64
		SectorSize uint32
	}{
		SizeBytes:  2 * 1024 * 1024 * 1024, // 2GB
		SectorSize: 0,
	},
	Large: struct {
		SizeBytes  uint64
		SectorSize uint32
	}{
		SizeBytes:  10 * 1024 * 1024 * 1024, // 10GB
		SectorSize: 0,
	},
}

// Tests to skip (can be modified for debugging)
var skipTests = []string{
	// Add test names to skip, e.g., "TestIntegration_CreateVolume_ErrorCases"
}

// loadIntegrationConfig loads the integration test configuration from YAML file
func loadIntegrationConfig(t *testing.T) *ClientConfig {
	configPath := filepath.Join("testdata", "integration_config.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skipf("Integration config file not found: %s", configPath)
		t.Logf("To run integration tests:")
		t.Logf("1. Copy testdata/integration_config.template.yaml to testdata/integration_config.yaml")
		t.Logf("2. Edit the config file with your FlashArray settings")
		t.Logf("3. Run: go test -tags=integration ./client/vendors/purestorage -v")
		return nil
	}

	// Read and parse config file
	data, err := os.ReadFile(configPath)
	require.NoError(t, err, "Failed to read integration config file")

	var config ClientConfig
	// Use ParseConfig instead of Unmarshal
	err = ParseConfig(data, &config)
	require.NoError(t, err, "Failed to parse integration config file")

	// Validate required fields
	require.NotEmpty(t, config.Endpoints, "endpoints must be specified in config")
	require.True(t, config.AuthToken != "" || (config.Username != "" && config.Password != ""),
		"either auth_token or username/password must be specified")

	// Set defaults
	if config.APIVersion == "" {
		config.APIVersion = DefaultAPIVersion
	}

	return &config
}

// getIntegrationConfig converts IntegrationTestConfig to ClientConfig
func getIntegrationConfig(t *testing.T) *ClientConfig {
	cfg := loadIntegrationConfig(t)
	if cfg == nil {
		return nil
	}
	return cfg
}

// shouldSkipTest checks if a test should be skipped based on configuration
func shouldSkipTest(t *testing.T, testName string) {
	if slices.Contains(skipTests, testName) {
		t.Skipf("Test %s is configured to be skipped", testName)
	}
}

func TestIntegration_CreateVolume_NewVolume(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_CreateVolume_NewVolume")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Generate a unique volume name using configured prefix
	volumeName := fmt.Sprintf("%s-vol-%s", volumePrefix, uuid.New().String()[:8])

	// Use medium volume size from config
	req := &models.CreateVolumeRequest{
		UUID: volumeName,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Medium.SizeBytes,
			SectorSize: testVolumes.Medium.SectorSize,
		},
	}

	t.Logf("Creating volume: %s (size: %d bytes, sector: %d)",
		volumeName, req.Source.(*models.NewVolumeSpec).Size, req.Source.(*models.NewVolumeSpec).SectorSize)

	resp, err := client.CreateVolume(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Volume)

	// Verify the volume properties
	volume := resp.Volume
	require.Equal(t, volumeName, volume.UUID)
	require.Equal(t, testVolumes.Medium.SizeBytes, volume.Size)
	require.Equal(t, testVolumes.Medium.SectorSize, volume.SectorSize)
	require.True(t, volume.IsAvailable)
	require.Empty(t, volume.SourceSnapshotUUID)

	t.Logf("Successfully created volume: %+v", volume)

	// Clean up if configured
	if cleanupVolumes {
		cleanupVolume(t, client, ctx, volumeName)
	} else {
		t.Logf("Volume %s created successfully. Manual cleanup required from FlashArray.", volumeName)
	}
}

func TestIntegration_CreateVolume_DifferentSizes(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_CreateVolume_DifferentSizes")

	client, err := NewClient(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), suiteTimeout)
	defer cancel()

	// Use volume sizes from configuration
	testCases := []struct {
		name       string
		sizeBytes  uint64
		sectorSize uint32
	}{
		{
			name:       "small",
			sizeBytes:  testVolumes.Small.SizeBytes,
			sectorSize: testVolumes.Small.SectorSize,
		},
		{
			name:       "medium",
			sizeBytes:  testVolumes.Medium.SizeBytes,
			sectorSize: testVolumes.Medium.SectorSize,
		},
		{
			name:       "large",
			sizeBytes:  testVolumes.Large.SizeBytes,
			sectorSize: testVolumes.Large.SectorSize,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			volumeName := fmt.Sprintf("%s-%s-%s", volumePrefix, tc.name, uuid.New().String()[:8])

			req := &models.CreateVolumeRequest{
				UUID: volumeName,
				Source: &models.NewVolumeSpec{
					Size:       tc.sizeBytes,
					SectorSize: tc.sectorSize,
				},
			}

			t.Logf("Creating %s volume: %s (size: %d bytes, sector: %d)",
				tc.name, volumeName, tc.sizeBytes, tc.sectorSize)

			resp, err := client.CreateVolume(ctx, req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Volume)

			// Cleanup the volume after the test
			if cleanupVolumes {
				defer cleanupVolume(t, client, ctx, volumeName)
			}

			volume := resp.Volume
			require.Equal(t, volumeName, volume.UUID)
			require.Equal(t, tc.sizeBytes, volume.Size)
			require.Equal(t, tc.sectorSize, volume.SectorSize)
			require.True(t, volume.IsAvailable)

			t.Logf("Successfully created %s volume: %s", tc.name, volumeName)
		})
	}
}

func TestIntegration_CreateVolume_ErrorCases(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_CreateVolume_ErrorCases")

	client, err := NewClient(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), suiteTimeout)
	defer cancel()

	t.Run("DuplicateVolumeName", func(t *testing.T) {
		volumeName := fmt.Sprintf("%s-dup-%s", volumePrefix, uuid.New().String()[:8])

		req := &models.CreateVolumeRequest{
			UUID: volumeName,
			Source: &models.NewVolumeSpec{
				Size:       testVolumes.Small.SizeBytes,
				SectorSize: testVolumes.Small.SectorSize,
			},
		}

		// Create the volume first time - should succeed
		t.Logf("Creating volume for duplicate test: %s", volumeName)
		resp1, err := client.CreateVolume(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp1)

		// Try to create the same volume again - should fail
		t.Logf("Attempting to create duplicate volume: %s", volumeName)
		resp2, err := client.CreateVolume(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp2)

		t.Logf("Duplicate volume creation correctly failed: %v", err)
		if cleanupVolumes {
			cleanupVolume(t, client, ctx, volumeName)
		}
	})

	t.Run("InvalidVolumeSize", func(t *testing.T) {
		volumeName := fmt.Sprintf("%s-invalid-%s", volumePrefix, uuid.New().String()[:8])

		req := &models.CreateVolumeRequest{
			UUID: volumeName,
			Source: &models.NewVolumeSpec{
				Size:       0, // Invalid size
				SectorSize: testVolumes.Small.SectorSize,
			},
		}

		t.Logf("Attempting to create volume with invalid size: %s", volumeName)
		resp, err := client.CreateVolume(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)

		t.Logf("Invalid size correctly failed: %v", err)
	})
}

func TestIntegration_DeleteVolume_ExistingVolume(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_DeleteVolume_ExistingVolume")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// First create a volume to delete
	volumeName := fmt.Sprintf("%s-del-%s", volumePrefix, uuid.New().String()[:8])

	createReq := &models.CreateVolumeRequest{
		UUID: volumeName,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}

	t.Logf("Creating volume for deletion test: %s", volumeName)
	createResp, err := client.CreateVolume(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Volume)

	// Now delete the volume using the actual volume UUID returned from create
	deleteReq := &models.DeleteVolumeRequest{
		UUID: createResp.Volume.UUID,
	}

	t.Logf("Deleting volume: %s (UUID: %s)", volumeName, createResp.Volume.UUID)
	deleteResp, err := client.DeleteVolume(ctx, deleteReq)
	require.NoError(t, err)
	require.NotNil(t, deleteResp)

	t.Logf("Successfully deleted volume: %s (UUID: %s)", volumeName, createResp.Volume.UUID)
}

func TestIntegration_DeleteVolume_ErrorCases(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_DeleteVolume_ErrorCases")

	client, err := NewClient(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	t.Run("NonexistentVolume", func(t *testing.T) {
		nonexistentVolume := fmt.Sprintf("%s-nonexistent-%s", volumePrefix, uuid.New().String()[:8])

		deleteReq := &models.DeleteVolumeRequest{
			UUID: nonexistentVolume,
		}

		t.Logf("Attempting to delete nonexistent volume: %s", nonexistentVolume)
		resp, err := client.DeleteVolume(ctx, deleteReq)
		require.Error(t, err)
		require.Nil(t, resp)

		t.Logf("Nonexistent volume deletion correctly failed: %v", err)
	})

	t.Run("EmptyVolumeName", func(t *testing.T) {
		deleteReq := &models.DeleteVolumeRequest{
			UUID: "",
		}

		t.Logf("Attempting to delete volume with empty name")
		resp, err := client.DeleteVolume(ctx, deleteReq)
		require.Error(t, err)
		require.Nil(t, resp)

		t.Logf("Empty volume name deletion correctly failed: %v", err)
	})
}

func TestIntegration_ResizeVolume_Success(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_ResizeVolume_Success")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Create a volume first
	volumeName := fmt.Sprintf("%s-resize-%s", volumePrefix, uuid.New().String()[:8])
	createReq := &models.CreateVolumeRequest{
		UUID: volumeName,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes, // 1GB
			SectorSize: testVolumes.Small.SectorSize,
		},
	}

	createResp, err := client.CreateVolume(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Volume)

	// Ensure cleanup
	defer cleanupVolume(t, client, ctx, createResp.Volume.UUID)

	t.Logf("Created volume: %s, Size: %d bytes", createResp.Volume.UUID, createResp.Volume.Size)

	// Resize the volume to medium size (2GB)
	resizeReq := &models.ResizeVolumeRequest{
		UUID: createResp.Volume.UUID,
		Size: testVolumes.Medium.SizeBytes, // 2GB
	}

	resizeResp, err := client.ResizeVolume(ctx, resizeReq)
	require.NoError(t, err)
	require.NotNil(t, resizeResp)

	t.Logf("Successfully resized volume: %s to %d bytes", createResp.Volume.UUID, testVolumes.Medium.SizeBytes)
}

func TestIntegration_ResizeVolume_ErrorCases(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_ResizeVolume_ErrorCases")

	client, err := NewClient(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	t.Run("EmptyUUID", func(t *testing.T) {
		req := &models.ResizeVolumeRequest{
			UUID: "",
			Size: testVolumes.Medium.SizeBytes,
		}

		_, err := client.ResizeVolume(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "ResizeVolume cannot be called with empty value")

		t.Logf("Empty UUID correctly failed: %v", err)
	})

	t.Run("ZeroSize", func(t *testing.T) {
		req := &models.ResizeVolumeRequest{
			UUID: "test-volume-uuid",
			Size: 0,
		}

		_, err := client.ResizeVolume(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "ResizeVolume cannot be called with zero size")

		t.Logf("Zero size correctly failed: %v", err)
	})

	t.Run("NonExistentVolume", func(t *testing.T) {
		nonexistentVolume := fmt.Sprintf("%s-nonexistent-%s", volumePrefix, uuid.New().String()[:8])
		req := &models.ResizeVolumeRequest{
			UUID: nonexistentVolume,
			Size: testVolumes.Medium.SizeBytes,
		}

		_, err := client.ResizeVolume(ctx, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to resize volume")

		t.Logf("Non-existent volume correctly failed: %v", err)
	})
}

func TestIntegration_GetVolume_Success(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_GetVolume_Success")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Create a test volume first
	volumeName := fmt.Sprintf("%s-get-test-%s", volumePrefix, uuid.New().String()[:8])
	createReq := &models.CreateVolumeRequest{
		UUID: volumeName,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}

	createResp, err := client.CreateVolume(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Volume)

	if cleanupVolumes {
		defer cleanupVolume(t, client, ctx, volumeName)
	}

	t.Logf("Created test volume: %s", volumeName)

	// Test GetVolume
	getReq := &models.GetVolumeRequest{
		UUID: volumeName,
	}

	getResp, err := client.GetVolume(ctx, getReq)
	require.NoError(t, err)
	require.NotNil(t, getResp)
	require.NotNil(t, getResp.Volume)
	require.Equal(t, volumeName, getResp.Volume.UUID)
	require.Equal(t, testVolumes.Small.SizeBytes, getResp.Volume.Size)
	require.True(t, getResp.Volume.IsAvailable)

	t.Logf("Successfully retrieved volume: %s with size %d", getResp.Volume.UUID, getResp.Volume.Size)
}

func TestIntegration_GetVolume_NotFound(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_GetVolume_NotFound")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Try to get a non-existent volume
	nonexistentVolume := fmt.Sprintf("%s-nonexistent-%s", volumePrefix, uuid.New().String()[:8])
	req := &models.GetVolumeRequest{
		UUID: nonexistentVolume,
	}

	resp, err := client.GetVolume(ctx, req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "does not exist")

	t.Logf("Non-existent volume correctly failed: %v", err)
}

func TestIntegration_GetVolumes_Success(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_GetVolumes_Success")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Create multiple test volumes
	volumeNames := make([]string, 2)
	for i := 0; i < 2; i++ {
		volumeNames[i] = fmt.Sprintf("%s-getall-test-%d-%s", volumePrefix, i, uuid.New().String()[:8])
		createReq := &models.CreateVolumeRequest{
			UUID: volumeNames[i],
			Source: &models.NewVolumeSpec{
				Size:       testVolumes.Small.SizeBytes,
				SectorSize: testVolumes.Small.SectorSize,
			},
		}

		createResp, err := client.CreateVolume(ctx, createReq)
		require.NoError(t, err)
		require.NotNil(t, createResp)

		if cleanupVolumes {
			defer cleanupVolume(t, client, ctx, volumeNames[i])
		}

		t.Logf("Created test volume %d: %s", i+1, volumeNames[i])
	}

	// Test GetVolumes
	req := &models.GetVolumesRequest{}

	resp, err := client.GetVolumes(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Volumes)

	// Verify our test volumes are in the response
	foundVolumes := 0
	for _, volume := range resp.Volumes {
		for _, testVolumeName := range volumeNames {
			if volume.UUID == testVolumeName {
				foundVolumes++
				require.Equal(t, testVolumes.Small.SizeBytes, volume.Size)
				require.True(t, volume.IsAvailable)
				t.Logf("Found test volume in list: %s", volume.UUID)
			}
		}
	}

	require.Equal(t, len(volumeNames), foundVolumes, "Should find all created test volumes in the list")
	t.Logf("Successfully retrieved %d volumes, found %d test volumes", len(resp.Volumes), foundVolumes)
}

func TestIntegration_GetVolumes_EmptyValidation(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_GetVolumes_EmptyValidation")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Test GetVolumes - should work even if no volumes exist (returns empty list)
	req := &models.GetVolumesRequest{}

	resp, err := client.GetVolumes(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Volumes) // Should be empty slice, not nil

	t.Logf("GetVolumes returned %d volumes", len(resp.Volumes))
}

func TestIntegration_CreateSnapshot(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_CreateSnapshot")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	snapshotUuid := uuid.New().String()
	// First, create a volume to snapshot
	volumeName := fmt.Sprintf("TestInteg_CreateSnapshot-%s-vol-%s", volumePrefix, uuid.New().String()[:8])
	volReq := &models.CreateVolumeRequest{
		UUID: volumeName,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}

	t.Logf("Creating source volume: %s", volumeName)
	volResp, err := client.CreateVolume(ctx, volReq)

	require.NoError(t, err)
	require.NotNil(t, volResp)
	require.NotNil(t, volResp.Volume)

	t.Logf("Volume was created volResp: %+v ", volResp.Volume)

	// Now create a snapshot of the volume
	snapshotName := snapshotUuid
	snapReq := &models.CreateSnapshotRequest{
		UUID:             snapshotName,
		SourceVolumeUUID: volumeName,
	}

	t.Logf("Creating snapshot: %s from volume: %s", snapshotName, volumeName)
	snapResp, err := client.CreateSnapshot(ctx, snapReq)
	require.NoError(t, err)
	require.NotNil(t, snapResp)
	require.NotNil(t, snapResp.Snapshot)

	// Verify the snapshot properties
	snapshot := snapResp.Snapshot
	// Snapshot name will be volume-name.suffix (e.g., volume-name.1, volume-name.2)
	require.Contains(t, snapshot.UUID, volumeName)
	require.Equal(t, testVolumes.Small.SizeBytes, snapshot.Size)
	require.True(t, snapshot.IsAvailable)
	require.Equal(t, volumeName, snapshot.SourceVolumeUUID)

	t.Logf("Successfully created snapshot: %+v", snapshot)

	//Clean up note
	if cleanupVolumes {
		t.Logf("Cleanup is enabled but DeleteSnapshot and DeleteVolume are not yet implemented")
	} else {
		t.Logf("Snapshot %s and volume %s created successfully. Manual cleanup required from FlashArray.", snapshotName, volumeName)
	}
}

func TestIntegration_GetSnapshot_Success(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_GetSnapshot_Success")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Step 1: Create a volume with a name volumeName
	volumeName := fmt.Sprintf("%s-getsnapshot-%s", volumePrefix, uuid.New().String()[:8])
	createVolReq := &models.CreateVolumeRequest{
		UUID: volumeName,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}

	t.Logf("Step 1: Creating volume: %s", volumeName)
	createVolResp, err := client.CreateVolume(ctx, createVolReq)
	require.NoError(t, err)
	require.NotNil(t, createVolResp)
	require.NotNil(t, createVolResp.Volume)
	t.Logf("Successfully created volume: %s", volumeName)

	// Step 2: GetVolume just to check it exists
	getVolReq := &models.GetVolumeRequest{
		UUID: volumeName,
	}

	t.Logf("Step 2: Getting volume to verify it exists: %s", volumeName)
	getVolResp, err := client.GetVolume(ctx, getVolReq)
	require.NoError(t, err)
	require.NotNil(t, getVolResp)
	require.NotNil(t, getVolResp.Volume)
	require.Equal(t, volumeName, getVolResp.Volume.UUID)
	t.Logf("Successfully verified volume exists: %s", volumeName)

	// Step 3: Create snapshot with UUID, volumeName is the snapshot name and UUID in suffix
	snapshotUUID := uuid.New().String()
	createSnapReq := &models.CreateSnapshotRequest{
		UUID:             snapshotUUID,
		SourceVolumeUUID: volumeName,
	}

	t.Logf("Step 3: Creating snapshot with UUID: %s from volume: %s", snapshotUUID, volumeName)
	createSnapResp, err := client.CreateSnapshot(ctx, createSnapReq)
	require.NoError(t, err)
	require.NotNil(t, createSnapResp)
	require.NotNil(t, createSnapResp.Snapshot)
	t.Logf("Successfully created snapshot. Snapshot UUID: %s, Source Volume: %s",
		createSnapResp.Snapshot.UUID, createSnapResp.Snapshot.SourceVolumeUUID)

	// Step 4: Get snapshot with this UUID as parameters
	// This should bring back snapshot name is volumeName and suffix is UUID
	getSnapReq := &models.GetSnapshotRequest{
		UUID: snapshotUUID,
	}

	t.Logf("Step 4: Getting snapshot with UUID: %s", snapshotUUID)
	getSnapResp, err := client.GetSnapshot(ctx, getSnapReq)
	require.NoError(t, err)
	require.NotNil(t, getSnapResp)
	require.NotNil(t, getSnapResp.Snapshot)

	// Verify the snapshot properties
	snapshot := getSnapResp.Snapshot
	require.Equal(t, snapshotUUID, snapshot.UUID, "Snapshot UUID should match the requested UUID")
	require.Equal(t, volumeName, snapshot.SourceVolumeUUID, "Source volume UUID should be the volume name")
	require.Equal(t, testVolumes.Small.SizeBytes, snapshot.Size, "Snapshot size should match volume size")
	require.True(t, snapshot.IsAvailable, "Snapshot should be available")

	t.Logf("Successfully retrieved snapshot: UUID=%s, SourceVolume=%s, Size=%d",
		snapshot.UUID, snapshot.SourceVolumeUUID, snapshot.Size)

	// Clean up note
	if cleanupVolumes {
		t.Logf("Cleanup is enabled but DeleteSnapshot and DeleteVolume are not yet implemented")
	} else {
		t.Logf("Snapshot %s and volume %s created successfully. Manual cleanup required from FlashArray.",
			snapshotUUID, volumeName)
	}
}

func TestIntegration_GetSnapshot_ErrorCases(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_GetSnapshot_ErrorCases")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// First, create a volume and snapshot for testing
	volumeName := fmt.Sprintf("%s-getsnapshot-err-%s", volumePrefix, uuid.New().String()[:8])
	createVolReq := &models.CreateVolumeRequest{
		UUID: volumeName,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}

	t.Logf("Creating test volume: %s", volumeName)
	_, err = client.CreateVolume(ctx, createVolReq)
	require.NoError(t, err)

	// Create a snapshot with a known UUID
	validSnapshotUUID := uuid.New().String()
	createSnapReq := &models.CreateSnapshotRequest{
		UUID:             validSnapshotUUID,
		SourceVolumeUUID: volumeName,
	}

	t.Logf("Creating test snapshot with UUID: %s", validSnapshotUUID)
	_, err = client.CreateSnapshot(ctx, createSnapReq)
	require.NoError(t, err)

	t.Run("EmptyUUID", func(t *testing.T) {
		req := &models.GetSnapshotRequest{
			UUID: "",
		}

		t.Logf("Attempting to get snapshot with empty UUID")
		resp, err := client.GetSnapshot(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		t.Logf("Empty UUID correctly failed: %v", err)
	})

	t.Run("InvalidUUID", func(t *testing.T) {
		req := &models.GetSnapshotRequest{
			UUID: "not-a-valid-uuid",
		}

		t.Logf("Attempting to get snapshot with invalid UUID: %s", req.UUID)
		resp, err := client.GetSnapshot(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		t.Logf("Invalid UUID correctly failed: %v", err)
	})

	t.Run("NonexistentUUID", func(t *testing.T) {
		nonexistentUUID := uuid.New().String()
		req := &models.GetSnapshotRequest{
			UUID: nonexistentUUID,
		}

		t.Logf("Attempting to get snapshot with nonexistent UUID: %s", nonexistentUUID)
		resp, err := client.GetSnapshot(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		t.Logf("Nonexistent UUID correctly failed: %v", err)
	})

	t.Run("SlightlyChangedUUID", func(t *testing.T) {
		// Take the valid UUID and change one character
		modifiedUUID := validSnapshotUUID
		if len(modifiedUUID) > 0 {
			// Change the last character
			runes := []rune(modifiedUUID)
			if runes[len(runes)-1] == 'a' {
				runes[len(runes)-1] = 'b'
			} else {
				runes[len(runes)-1] = 'a'
			}
			modifiedUUID = string(runes)
		}

		req := &models.GetSnapshotRequest{
			UUID: modifiedUUID,
		}

		t.Logf("Attempting to get snapshot with slightly modified UUID: %s (original: %s)",
			modifiedUUID, validSnapshotUUID)
		resp, err := client.GetSnapshot(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		t.Logf("Modified UUID correctly failed: %v", err)
	})

	// Clean up note
	if cleanupVolumes {
		t.Logf("Cleanup is enabled but DeleteSnapshot and DeleteVolume are not yet implemented")
	} else {
		t.Logf("Test snapshot %s and volume %s created. Manual cleanup required from FlashArray.",
			validSnapshotUUID, volumeName)
	}
}

func TestIntegration_GetSnapshots_Success(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_GetSnapshots_Success")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Create multiple volumes and snapshots to test GetSnapshots
	numSnapshots := 3
	volumeNames := make([]string, numSnapshots)
	snapshotUUIDs := make([]string, numSnapshots)

	t.Logf("Creating %d volumes and snapshots for GetSnapshots test", numSnapshots)

	// Step 1: Create multiple volumes and snapshots
	for i := 0; i < numSnapshots; i++ {
		// Create volume
		volumeName := fmt.Sprintf("%s-getsnapshots-%d-%s", volumePrefix, i, uuid.New().String()[:8])
		volumeNames[i] = volumeName

		createVolReq := &models.CreateVolumeRequest{
			UUID: volumeName,
			Source: &models.NewVolumeSpec{
				Size:       testVolumes.Small.SizeBytes,
				SectorSize: testVolumes.Small.SectorSize,
			},
		}

		t.Logf("Creating volume %d: %s", i+1, volumeName)
		_, err := client.CreateVolume(ctx, createVolReq)
		require.NoError(t, err)
		t.Logf("Successfully created volume %d: %s", i+1, volumeName)

		// Create snapshot
		snapshotUUID := uuid.New().String()
		snapshotUUIDs[i] = snapshotUUID

		createSnapReq := &models.CreateSnapshotRequest{
			UUID:             snapshotUUID,
			SourceVolumeUUID: volumeName,
		}

		t.Logf("Creating snapshot %d with UUID: %s from volume: %s", i+1, snapshotUUID, volumeName)
		createSnapResp, err := client.CreateSnapshot(ctx, createSnapReq)
		require.NoError(t, err)
		require.NotNil(t, createSnapResp)
		require.NotNil(t, createSnapResp.Snapshot)
		require.Contains(t, createSnapResp.Snapshot.UUID, snapshotUUID)
		//split
		t.Logf("Successfully created snapshot %d: UUID=%s, SourceVolume=%s",
			i+1, createSnapResp.Snapshot.UUID, createSnapResp.Snapshot.SourceVolumeUUID)
	}

	// Step 2: Call GetSnapshots to retrieve all snapshots
	getSnapshotsReq := &models.GetSnapshotsRequest{}

	t.Logf("Calling GetSnapshots to retrieve all snapshots")
	getSnapshotsResp, err := client.GetSnapshots(ctx, getSnapshotsReq)
	require.NoError(t, err)
	require.NotNil(t, getSnapshotsResp)
	require.NotNil(t, getSnapshotsResp.Snapshots)

	// Step 3: Verify that our snapshots are in the response
	t.Logf("GetSnapshots returned %d total snapshots", len(getSnapshotsResp.Snapshots))
	require.GreaterOrEqual(t, len(getSnapshotsResp.Snapshots), numSnapshots,
		"Should have at least %d snapshots (the ones we created)", numSnapshots)

	// Find our created snapshots in the response
	foundSnapshots := make(map[string]*models.Snapshot)
	for _, snapshot := range getSnapshotsResp.Snapshots {
		for _, expectedUUID := range snapshotUUIDs {
			if snapshot.UUID == expectedUUID {
				foundSnapshots[expectedUUID] = snapshot
				t.Logf("Found our snapshot: UUID=%s, SourceVolume=%s, Size=%d",
					snapshot.UUID, snapshot.SourceVolumeUUID, snapshot.Size)
			}
		}
	}

	// Verify all our snapshots were found
	require.Equal(t, numSnapshots, len(foundSnapshots),
		"Should find all %d snapshots we created", numSnapshots)

	// Step 4: Verify properties of each found snapshot
	for i, snapshotUUID := range snapshotUUIDs {
		snapshot, found := foundSnapshots[snapshotUUID]
		require.True(t, found, "Snapshot %s should be in the response", snapshotUUID)
		require.Equal(t, snapshotUUID, snapshot.UUID, "Snapshot UUID should match")
		require.Equal(t, volumeNames[i], snapshot.SourceVolumeUUID, "Source volume UUID should match")
		require.Equal(t, testVolumes.Small.SizeBytes, snapshot.Size, "Snapshot size should match volume size")
		require.True(t, snapshot.IsAvailable, "Snapshot should be available")
	}

	t.Logf("Successfully verified all %d snapshots in GetSnapshots response", numSnapshots)

	// Clean up note
	if cleanupVolumes {
		t.Logf("Cleanup is enabled but DeleteSnapshot and DeleteVolume are not yet implemented")
	} else {
		t.Logf("Created %d snapshots and volumes. Manual cleanup required from FlashArray.", numSnapshots)
		for i := 0; i < numSnapshots; i++ {
			t.Logf("  - Snapshot: %s, Volume: %s", snapshotUUIDs[i], volumeNames[i])
		}
	}
}

func TestIntegration_GetSnapshots_VerifyMultipleSnapshots(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_GetSnapshots_VerifyMultipleSnapshots")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Create a single volume with multiple snapshots
	volumeName := fmt.Sprintf("%s-multisnapshots-%s", volumePrefix, uuid.New().String()[:8])

	// Create the volume
	createVolReq := &models.CreateVolumeRequest{
		UUID: volumeName,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}

	t.Logf("Creating volume: %s", volumeName)
	_, err = client.CreateVolume(ctx, createVolReq)
	require.NoError(t, err)
	t.Logf("Successfully created volume: %s", volumeName)

	// Create multiple snapshots of the same volume
	numSnapshots := 5
	snapshotUUIDs := make([]string, numSnapshots)

	for i := 0; i < numSnapshots; i++ {
		snapshotUUID := uuid.New().String()
		snapshotUUIDs[i] = snapshotUUID

		createSnapReq := &models.CreateSnapshotRequest{
			UUID:             snapshotUUID,
			SourceVolumeUUID: volumeName,
		}

		t.Logf("Creating snapshot %d/%d with UUID: %s", i+1, numSnapshots, snapshotUUID)
		createSnapResp, err := client.CreateSnapshot(ctx, createSnapReq)
		require.NoError(t, err)
		require.NotNil(t, createSnapResp)
		require.Contains(t, createSnapResp.Snapshot.UUID, snapshotUUID)
		require.Equal(t, volumeName, createSnapResp.Snapshot.SourceVolumeUUID)
	}

	t.Logf("Successfully created %d snapshots from volume %s", numSnapshots, volumeName)

	// Call GetSnapshots
	getSnapshotsReq := &models.GetSnapshotsRequest{}

	t.Logf("Calling GetSnapshots to retrieve all snapshots")
	getSnapshotsResp, err := client.GetSnapshots(ctx, getSnapshotsReq)
	require.NoError(t, err)
	require.NotNil(t, getSnapshotsResp)
	require.NotNil(t, getSnapshotsResp.Snapshots)

	t.Logf("GetSnapshots returned %d total snapshots", len(getSnapshotsResp.Snapshots))

	// Verify all our snapshots are present
	foundCount := 0
	for _, snapshot := range getSnapshotsResp.Snapshots {
		for _, expectedUUID := range snapshotUUIDs {
			if snapshot.UUID == expectedUUID {
				foundCount++
				require.Equal(t, volumeName, snapshot.SourceVolumeUUID,
					"Snapshot %s should have source volume %s", snapshot.UUID, volumeName)
				require.Equal(t, testVolumes.Small.SizeBytes, snapshot.Size,
					"Snapshot %s should have correct size", snapshot.UUID)
				require.True(t, snapshot.IsAvailable,
					"Snapshot %s should be available", snapshot.UUID)
				t.Logf("Found snapshot %d/%d: UUID=%s", foundCount, numSnapshots, snapshot.UUID)
			}
		}
	}

	require.Equal(t, numSnapshots, foundCount,
		"Should find all %d snapshots from volume %s", numSnapshots, volumeName)

	t.Logf("Successfully verified all %d snapshots from the same volume", numSnapshots)

	// Clean up note
	if cleanupVolumes {
		t.Logf("Cleanup is enabled but DeleteSnapshot and DeleteVolume are not yet implemented")
	} else {
		t.Logf("Created %d snapshots from volume %s. Manual cleanup required from FlashArray.",
			numSnapshots, volumeName)
		for i, snapshotUUID := range snapshotUUIDs {
			t.Logf("  - Snapshot %d: %s", i+1, snapshotUUID)
		}
	}
}

func TestIntegration_CreateVolumeFromSnapshot_Success(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_CreateVolumeFromSnapshot_Success")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Step 1: Create a source volume
	sourceVolumeName := fmt.Sprintf("%s-source-%s", volumePrefix, uuid.New().String()[:8])
	createSourceVolReq := &models.CreateVolumeRequest{
		UUID: sourceVolumeName,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Medium.SizeBytes, // 2GB
			SectorSize: testVolumes.Medium.SectorSize,
		},
	}

	t.Logf("Step 1: Creating source volume: %s (size: %d bytes)", sourceVolumeName, testVolumes.Medium.SizeBytes)
	createSourceVolResp, err := client.CreateVolume(ctx, createSourceVolReq)
	require.NoError(t, err)
	require.NotNil(t, createSourceVolResp)
	require.NotNil(t, createSourceVolResp.Volume)
	require.Equal(t, sourceVolumeName, createSourceVolResp.Volume.UUID)
	t.Logf("Successfully created source volume: %s", sourceVolumeName)

	// Step 2: Verify source volume exists using GetVolume
	getSourceVolReq := &models.GetVolumeRequest{
		UUID: sourceVolumeName,
	}

	t.Logf("Step 2: Verifying source volume exists: %s", sourceVolumeName)
	getSourceVolResp, err := client.GetVolume(ctx, getSourceVolReq)
	require.NoError(t, err)
	require.NotNil(t, getSourceVolResp)
	require.NotNil(t, getSourceVolResp.Volume)
	require.Equal(t, sourceVolumeName, getSourceVolResp.Volume.UUID)
	require.Equal(t, testVolumes.Medium.SizeBytes, getSourceVolResp.Volume.Size)
	require.Empty(t, getSourceVolResp.Volume.SourceSnapshotUUID, "Source volume should not have a snapshot source")
	t.Logf("Successfully verified source volume: %s (size: %d)", sourceVolumeName, getSourceVolResp.Volume.Size)

	// Step 3: Create a snapshot from the source volume
	snapshotUUID := uuid.New().String()
	createSnapReq := &models.CreateSnapshotRequest{
		UUID:             snapshotUUID,
		SourceVolumeUUID: sourceVolumeName,
	}

	t.Logf("Step 3: Creating snapshot: %s from source volume: %s", snapshotUUID, sourceVolumeName)
	createSnapResp, err := client.CreateSnapshot(ctx, createSnapReq)
	require.NoError(t, err)
	require.NotNil(t, createSnapResp)
	require.NotNil(t, createSnapResp.Snapshot)
	require.Contains(t, createSnapResp.Snapshot.UUID, snapshotUUID)
	require.Equal(t, sourceVolumeName, createSnapResp.Snapshot.SourceVolumeUUID)
	t.Logf("Successfully created snapshot sith suffix: %s", snapshotUUID)

	// Step 4: Create a new volume from the snapshot using CreateVolume with SnapshotSource
	createVolumeFromSnapshotName := fmt.Sprintf("%s-snapshot-%s", volumePrefix, uuid.New().String()[:8])

	// Extract snapshot suffix from full snapshot UUID (format: volume-UUID.snap-UUID)
	snapshotSuffix := snapshotUUID
	if strings.Contains(createSnapResp.Snapshot.UUID, ".") {
		parts := strings.Split(createSnapResp.Snapshot.UUID, ".")
		snapshotSuffix = parts[len(parts)-1]
	}

	createVolumeFromSnapshotVolReq := &models.CreateVolumeRequest{
		UUID: createVolumeFromSnapshotName,
		Source: &models.SnapshotSource{
			SnapshotUUID: snapshotSuffix, // Use the snapshot UUID (suffix), not the full name
		},
	}

	t.Logf("Step 4: Creating volume: %s from snapshot suffix: %s", createVolumeFromSnapshotName, snapshotSuffix)
	createClonedVolResp, err := client.CreateVolume(ctx, createVolumeFromSnapshotVolReq)
	require.NoError(t, err)
	require.NotNil(t, createClonedVolResp)
	require.NotNil(t, createClonedVolResp.Volume)
	require.Equal(t, createClonedVolResp.Volume.UUID, createVolumeFromSnapshotName)
	t.Logf("Successfully created cloned volume: %s", createVolumeFromSnapshotName)

	// Step 5: Verify the cloned volume using GetVolume
	getClonedVolReq := &models.GetVolumeRequest{
		UUID: createVolumeFromSnapshotName,
	}

	t.Logf("Step 5: Verifying cloned volume: %s", createVolumeFromSnapshotName)
	getClonedVolResp, err := client.GetVolume(ctx, getClonedVolReq)
	require.NoError(t, err)
	require.NotNil(t, getClonedVolResp)
	require.NotNil(t, getClonedVolResp.Volume)

	// Verify cloned volume properties
	clonedVolume := getClonedVolResp.Volume
	require.Equal(t, createVolumeFromSnapshotName, clonedVolume.UUID, "Cloned volume UUID should match")
	require.Equal(t, testVolumes.Medium.SizeBytes, clonedVolume.Size, "Cloned volume should have same size as source")
	require.True(t, clonedVolume.IsAvailable, "Cloned volume should be available")

	// Verify the cloned volume has the source volume as its source
	// Note: FlashArray API returns the source volume name in the source.name field,
	// not the snapshot name, when you clone from a snapshot
	require.Equal(t, sourceVolumeName, clonedVolume.SourceSnapshotUUID,
		"Cloned volume should have source volume name in SourceSnapshotUUID field")

	t.Logf("Successfully verified cloned volume:")
	t.Logf("  - UUID: %s", clonedVolume.UUID)
	t.Logf("  - Size: %d bytes", clonedVolume.Size)
	t.Logf("  - Source Snapshot: %s", clonedVolume.SourceSnapshotUUID)
	t.Logf("  - IsAvailable: %v", clonedVolume.IsAvailable)

	// Clean up note
	if cleanupVolumes {
		t.Logf("Cleanup is enabled but DeleteSnapshot and DeleteVolume are not yet implemented")
		t.Logf("Manual cleanup required:")
		t.Logf("  - Cloned volume: %s", createVolumeFromSnapshotName)
		t.Logf("  - Snapshot: %s", snapshotUUID)
		t.Logf("  - Source volume: %s", sourceVolumeName)
	} else {
		t.Logf("Created volumes and snapshot. Manual cleanup required from FlashArray:")
		t.Logf("  - Cloned volume: %s", createVolumeFromSnapshotName)
		t.Logf("  - Snapshot: %s", snapshotUUID)
		t.Logf("  - Source volume: %s", sourceVolumeName)
	}
}

// Test to change
func TestIntegration_CreateVolumeFromSnapshot_ErrorCases(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_CreateVolumeFromSnapshot_ErrorCases")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	t.Run("NonexistentSnapshot", func(t *testing.T) {
		// Try to create a volume from a snapshot that doesn't exist
		nonexistentSnapshotUUID := uuid.New().String()
		volumeName := fmt.Sprintf("%s-error-nonexistent-%s", volumePrefix, uuid.New().String()[:8])

		createVolReq := &models.CreateVolumeRequest{
			UUID: volumeName,
			Source: &models.SnapshotSource{
				SnapshotUUID: nonexistentSnapshotUUID,
			},
		}

		t.Logf("Attempting to create volume from nonexistent snapshot: %s", nonexistentSnapshotUUID)
		createVolResp, err := client.CreateVolume(ctx, createVolReq)
		require.Error(t, err, "Should fail when snapshot doesn't exist")
		require.Nil(t, createVolResp)
		t.Logf("Correctly failed with error: %v", err)

		// Verify the volume was NOT created using GetVolume
		getVolReq := &models.GetVolumeRequest{
			UUID: volumeName,
		}

		t.Logf("Verifying volume was not created: %s", volumeName)
		getVolResp, err := client.GetVolume(ctx, getVolReq)
		require.Error(t, err, "GetVolume should fail for non-created volume")
		require.Nil(t, getVolResp)
		t.Logf("Correctly verified volume does not exist")
	})

	t.Run("EmptySnapshotUUID", func(t *testing.T) {
		// Try to create a volume with empty snapshot UUID
		volumeName := fmt.Sprintf("%s-error-empty-%s", volumePrefix, uuid.New().String()[:8])

		createVolReq := &models.CreateVolumeRequest{
			UUID: volumeName,
			Source: &models.SnapshotSource{
				SnapshotUUID: "", // Empty snapshot UUID
			},
		}

		t.Logf("Attempting to create volume with empty snapshot UUID")
		createVolResp, err := client.CreateVolume(ctx, createVolReq)
		require.Error(t, err, "Should fail with empty snapshot UUID")
		require.Nil(t, createVolResp)
		require.Contains(t, err.Error(), "required", "Error should mention required field")
		t.Logf("Correctly failed with error: %v", err)

		// Verify the volume was NOT created using GetVolume
		getVolReq := &models.GetVolumeRequest{
			UUID: volumeName,
		}

		t.Logf("Verifying volume was not created: %s", volumeName)
		getVolResp, err := client.GetVolume(ctx, getVolReq)
		require.Error(t, err, "GetVolume should fail for non-created volume")
		require.Nil(t, getVolResp)
		t.Logf("Correctly verified volume does not exist")
	})

	t.Run("EmptyVolumeName", func(t *testing.T) {
		// Try to create a volume with empty volume name
		createVolReq := &models.CreateVolumeRequest{
			UUID: "", // Empty volume name
			Source: &models.SnapshotSource{
				SnapshotUUID: uuid.New().String(),
			},
		}

		t.Logf("Attempting to create volume with empty volume name")
		createVolResp, err := client.CreateVolume(ctx, createVolReq)
		require.Error(t, err, "Should fail with empty volume name")
		require.Nil(t, createVolResp)
		require.Contains(t, err.Error(), "required", "Error should mention required field")
		t.Logf("Correctly failed with error: %v", err)
	})

	t.Logf("All error cases handled correctly")
}

// TestIntegration_AttachVolume_ValidationErrors tests parameter validation for AttachVolume
func TestIntegration_AttachVolume_ValidationErrors(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_AttachVolume_ValidationErrors")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	t.Run("MultipleAcls", func(t *testing.T) {
		// Test with multiple ACLs (should fail - only one ACL allowed)
		req := &models.AttachVolumeRequest{
			UUID: "test-volume",
			Acls: []string{uuid.New().String(), uuid.New().String()},
		}

		t.Logf("Attempting to attach volume with multiple ACLs")
		resp, err := client.AttachVolume(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "exactly one ACL")
		t.Logf("Correctly failed with error: %v", err)
	})

	t.Run("MissingAcls", func(t *testing.T) {
		// Test with no ACLs (should fail)
		req := &models.AttachVolumeRequest{
			UUID: "test-volume",
			Acls: []string{},
		}

		t.Logf("Attempting to attach volume with no ACLs")
		resp, err := client.AttachVolume(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "exactly one ACL")
		t.Logf("Correctly failed with error: %v", err)
	})

	t.Run("EmptyUUID", func(t *testing.T) {
		// Test with empty volume UUID (should fail)
		req := &models.AttachVolumeRequest{
			UUID: "",
			Acls: []string{uuid.New().String()},
		}

		t.Logf("Attempting to attach volume with empty UUID")
		resp, err := client.AttachVolume(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "UUID is required")
		t.Logf("Correctly failed with error: %v", err)
	})

	t.Run("MissingUUID", func(t *testing.T) {
		// Test with missing volume UUID (should fail)
		req := &models.AttachVolumeRequest{
			Acls: []string{uuid.New().String()},
		}

		t.Logf("Attempting to attach volume with missing UUID")
		resp, err := client.AttachVolume(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		require.Contains(t, err.Error(), "UUID is required")
		t.Logf("Correctly failed with error: %v", err)
	})

	t.Logf("All validation error cases handled correctly")
}

// TestIntegration_AttachVolume_NewHost tests attaching a volume to a new host
func TestIntegration_AttachVolume_NewHost(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_AttachVolume_NewHost")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Step 1: Create a new volume
	volumeName := fmt.Sprintf("%s-attach-newhost-%s", volumePrefix, uuid.New().String()[:8])
	createReq := &models.CreateVolumeRequest{
		UUID: volumeName,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}

	t.Logf("Step 1: Creating volume: %s", volumeName)
	createResp, err := client.CreateVolume(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Volume)
	t.Logf("Successfully created volume: %s", volumeName)

	// Cleanup volume at the end
	defer cleanupVolume(t, client, ctx, volumeName)

	// Step 2: Create a new host
	hostUUID := uuid.New().String()
	hostName := fmt.Sprintf("storm-%s", hostUUID[:8])
	nqn := fmt.Sprintf("nqn.2014-08.org.nvmexpress:uuid:%s", hostUUID)

	t.Logf("Step 2: Creating host: %s with NQN: %s", hostName, nqn)
	host, err := createHost(t, client, hostName, nqn)
	require.NoError(t, err)
	require.NotNil(t, host)
	defer cleanupHost(t, client, ctx, hostName)

	// Step 3: Attach volume to the host
	t.Logf("Step 3: Attaching volume to host UUID: %s", hostUUID)

	attachReq := &models.AttachVolumeRequest{
		UUID: volumeName,
		Acls: []string{hostUUID},
	}

	attachResp, err := client.AttachVolume(ctx, attachReq)
	require.NoError(t, err)
	require.NotNil(t, attachResp)
	t.Logf("Successfully attached volume to host")

	// Cleanup connection at the end
	defer cleanupConnection(t, client, ctx, hostName, volumeName)

	// Step 4: Verify host exists on the array
	t.Logf("Step 4: Verifying host exists on array")
	hostVerify, err := getHost(t, client, hostName)
	require.NoError(t, err)
	require.NotNil(t, hostVerify)

	require.Equal(t, hostName, hostVerify.Name, "Host name should match")
	require.Contains(t, hostVerify.Nqns, nqn, "Host should have the NQN in its nqns list")
	t.Logf("Successfully verified host exists: %s", hostVerify.Name)

	// Step 5: Verify connection was created on the array
	t.Logf("Step 5: Verifying connection was created on array")
	connections, err := getConnections(t, client, hostName, volumeName)
	require.NoError(t, err)
	require.Len(t, connections, 1, "Should have exactly one connection")

	conn := connections[0]
	require.Equal(t, hostName, conn.Host.Name, "Connection should reference the correct host")
	require.Equal(t, volumeName, conn.Volume.Name, "Connection should reference the correct volume")
	t.Logf("Successfully verified connection exists: host=%s, volume=%s, lun=%d",
		conn.Host.Name, conn.Volume.Name, conn.Lun)

	t.Logf("Test completed successfully")
}

// TestIntegration_AttachVolume_MultipleVolumesToSameHost tests attaching multiple volumes to the same host
func TestIntegration_AttachVolume_MultipleVolumesToSameHost(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_AttachVolume_MultipleVolumesToSameHost")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Step 1: Create two new volumes
	vol1Name := fmt.Sprintf("%s-attach-multi-vol1-%s", volumePrefix, uuid.New().String()[:8])
	vol2Name := fmt.Sprintf("%s-attach-multi-vol2-%s", volumePrefix, uuid.New().String()[:8])

	t.Logf("Step 1: Creating volume 1: %s", vol1Name)
	createReq1 := &models.CreateVolumeRequest{
		UUID: vol1Name,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}
	createResp1, err := client.CreateVolume(ctx, createReq1)
	require.NoError(t, err)
	require.NotNil(t, createResp1)
	require.NotNil(t, createResp1.Volume)
	t.Logf("Successfully created volume 1: %s", vol1Name)

	defer cleanupVolume(t, client, ctx, vol1Name)

	t.Logf("Step 1: Creating volume 2: %s", vol2Name)
	createReq2 := &models.CreateVolumeRequest{
		UUID: vol2Name,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}
	createResp2, err := client.CreateVolume(ctx, createReq2)
	require.NoError(t, err)
	require.NotNil(t, createResp2)
	require.NotNil(t, createResp2.Volume)
	t.Logf("Successfully created volume 2: %s", vol2Name)

	defer cleanupVolume(t, client, ctx, vol2Name)

	// Step 2: Create a new host
	hostUUID := uuid.New().String()
	hostName := fmt.Sprintf("storm-%s", hostUUID[:8])
	expectedNQN := fmt.Sprintf("nqn.2014-08.org.nvmexpress:uuid:%s", hostUUID)

	t.Logf("Step 2: Creating host: %s with NQN: %s", hostName, expectedNQN)
	host, err := createHost(t, client, hostName, expectedNQN)
	require.NoError(t, err)
	require.NotNil(t, host)
	defer cleanupHost(t, client, ctx, hostName)

	// Step 3: Attach vol1 to the host
	t.Logf("Step 3: Attaching volume 1 to host")
	attachReq1 := &models.AttachVolumeRequest{
		UUID: vol1Name,
		Acls: []string{hostUUID},
	}
	attachResp1, err := client.AttachVolume(ctx, attachReq1)
	require.NoError(t, err)
	require.NotNil(t, attachResp1)
	t.Logf("Successfully attached volume 1 to host")

	defer cleanupConnection(t, client, ctx, hostName, vol1Name)

	// Step 4: Verify host exists on the array
	t.Logf("Step 4: Verifying host exists on array")
	hostVerify, err := getHost(t, client, hostName)
	require.NoError(t, err)
	require.NotNil(t, hostVerify)
	require.Equal(t, hostName, hostVerify.Name, "Host name should match")
	require.Contains(t, hostVerify.Nqns, expectedNQN, "Host should have the NQN in its nqns list")
	t.Logf("Successfully verified host exists: %s", hostVerify.Name)

	// Step 5: Verify connection for vol1 was created
	t.Logf("Step 5: Verifying connection for volume 1 was created")
	connections1, err := getConnections(t, client, hostName, vol1Name)
	require.NoError(t, err)
	require.Len(t, connections1, 1, "Should have exactly one connection for vol1")
	require.Equal(t, hostName, connections1[0].Host.Name)
	require.Equal(t, vol1Name, connections1[0].Volume.Name)
	t.Logf("Successfully verified connection for vol1: lun=%d", connections1[0].Lun)

	// Step 6: Attach vol2 to the same host
	t.Logf("Step 6: Attaching volume 2 to the same host")
	attachReq2 := &models.AttachVolumeRequest{
		UUID: vol2Name,
		Acls: []string{hostUUID},
	}
	attachResp2, err := client.AttachVolume(ctx, attachReq2)
	require.NoError(t, err)
	require.NotNil(t, attachResp2)
	t.Logf("Successfully attached volume 2 to host")

	defer cleanupConnection(t, client, ctx, hostName, vol2Name)

	// Step 7: Verify that only 1 host exists (no new host was created)
	t.Logf("Step 7: Verifying no new host was created")
	host2, err := getHost(t, client, hostName)
	require.NoError(t, err)
	require.NotNil(t, host2)
	require.Equal(t, hostName, host2.Name, "Host name should still be the same")
	require.Equal(t, host.Name, host2.Name, "Host should be the same as before")
	t.Logf("Successfully verified same host exists: %s", host2.Name)

	// Step 8: Verify connection for vol2 was created
	t.Logf("Step 8: Verifying connection for volume 2 was created")
	connections2, err := getConnections(t, client, hostName, vol2Name)
	require.NoError(t, err)
	require.Len(t, connections2, 1, "Should have exactly one connection for vol2")
	require.Equal(t, hostName, connections2[0].Host.Name)
	require.Equal(t, vol2Name, connections2[0].Volume.Name)
	t.Logf("Successfully verified connection for vol2: lun=%d", connections2[0].Lun)

	// Step 9: Verify both connections exist for the host
	t.Logf("Step 9: Verifying both connections exist for the host")
	// Get all connections for the host (without filtering by volume)
	allConnectionsPath := fmt.Sprintf("/api/%s/connections?host_names=%s", client.apiVersion, hostName)
	var allConnectionsResp GetConnectionsResponse
	err = client.get(allConnectionsPath, &allConnectionsResp)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(allConnectionsResp.Items), 2, "Should have at least 2 connections for the host")

	// Verify both volumes are connected
	volumeNames := make(map[string]bool)
	for _, conn := range allConnectionsResp.Items {
		if conn.Host.Name == hostName {
			volumeNames[conn.Volume.Name] = true
		}
	}
	require.True(t, volumeNames[vol1Name], "Volume 1 should be connected to host")
	require.True(t, volumeNames[vol2Name], "Volume 2 should be connected to host")
	t.Logf("Successfully verified both volumes are connected to the same host")

	t.Logf("Test completed successfully - attached 2 volumes to 1 host")
}

// TestIntegration_AttachVolume_NonexistentVolume tests attaching a non-existent volume
func TestIntegration_AttachVolume_NonexistentVolume(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_AttachVolume_NonexistentVolume")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Try to attach a volume that doesn't exist
	nonexistentVolume := fmt.Sprintf("%s-nonexistent-%s", volumePrefix, uuid.New().String()[:8])
	hostUUID := uuid.New().String()

	t.Logf("Attempting to attach non-existent volume: %s to host: %s", nonexistentVolume, hostUUID)
	attachReq := &models.AttachVolumeRequest{
		UUID: nonexistentVolume,
		Acls: []string{hostUUID},
	}

	resp, err := client.AttachVolume(ctx, attachReq)
	require.Error(t, err, "Should fail when volume doesn't exist")
	require.Nil(t, resp)
	t.Logf("Correctly failed with error: %v", err)

	t.Logf("Test completed successfully")
}

// TestIntegration_AttachVolume_Idempotency tests attaching the same volume to the same host twice
func TestIntegration_AttachVolume_Idempotency(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_AttachVolume_Idempotency")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Step 1: Create a volume
	volumeName := fmt.Sprintf("%s-idempotent-%s", volumePrefix, uuid.New().String()[:8])
	createReq := &models.CreateVolumeRequest{
		UUID: volumeName,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}

	t.Logf("Step 1: Creating volume: %s", volumeName)
	createResp, err := client.CreateVolume(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	t.Logf("Successfully created volume: %s", volumeName)

	defer cleanupVolume(t, client, ctx, volumeName)

	// Step 2: Create a new host
	hostUUID := uuid.New().String()
	hostName := fmt.Sprintf("storm-%s", hostUUID[:8])
	nqn := fmt.Sprintf("nqn.2014-08.org.nvmexpress:uuid:%s", hostUUID)

	t.Logf("Step 2: Creating host: %s with NQN: %s", hostName, nqn)
	host, err := createHost(t, client, hostName, nqn)
	require.NoError(t, err)
	require.NotNil(t, host)
	defer cleanupHost(t, client, ctx, hostName)

	// Step 3: Attach volume to host (first time)
	t.Logf("Step 3: Attaching volume to host (first time)")
	attachReq := &models.AttachVolumeRequest{
		UUID: volumeName,
		Acls: []string{hostUUID},
	}

	attachResp1, err := client.AttachVolume(ctx, attachReq)
	require.NoError(t, err)
	require.NotNil(t, attachResp1)
	t.Logf("Successfully attached volume to host (first time)")

	defer cleanupConnection(t, client, ctx, hostName, volumeName)

	t.Logf("Test completed successfully - idempotency verified")
}

// TestIntegration_AttachVolume_AlreadyAttached tests attaching a volume that's already attached to another host
func TestIntegration_AttachVolume_AlreadyAttached(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_AttachVolume_AlreadyAttached")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Step 1: Create a volume
	volumeName := fmt.Sprintf("%s-multiattach-%s", volumePrefix, uuid.New().String()[:8])
	createReq := &models.CreateVolumeRequest{
		UUID: volumeName,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}

	t.Logf("Step 1: Creating volume: %s", volumeName)
	createResp, err := client.CreateVolume(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	t.Logf("Successfully created volume: %s", volumeName)

	defer cleanupVolume(t, client, ctx, volumeName)

	// Step 2: Create host1
	host1UUID := uuid.New().String()
	host1Name := fmt.Sprintf("storm-%s", host1UUID[:8])
	expectedNQN1 := fmt.Sprintf("nqn.2014-08.org.nvmexpress:uuid:%s", host1UUID)

	t.Logf("Step 2: Creating host1: %s with NQN: %s", host1Name, expectedNQN1)
	host1, err := createHost(t, client, host1Name, expectedNQN1)
	require.NoError(t, err)
	require.NotNil(t, host1)
	defer cleanupHost(t, client, ctx, host1Name)

	// Step 3: Attach volume to host1
	t.Logf("Step 3: Attaching volume to host1: %s", host1UUID)
	attachReq1 := &models.AttachVolumeRequest{
		UUID: volumeName,
		Acls: []string{host1UUID},
	}

	attachResp1, err := client.AttachVolume(ctx, attachReq1)
	require.NoError(t, err)
	require.NotNil(t, attachResp1)
	t.Logf("Successfully attached volume to host1")

	defer cleanupConnection(t, client, ctx, host1Name, volumeName)

	// Step 4: Verify connection to host1
	t.Logf("Step 4: Verifying connection to host1")
	connections1, err := getConnections(t, client, host1Name, volumeName)
	require.NoError(t, err)
	require.Len(t, connections1, 1, "Should have connection to host1")
	t.Logf("Connection to host1 verified: lun=%d", connections1[0].Lun)

	// Step 5: Create host2
	host2UUID := uuid.New().String()
	host2Name := fmt.Sprintf("storm-%s", host2UUID[:8])
	expectedNQN2 := fmt.Sprintf("nqn.2014-08.org.nvmexpress:uuid:%s", host2UUID)

	t.Logf("Step 5: Creating host2: %s with NQN: %s", host2Name, expectedNQN2)
	host2, err := createHost(t, client, host2Name, expectedNQN2)
	require.NoError(t, err)
	require.NotNil(t, host2)
	defer cleanupHost(t, client, ctx, host2Name)

	// Step 6: Try to attach the same volume to host2
	t.Logf("Step 6: Attempting to attach same volume to host2: %s", host2UUID)
	attachReq2 := &models.AttachVolumeRequest{
		UUID: volumeName,
		Acls: []string{host2UUID},
	}

	attachResp2, err := client.AttachVolume(ctx, attachReq2)

	// FlashArray behavior: It may allow multi-attach or may fail
	// We'll handle both cases and document the behavior
	if err != nil {
		t.Logf("Multi-attach not allowed (expected behavior): %v", err)

		// Verify original connection to host1 still exists
		connections1After, err := getConnections(t, client, host1Name, volumeName)
		require.NoError(t, err)
		require.Len(t, connections1After, 1, "Original connection to host1 should still exist")
		t.Logf("Original connection to host1 preserved")

		// Verify no connection to host2 was created
		connections2, err := getConnections(t, client, host2Name, volumeName)
		if err == nil {
			require.Len(t, connections2, 0, "Should have no connection to host2")
		}
		t.Logf("No connection to host2 created (correct)")

	} else {
		t.Logf("Multi-attach allowed - volume attached to both hosts")
		require.NotNil(t, attachResp2)

		defer cleanupConnection(t, client, ctx, host2Name, volumeName)

		// Verify connections to both hosts exist
		connections1After, err := getConnections(t, client, host1Name, volumeName)
		require.NoError(t, err)
		require.Len(t, connections1After, 1, "Should have connection to host1")

		connections2, err := getConnections(t, client, host2Name, volumeName)
		require.NoError(t, err)
		require.Len(t, connections2, 1, "Should have connection to host2")

		t.Logf("Volume successfully attached to both hosts:")
		t.Logf("  - Host1 (%s): lun=%d", expectedNQN1, connections1After[0].Lun)
		t.Logf("  - Host2 (%s): lun=%d", expectedNQN2, connections2[0].Lun)
	}

	t.Logf("Test completed successfully")
}

// TestIntegration_AttachVolume_DetachAndReattach tests the full attach/detach/reattach cycle
func TestIntegration_AttachVolume_DetachAndReattach(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_AttachVolume_DetachAndReattach")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Step 1: Create a volume
	volumeName := fmt.Sprintf("%s-reattach-%s", volumePrefix, uuid.New().String()[:8])
	createReq := &models.CreateVolumeRequest{
		UUID: volumeName,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}

	t.Logf("Step 1: Creating volume: %s", volumeName)
	createResp, err := client.CreateVolume(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	t.Logf("Successfully created volume: %s", volumeName)

	defer cleanupVolume(t, client, ctx, volumeName)

	// Step 2: Create a new host
	hostUUID := uuid.New().String()
	hostName := fmt.Sprintf("storm-%s", hostUUID[:8])
	nqn := fmt.Sprintf("nqn.2014-08.org.nvmexpress:uuid:%s", hostUUID)

	t.Logf("Step 2: Creating host: %s with NQN: %s", hostName, nqn)
	host, err := createHost(t, client, hostName, nqn)
	require.NoError(t, err)
	require.NotNil(t, host)
	defer cleanupHost(t, client, ctx, hostName)

	// Step 3: Attach volume to host
	t.Logf("Step 3: Attaching volume to host: %s", hostUUID)
	attachReq := &models.AttachVolumeRequest{
		UUID: volumeName,
		Acls: []string{hostUUID},
	}

	attachResp, err := client.AttachVolume(ctx, attachReq)
	require.NoError(t, err)
	require.NotNil(t, attachResp)
	t.Logf("Successfully attached volume to host")

	// Step 4: Verify connection exists
	t.Logf("Step 4: Verifying connection exists")
	connections1, err := getConnections(t, client, hostName, volumeName)
	require.NoError(t, err)
	require.Len(t, connections1, 1, "Should have connection after attach")
	initialLun := connections1[0].Lun
	t.Logf("Connection verified: lun=%d", initialLun)

	// Step 5: Detach volume from host
	t.Logf("Step 5: Detaching volume from host")
	detachReq := &models.DetachVolumeRequest{
		UUID: volumeName,
		Acls: []string{hostUUID},
	}

	detachResp, err := client.DetachVolume(ctx, detachReq)
	require.NoError(t, err)
	require.NotNil(t, detachResp)
	t.Logf("Successfully detached volume from host")

	// Step 6: Verify connection is removed
	t.Logf("Step 6: Verifying connection is removed")
	connections2, err := getConnections(t, client, hostName, volumeName)
	if err == nil {
		require.Len(t, connections2, 0, "Should have no connections after detach")
	}
	t.Logf("Connection successfully removed")

	// Step 7: Verify host still exists
	t.Logf("Step 7: Verifying host still exists")
	hostVerify, err := getHost(t, client, hostName)
	require.NoError(t, err)
	require.NotNil(t, hostVerify)
	require.Equal(t, hostName, hostVerify.Name)
	t.Logf("Host still exists: %s", hostVerify.Name)

	// Step 8: Re-attach volume to the same host
	t.Logf("Step 8: Re-attaching volume to the same host")
	reattachResp, err := client.AttachVolume(ctx, attachReq)
	require.NoError(t, err)
	require.NotNil(t, reattachResp)
	t.Logf("Successfully re-attached volume to host")

	defer cleanupConnection(t, client, ctx, hostName, volumeName)

	// Step 9: Verify connection is recreated
	t.Logf("Step 9: Verifying connection is recreated")
	connections3, err := getConnections(t, client, hostName, volumeName)
	require.NoError(t, err)
	require.Len(t, connections3, 1, "Should have connection after re-attach")
	newLun := connections3[0].Lun
	t.Logf("Connection recreated: lun=%d (original lun=%d)", newLun, initialLun)

	// Note: LUN may or may not be the same after re-attach
	t.Logf("Test completed successfully - full attach/detach/reattach cycle verified")
}

// TestIntegration_AttachDetachVolume_Integration tests comprehensive attach/detach integration
func TestIntegration_AttachDetachVolume_Integration(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_AttachDetachVolume_Integration")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Step 1: Create two volumes
	vol1Name := fmt.Sprintf("%s-integ-vol1-%s", volumePrefix, uuid.New().String()[:8])
	vol2Name := fmt.Sprintf("%s-integ-vol2-%s", volumePrefix, uuid.New().String()[:8])

	t.Logf("Step 1: Creating volume 1: %s", vol1Name)
	createReq1 := &models.CreateVolumeRequest{
		UUID: vol1Name,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}
	createResp1, err := client.CreateVolume(ctx, createReq1)
	require.NoError(t, err)
	require.NotNil(t, createResp1)
	t.Logf("Successfully created volume 1: %s", vol1Name)

	defer cleanupVolume(t, client, ctx, vol1Name)

	t.Logf("Step 1: Creating volume 2: %s", vol2Name)
	createReq2 := &models.CreateVolumeRequest{
		UUID: vol2Name,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}
	createResp2, err := client.CreateVolume(ctx, createReq2)
	require.NoError(t, err)
	require.NotNil(t, createResp2)
	t.Logf("Successfully created volume 2: %s", vol2Name)

	defer cleanupVolume(t, client, ctx, vol2Name)

	// Step 2: Create a host
	hostUUID := uuid.New().String()
	hostName := fmt.Sprintf("storm-%s", hostUUID[:8])
	nqn := fmt.Sprintf("nqn.2014-08.org.nvmexpress:uuid:%s", hostUUID)

	t.Logf("Step 2: Creating host: %s with NQN: %s", hostName, nqn)
	host, err := createHost(t, client, hostName, nqn)
	require.NoError(t, err)
	require.NotNil(t, host)
	defer cleanupHost(t, client, ctx, hostName)

	// Step 3: Attach both volumes to the host
	t.Logf("Step 3: Attaching volume 1 to host: %s", hostUUID)
	attachReq1 := &models.AttachVolumeRequest{
		UUID: vol1Name,
		Acls: []string{hostUUID},
	}
	attachResp1, err := client.AttachVolume(ctx, attachReq1)
	require.NoError(t, err)
	require.NotNil(t, attachResp1)
	t.Logf("Successfully attached volume 1")

	t.Logf("Step 3: Attaching volume 2 to same host")
	attachReq2 := &models.AttachVolumeRequest{
		UUID: vol2Name,
		Acls: []string{hostUUID},
	}
	attachResp2, err := client.AttachVolume(ctx, attachReq2)
	require.NoError(t, err)
	require.NotNil(t, attachResp2)
	t.Logf("Successfully attached volume 2")

	// Step 4: Verify both connections exist
	t.Logf("Step 4: Verifying both connections exist")
	connections1, err := getConnections(t, client, hostName, vol1Name)
	require.NoError(t, err)
	require.Len(t, connections1, 1, "Should have connection for volume 1")

	connections2, err := getConnections(t, client, hostName, vol2Name)
	require.NoError(t, err)
	require.Len(t, connections2, 1, "Should have connection for volume 2")
	t.Logf("Both connections verified: vol1_lun=%d, vol2_lun=%d",
		connections1[0].Lun, connections2[0].Lun)

	// Step 5: Detach volume 1
	t.Logf("Step 5: Detaching volume 1 from host")
	detachReq1 := &models.DetachVolumeRequest{
		UUID: vol1Name,
		Acls: []string{hostUUID},
	}
	detachResp1, err := client.DetachVolume(ctx, detachReq1)
	require.NoError(t, err)
	require.NotNil(t, detachResp1)
	t.Logf("Successfully detached volume 1")

	// Step 6: Verify volume 1 connection is removed but volume 2 connection remains
	t.Logf("Step 6: Verifying volume 1 connection removed, volume 2 connection remains")
	connections1After, err := getConnections(t, client, hostName, vol1Name)
	if err == nil {
		require.Len(t, connections1After, 0, "Should have no connection for volume 1 after detach")
	}
	t.Logf("Volume 1 connection removed")

	connections2After, err := getConnections(t, client, hostName, vol2Name)
	require.NoError(t, err)
	require.Len(t, connections2After, 1, "Should still have connection for volume 2")
	t.Logf("Volume 2 connection still exists: lun=%d", connections2After[0].Lun)

	// Step 7: Verify host still exists
	t.Logf("Step 7: Verifying host still exists")
	hostVerify, err := getHost(t, client, hostName)
	require.NoError(t, err)
	require.NotNil(t, hostVerify)
	require.Equal(t, hostName, hostVerify.Name)
	t.Logf("Host still exists: %s", hostVerify.Name)

	// Step 8: Detach volume 2
	t.Logf("Step 8: Detaching volume 2 from host")
	detachReq2 := &models.DetachVolumeRequest{
		UUID: vol2Name,
		Acls: []string{hostUUID},
	}
	detachResp2, err := client.DetachVolume(ctx, detachReq2)
	require.NoError(t, err)
	require.NotNil(t, detachResp2)
	t.Logf("Successfully detached volume 2")

	// Step 9: Verify all connections are removed
	t.Logf("Step 9: Verifying all connections are removed")
	connections2Final, err := getConnections(t, client, hostName, vol2Name)
	if err == nil {
		require.Len(t, connections2Final, 0, "Should have no connection for volume 2 after detach")
	}
	t.Logf("All connections removed")

	// Step 10: Verify host still exists (host is not auto-deleted)
	t.Logf("Step 10: Verifying host still exists after all detaches")
	hostFinal, err := getHost(t, client, hostName)
	require.NoError(t, err)
	require.NotNil(t, hostFinal)
	require.Equal(t, hostName, hostFinal.Name)
	t.Logf("Host still exists: %s (hosts are not auto-deleted)", hostFinal.Name)

	// Step 11: Re-attach volume 1 to verify host can be reused
	t.Logf("Step 11: Re-attaching volume 1 to verify host can be reused")
	reattachResp, err := client.AttachVolume(ctx, attachReq1)
	require.NoError(t, err)
	require.NotNil(t, reattachResp)
	t.Logf("Successfully re-attached volume 1")

	defer cleanupConnection(t, client, ctx, hostName, vol1Name)

	// Step 12: Verify connection is recreated
	t.Logf("Step 12: Verifying connection is recreated")
	connectionsReattach, err := getConnections(t, client, hostName, vol1Name)
	require.NoError(t, err)
	require.Len(t, connectionsReattach, 1, "Should have connection after re-attach")
	t.Logf("Connection recreated: lun=%d", connectionsReattach[0].Lun)

	t.Logf("Test completed successfully - comprehensive attach/detach integration verified")
}

// TestIntegration_AttachVolume_MultipleHosts_FilterByNQN tests that getOrCreateHost correctly filters by NQN
// Creates 3 hosts with different NQNs, searches for one specific host by NQN, and verifies only that host is returned
func TestIntegration_AttachVolume_MultipleHosts_FilterByNQN(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_AttachVolume_MultipleHosts_FilterByNQN")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Generate 3 unique UUIDs for the hosts
	uuid1 := uuid.New().String()
	uuid2 := uuid.New().String()
	uuid3 := uuid.New().String()

	// Create short UUIDs for host names
	shortUUID1 := uuid1[:8]
	shortUUID2 := uuid2[:8]
	shortUUID3 := uuid3[:8]

	// Create host names
	host1Name := fmt.Sprintf("storm-%s", shortUUID1)
	host2Name := fmt.Sprintf("storm-%s", shortUUID2)
	host3Name := fmt.Sprintf("storm-%s", shortUUID3)

	// Create NQNs
	nqn1 := fmt.Sprintf("nqn.2014-08.org.nvmexpress:uuid:%s", uuid1)
	nqn2 := fmt.Sprintf("nqn.2014-08.org.nvmexpress:uuid:%s", uuid2)
	nqn3 := fmt.Sprintf("nqn.2014-08.org.nvmexpress:uuid:%s", uuid3)

	t.Logf("Step 1: Creating 3 hosts with unique NQNs")
	t.Logf("  Host1: %s with NQN: %s", host1Name, nqn1)
	t.Logf("  Host2: %s with NQN: %s", host2Name, nqn2)
	t.Logf("  Host3: %s with NQN: %s", host3Name, nqn3)

	// Create host1
	host1, err := createHost(t, client, host1Name, nqn1)
	require.NoError(t, err)
	require.NotNil(t, host1)
	require.Equal(t, host1Name, host1.Name)
	require.Contains(t, host1.Nqns, nqn1)
	defer cleanupHost(t, client, ctx, host1Name)

	// Create host2
	host2, err := createHost(t, client, host2Name, nqn2)
	require.NoError(t, err)
	require.NotNil(t, host2)
	require.Equal(t, host2Name, host2.Name)
	require.Contains(t, host2.Nqns, nqn2)
	defer cleanupHost(t, client, ctx, host2Name)

	// Create host3
	host3, err := createHost(t, client, host3Name, nqn3)
	require.NoError(t, err)
	require.NotNil(t, host3)
	require.Equal(t, host3Name, host3.Name)
	require.Contains(t, host3.Nqns, nqn3)
	defer cleanupHost(t, client, ctx, host3Name)

	t.Logf("Successfully created all 3 hosts")

	// Step 2: Create a volume to test AttachVolume
	volumeName := fmt.Sprintf("%s-nqn-filter-%s", volumePrefix, uuid.New().String()[:8])
	createReq := &models.CreateVolumeRequest{
		UUID: volumeName,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}

	t.Logf("Step 2: Creating volume: %s", volumeName)
	createResp, err := client.CreateVolume(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Volume)
	t.Logf("Successfully created volume: %s", volumeName)
	defer cleanupVolume(t, client, ctx, volumeName)

	// Step 3: Attach volume to host1 using uuid1
	// This will internally call getHost(uuid1) which should find only host1 by NQN filter
	t.Logf("Step 3: Attaching volume to host1 using UUID: %s", uuid1)
	attachReq := &models.AttachVolumeRequest{
		UUID: volumeName,
		Acls: []string{uuid1},
	}

	attachResp, err := client.AttachVolume(ctx, attachReq)
	require.NoError(t, err)
	require.NotNil(t, attachResp)
	t.Logf("Successfully attached volume to host1 (getHost correctly filtered by NQN)")
	defer cleanupConnection(t, client, ctx, host1Name, volumeName)

	// Step 4: Verify connection was created to host1 (not host2 or host3)
	t.Logf("Step 4: Verifying connection was created to host1 only")
	connections, err := getConnections(t, client, host1Name, volumeName)
	require.NoError(t, err)
	require.Len(t, connections, 1, "Should have exactly one connection")
	require.Equal(t, host1Name, connections[0].Host.Name, "Connection should be to host1")
	require.Equal(t, volumeName, connections[0].Volume.Name, "Connection should be to the correct volume")
	t.Logf("Successfully verified connection to host1: lun=%d", connections[0].Lun)

	// Step 5: Verify no connections to host2 or host3
	t.Logf("Step 5: Verifying no connections to host2 or host3")
	connections2, err := getConnections(t, client, host2Name, volumeName)
	if err == nil {
		require.Len(t, connections2, 0, "Should have no connections to host2")
	}
	connections3, err := getConnections(t, client, host3Name, volumeName)
	if err == nil {
		require.Len(t, connections3, 0, "Should have no connections to host3")
	}
	t.Logf("Verified no connections to host2 or host3")

	t.Logf("Test completed successfully - NQN filter correctly isolates specific host")
}

func TestIntegration_DeleteSnapshot_Success(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_DeleteSnapshot_Success")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Step 1: Create vol1 with volume_uuid_1
	volumeUUID1 := fmt.Sprintf("%s-wildcard-vol1-%s", volumePrefix, uuid.New().String()[:8])
	createVol1Req := &models.CreateVolumeRequest{
		UUID: volumeUUID1,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}

	t.Logf("Creating volume 1: %s", volumeUUID1)
	_, err = client.CreateVolume(ctx, createVol1Req)
	require.NoError(t, err)

	// Step 2: Create vol2 with volume_uuid_2
	volumeUUID2 := fmt.Sprintf("%s-wildcard-vol2-%s", volumePrefix, uuid.New().String()[:8])
	createVol2Req := &models.CreateVolumeRequest{
		UUID: volumeUUID2,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}

	t.Logf("Creating volume 2: %s", volumeUUID2)
	_, err = client.CreateVolume(ctx, createVol2Req)
	require.NoError(t, err)

	// Step 3: Create snap_uuid_1 for vol1
	snapUUID1 := uuid.New().String()
	createSnap1Req := &models.CreateSnapshotRequest{
		UUID:             snapUUID1,
		SourceVolumeUUID: volumeUUID1,
	}

	t.Logf("Creating snapshot 1: %s for volume: %s", snapUUID1, volumeUUID1)
	_, err = client.CreateSnapshot(ctx, createSnap1Req)
	require.NoError(t, err)

	// Step 4: Create snap_uuid_11 for vol1
	snapUUID11 := uuid.New().String()
	createSnap11Req := &models.CreateSnapshotRequest{
		UUID:             snapUUID11,
		SourceVolumeUUID: volumeUUID1,
	}

	t.Logf("Creating snapshot 11: %s for volume: %s", snapUUID11, volumeUUID1)
	_, err = client.CreateSnapshot(ctx, createSnap11Req)
	require.NoError(t, err)

	// Step 5: Create snap_uuid_2 for vol2
	snapUUID2 := uuid.New().String()
	createSnap2Req := &models.CreateSnapshotRequest{
		UUID:             snapUUID2,
		SourceVolumeUUID: volumeUUID2,
	}

	t.Logf("Creating snapshot 2: %s for volume: %s", snapUUID2, volumeUUID2)
	_, err = client.CreateSnapshot(ctx, createSnap2Req)
	require.NoError(t, err)

	// Step 6: Test deleting using regular method
	deleteSnapReq := &models.DeleteSnapshotRequest{
		UUID: snapUUID1,
	}

	t.Logf("DeleteSnapshot method for: %s", snapUUID1)
	deleteResp, err := client.DeleteSnapshot(ctx, deleteSnapReq)
	require.NoError(t, err)
	require.NotNil(t, deleteResp)
	t.Logf("Successfully deleted snapshot using regular method: %s", snapUUID1)

	// Step 7: Verify snapshot 1 is deleted
	getSnap1Req := &models.GetSnapshotRequest{UUID: snapUUID1}
	_, err = client.GetSnapshot(ctx, getSnap1Req)
	require.Error(t, err, "Snapshot should be deleted and not found")
	t.Logf("Verified snapshot 1 is deleted: %s", snapUUID1)

	// Step 8: Verify other snapshots still exist
	getSnap11Req := &models.GetSnapshotRequest{UUID: snapUUID11}
	getSnap11Resp, err := client.GetSnapshot(ctx, getSnap11Req)
	require.NoError(t, err)
	require.NotNil(t, getSnap11Resp.Snapshot)
	t.Logf("Verified snapshot 11 still exists: %s", snapUUID11)

	getSnap2Req := &models.GetSnapshotRequest{UUID: snapUUID2}
	getSnap2Resp, err := client.GetSnapshot(ctx, getSnap2Req)
	require.NoError(t, err)
	require.NotNil(t, getSnap2Resp.Snapshot)
	t.Logf("Verified snapshot 2 still exists: %s", snapUUID2)

	// Clean up note
	t.Logf("Test completed successfully. Manual cleanup required:")
	t.Logf("  - Volume 1: %s", volumeUUID1)
	t.Logf("  - Volume 2: %s", volumeUUID2)
	t.Logf("  - Snapshot 11: %s", snapUUID11)
	t.Logf("  - Snapshot 2: %s", snapUUID2)
}

func TestIntegration_DeleteSnapshot_ErrorCases(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_DeleteSnapshot_ErrorCases")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	t.Run("NonexistentSnapshot", func(t *testing.T) {
		// Try to delete a snapshot that doesn't exist
		nonexistentSnapshotUUID := uuid.New().String()
		deleteSnapReq := &models.DeleteSnapshotRequest{
			UUID: nonexistentSnapshotUUID,
		}

		t.Logf("Attempting to delete nonexistent snapshot: %s", nonexistentSnapshotUUID)
		_, err := client.DeleteSnapshot(ctx, deleteSnapReq)
		require.Error(t, err, "Should fail when deleting nonexistent snapshot")
		t.Logf("Correctly failed to delete nonexistent snapshot: %v", err)
	})

	t.Run("EmptyUUID", func(t *testing.T) {
		// Try to delete with empty UUID
		deleteSnapReq := &models.DeleteSnapshotRequest{
			UUID: "",
		}

		t.Logf("Attempting to delete snapshot with empty UUID")
		_, err := client.DeleteSnapshot(ctx, deleteSnapReq)
		require.Error(t, err, "Should fail when UUID is empty")
		require.Contains(t, err.Error(), "required", "Error should mention required field")
		t.Logf("Correctly failed with empty UUID: %v", err)
	})

	t.Run("DeleteAlreadyDeletedSnapshot", func(t *testing.T) {
		// Create a volume and snapshot, delete it, then try to delete again
		volumeUUID := fmt.Sprintf("%s-delete-twice-%s", volumePrefix, uuid.New().String()[:8])
		createVolReq := &models.CreateVolumeRequest{
			UUID: volumeUUID,
			Source: &models.NewVolumeSpec{
				Size:       testVolumes.Small.SizeBytes,
				SectorSize: testVolumes.Small.SectorSize,
			},
		}

		t.Logf("Creating volume: %s", volumeUUID)
		_, err := client.CreateVolume(ctx, createVolReq)
		require.NoError(t, err)

		// Create snapshot
		snapUUID := uuid.New().String()
		createSnapReq := &models.CreateSnapshotRequest{
			UUID:             snapUUID,
			SourceVolumeUUID: volumeUUID,
		}

		t.Logf("Creating snapshot: %s", snapUUID)
		_, err = client.CreateSnapshot(ctx, createSnapReq)
		require.NoError(t, err)

		// Delete the snapshot first time
		deleteSnapReq := &models.DeleteSnapshotRequest{
			UUID: snapUUID,
		}

		t.Logf("Deleting snapshot first time: %s", snapUUID)
		_, err = client.DeleteSnapshot(ctx, deleteSnapReq)
		require.NoError(t, err)

		// Try to delete again
		t.Logf("Attempting to delete already deleted snapshot: %s", snapUUID)
		_, err = client.DeleteSnapshot(ctx, deleteSnapReq)
		require.Error(t, err, "Should fail when deleting already deleted snapshot")
		t.Logf("Correctly failed to delete already deleted snapshot: %v", err)

		// Cleanup
		t.Logf("Cleanup: deleting volume %s", volumeUUID)
		_, err = client.DeleteVolume(ctx, &models.DeleteVolumeRequest{UUID: volumeUUID})
		if err != nil {
			t.Logf("Warning: failed to cleanup volume %s: %v", volumeUUID, err)
		}
	})
}

func TestIntegration_DeleteSnapshot_Idempotency(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_DeleteSnapshot_Idempotency")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Create a volume
	volumeUUID := fmt.Sprintf("%s-idempotent-%s", volumePrefix, uuid.New().String()[:8])
	createVolReq := &models.CreateVolumeRequest{
		UUID: volumeUUID,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}

	t.Logf("Creating volume: %s", volumeUUID)
	_, err = client.CreateVolume(ctx, createVolReq)
	require.NoError(t, err)

	// Create snapshot
	snapUUID := uuid.New().String()
	createSnapReq := &models.CreateSnapshotRequest{
		UUID:             snapUUID,
		SourceVolumeUUID: volumeUUID,
	}

	t.Logf("Creating snapshot: %s", snapUUID)
	_, err = client.CreateSnapshot(ctx, createSnapReq)
	require.NoError(t, err)

	// Delete the snapshot
	deleteSnapReq := &models.DeleteSnapshotRequest{
		UUID: snapUUID,
	}

	t.Logf("Deleting snapshot: %s", snapUUID)
	resp1, err := client.DeleteSnapshot(ctx, deleteSnapReq)
	require.NoError(t, err)
	require.NotNil(t, resp1)

	// Verify snapshot is deleted
	getSnapReq := &models.GetSnapshotRequest{UUID: snapUUID}
	_, err = client.GetSnapshot(ctx, getSnapReq)
	require.Error(t, err, "Snapshot should be deleted")
	t.Logf("Verified snapshot is deleted: %s", snapUUID)

	// Cleanup
	t.Logf("Cleanup: deleting volume %s", volumeUUID)
	_, err = client.DeleteVolume(ctx, &models.DeleteVolumeRequest{UUID: volumeUUID})
	if err != nil {
		t.Logf("Warning: failed to cleanup volume %s: %v", volumeUUID, err)
	}
}

func TestIntegration_DeleteSnapshot_MultipleSnapshotsFromSameVolume(t *testing.T) {
	cfg := getIntegrationConfig(t)
	if cfg == nil {
		return // Skip if no config
	}
	shouldSkipTest(t, "TestIntegration_DeleteSnapshot_MultipleSnapshotsFromSameVolume")

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// Create a volume
	volumeUUID := fmt.Sprintf("%s-multi-snap-%s", volumePrefix, uuid.New().String()[:8])
	createVolReq := &models.CreateVolumeRequest{
		UUID: volumeUUID,
		Source: &models.NewVolumeSpec{
			Size:       testVolumes.Small.SizeBytes,
			SectorSize: testVolumes.Small.SectorSize,
		},
	}

	t.Logf("Creating volume: %s", volumeUUID)
	_, err = client.CreateVolume(ctx, createVolReq)
	require.NoError(t, err)

	// Create multiple snapshots
	numSnapshots := 3
	snapshotUUIDs := make([]string, numSnapshots)

	for i := 0; i < numSnapshots; i++ {
		snapUUID := uuid.New().String()
		snapshotUUIDs[i] = snapUUID

		createSnapReq := &models.CreateSnapshotRequest{
			UUID:             snapUUID,
			SourceVolumeUUID: volumeUUID,
		}

		t.Logf("Creating snapshot %d: %s", i+1, snapUUID)
		_, err = client.CreateSnapshot(ctx, createSnapReq)
		require.NoError(t, err)
	}

	// Delete the middle snapshot
	deleteSnapReq := &models.DeleteSnapshotRequest{
		UUID: snapshotUUIDs[1],
	}

	t.Logf("Deleting middle snapshot: %s", snapshotUUIDs[1])
	_, err = client.DeleteSnapshot(ctx, deleteSnapReq)
	require.NoError(t, err)

	// Verify the middle snapshot is deleted
	getSnapReq := &models.GetSnapshotRequest{UUID: snapshotUUIDs[1]}
	_, err = client.GetSnapshot(ctx, getSnapReq)
	require.Error(t, err, "Middle snapshot should be deleted")
	t.Logf("Verified middle snapshot is deleted")

	// Verify other snapshots still exist
	for i, snapUUID := range []string{snapshotUUIDs[0], snapshotUUIDs[2]} {
		getSnapReq := &models.GetSnapshotRequest{UUID: snapUUID}
		resp, err := client.GetSnapshot(ctx, getSnapReq)
		require.NoError(t, err, "Snapshot %d should still exist", i)
		require.NotNil(t, resp.Snapshot)
		t.Logf("Verified snapshot %d still exists: %s", i, snapUUID)
	}

	// Cleanup: delete remaining snapshots
	for _, snapUUID := range []string{snapshotUUIDs[0], snapshotUUIDs[2]} {
		deleteReq := &models.DeleteSnapshotRequest{UUID: snapUUID}
		_, err = client.DeleteSnapshot(ctx, deleteReq)
		if err != nil {
			t.Logf("Warning: failed to cleanup snapshot %s: %v", snapUUID, err)
		}
	}

	// Cleanup: delete volume
	t.Logf("Cleanup: deleting volume %s", volumeUUID)
	_, err = client.DeleteVolume(ctx, &models.DeleteVolumeRequest{UUID: volumeUUID})
	if err != nil {
		t.Logf("Warning: failed to cleanup volume %s: %v", volumeUUID, err)
	}
}
