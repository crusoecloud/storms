# PureStorage Integration Tests

This directory contains integration tests for the PureStorage FlashArray client. These tests run against a real FlashArray and require proper configuration.

## Prerequisites

1. **FlashArray Access**: You need access to a PureStorage FlashArray with administrative privileges
2. **API Token or Credentials**: Either an API token or username/password for authentication
3. **Network Access**: The test runner must be able to reach the FlashArray over HTTPS

## Configuration

Integration tests use a YAML configuration file instead of environment variables:

### Setup Configuration File

1. **Copy the template**:
   ```bash
   cp testdata/integration_config.template.yaml testdata/integration_config.yaml
   ```

2. **Edit the configuration** with your FlashArray details:
   ```yaml
   # FlashArray connection settings
   endpoints:
     - "your-flasharray.example.com"

   # Authentication - use either API token OR username/password
   auth_token: "your-api-token-here"    # Preferred method
   username: ""                         # Alternative: username
   password: ""                         # Alternative: password

   # API version (optional, defaults to DefaultAPIVersion)
   api_version: "2.45"
   ```

## Running Integration Tests

Integration tests are tagged with `//go:build integration` and must be explicitly enabled:

### Run All Integration Tests
```bash
go test -tags=integration ./client/vendors/purestorage -v -run TestIntegration_
```

### Run Specific Integration Test
```bash
go test -tags=integration ./client/vendors/purestorage -v -run TestIntegration_CreateVolume_NewVolume
```

### Run with Timeout (Recommended)
```bash
go test -tags=integration ./client/vendors/purestorage -v -timeout=5m
```

## Test Configuration

Test parameters are defined in the Go code (`integration_test.go`), not in the YAML config:
To skip specific tests, modify the `skipTests` slice in `integration_test.go`.

## Test Coverage

The integration tests cover:

### CreateVolume API Tests
- **TestIntegration_CreateVolume_NewVolume**: Creates a new 2GB volume with 4096-byte sectors
- **TestIntegration_CreateVolume_DifferentSizes**: Tests various volume sizes (512MB, 2GB, 10GB)
- **TestIntegration_CreateVolume_ErrorCases**: Tests error scenarios (duplicate names, invalid sizes)

## Important Notes

### Volume Cleanup
⚠️ **The integration tests create real volumes on your FlashArray!**

Currently, the tests do not automatically delete created volumes (since `DeleteVolume` is not yet implemented). 
You will need to manually clean up test volumes from the FlashArray management interface.

Test volumes are named with the pattern: `test-vol-{uuid}`, `test-dup-{uuid}`, etc.

### Test Safety
- Tests use unique volume names (with UUID suffixes) to avoid conflicts
- Tests include timeout contexts to prevent hanging
- Tests skip automatically if required environment variables are not set

### FlashArray Requirements
- FlashArray must support the configured API version (default: 2.30)
- Sufficient free space for test volumes
- API access enabled and properly configured

## Example Test Run

```bash
# Make sure you have configured testdata/integration_config.yaml with your FlashArray details

# Run integration tests
go test -tags=integration ./client/vendors/purestorage -v -timeout=5m -run TestIntegration_

# Expected output:
=== RUN   TestIntegration_CreateVolume_NewVolume
    integration_test.go:58: Creating volume: storms-test-vol-a1b2c3d4 (size: 2147483648 bytes, sector: 4096)
    integration_test.go:72: Successfully created volume: &{UUID:storms-test-vol-a1b2c3d4 Size:2147483648 SectorSize:4096 Acls:[] IsAvailable:true SourceSnapshotUUID:}
    integration_test.go:77: Volume storms-test-vol-a1b2c3d4 created successfully. Manual cleanup required from FlashArray.
--- PASS: TestIntegration_CreateVolume_NewVolume (2.34s)
...
```

## Troubleshooting

### Common Issues

1. **"Integration config file not found"**
   - Copy `testdata/integration_config.template.yaml` to `testdata/integration_config.yaml`
   - Configure it with your FlashArray details

2. **"failed to login and obtain session token"**
   - Check your API token or username/password
   - Verify the FlashArray endpoint is reachable
   - Ensure API access is enabled on the FlashArray

3. **"HTTP 401: Unauthorized"**
   - API token may be expired or invalid
   - Username/password may be incorrect

4. **"HTTP 400: Bad Request"**
   - Check the API version compatibility
   - Verify volume parameters (size)

5. **Network timeouts**
   - Check network connectivity to FlashArray

### Debug Mode

Enable debug logging to see detailed HTTP requests:
```bash
export ZEROLOG_LEVEL=debug
go test -tags=integration ./client/vendors/purestorage -v
```
