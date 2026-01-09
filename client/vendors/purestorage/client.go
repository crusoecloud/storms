//nolint:revive,lll // skeleton code
package purestorage

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"gitlab.com/crusoeenergy/island/storage/storms/client/models"
)

var (
	errNoEndpoints = errors.New("no endpoints provided")
	ErrServer      = errors.New("server error")
	ErrNotFound    = errors.New("not found")
)

const (
	RequestTimeoutSeconds = 60
	DefaultAPIVersion     = "2.20" // Default FlashArray REST API version
)

type Client struct {
	*http.Client
	endpoints  []string
	authToken  string
	username   string
	password   string
	apiVersion string

	// Thread-safe session token management
	sessionMutex sync.RWMutex
	sessionToken string // Session token obtained from login

	doFunc func(method string, url string, reqBody interface{}, resBody interface{}) error
}

func NewClient(cfg *ClientConfig) (*Client, error) {
	if len(cfg.Endpoints) == 0 {
		return nil, errNoEndpoints
	}

	// Create HTTP client with TLS configuration for FlashArray
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// FlashArray typically uses self-signed certificates
				InsecureSkipVerify: true, //nolint:gosec // ok
			},
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		Timeout: time.Second * RequestTimeoutSeconds,
	}

	// Use configured API version or default
	apiVersion := cfg.APIVersion
	if apiVersion == "" {
		apiVersion = DefaultAPIVersion
	}
	c := &Client{
		Client:     httpClient,
		endpoints:  cfg.Endpoints,
		authToken:  cfg.AuthToken,
		username:   cfg.Username,
		password:   cfg.Password,
		apiVersion: apiVersion,
	}
	c.doFunc = c.do

	log.Info().
		Strs("endpoints", cfg.Endpoints).
		Str("auth_token", maskToken(cfg.AuthToken)).
		Str("username", cfg.Username).
		Str("api_version", apiVersion).
		Msg("Created new PureStorage client")

	return c, nil
}

// maskToken masks the auth token for logging purposes.
func maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}

	return token[:4] + "***" + token[len(token)-4:]
}

// login exchanges the API token for a session token.
func (c *Client) login(endpoint string) (string, error) {
	loginURL := fmt.Sprintf("https://%s/api/%s/login", endpoint, c.apiVersion)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*RequestTimeoutSeconds)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to create login request: %w", err)
	}

	// Set headers for login request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("api-token", c.authToken)

	// Add HTTP basic auth if username and password are provided
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	res, err := c.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send login request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body) //nolint:errcheck // no error check.

		return "", fmt.Errorf("login failed with status %d: %s", res.StatusCode, string(body))
	}

	// Extract session token from x-auth-token header
	sessionToken := res.Header.Get("x-auth-token")
	if sessionToken == "" {
		return "", fmt.Errorf("no session token received in x-auth-token header")
	}

	log.Info().
		Str("endpoint", endpoint).
		Str("session_token", maskToken(sessionToken)).
		Msg("Successfully obtained session token")

	return sessionToken, nil
}

// ensureSessionToken ensures we have a valid session token (thread-safe).
func (c *Client) ensureSessionToken(endpoint string) error {
	// First check with read lock (fast path)
	c.sessionMutex.RLock()
	if c.sessionToken != "" {
		c.sessionMutex.RUnlock()

		return nil // Already have a session token
	}
	c.sessionMutex.RUnlock()

	// Need to obtain session token, use write lock
	c.sessionMutex.Lock()
	defer c.sessionMutex.Unlock()

	// Double-check after acquiring write lock (another goroutine might have set it)
	if c.sessionToken != "" {
		return nil
	}

	sessionToken, err := c.login(endpoint)
	if err != nil {
		return fmt.Errorf("failed to login and obtain session token: %w", err)
	}

	c.sessionToken = sessionToken

	return nil
}

// isAuthError checks if the error is an authentication-related error.
func (c *Client) isAuthError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// Check for common authentication error indicators
	return strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "Unauthorized") ||
		strings.Contains(errStr, "Forbidden") ||
		strings.Contains(errStr, "authentication")
}

// isNetworkError checks if the error is a network connectivity issue that warrants failover.
func (c *Client) isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// Check for network-level errors that indicate the endpoint is unreachable
	var netErr *net.OpError
	var dnsErr *net.DNSError
	var urlErr *url.Error
	var syscallErr *syscall.Errno

	// Check for various network error types
	if errors.As(err, &netErr) {
		return true // Network operation error (connection refused, timeout, etc.)
	}
	if errors.As(err, &dnsErr) {
		return true // DNS resolution error
	}
	if errors.As(err, &urlErr) {
		// URL errors can wrap network errors
		return c.isNetworkError(urlErr.Err)
	}
	if errors.As(err, &syscallErr) {
		// System call errors (connection refused, network unreachable, etc.)
		return true
	}

	// Check error message for common network error patterns
	errStr := strings.ToLower(err.Error())
	networkErrorPatterns := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"network is unreachable",
		"no such host",
		"no route to host",
		"timeout",
		"tls handshake timeout",
		"dial tcp",
		"i/o timeout",
		"context deadline exceeded",
	}

	for _, pattern := range networkErrorPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// do performs HTTP requests to FlashArray endpoints with failover support.
//
//nolint:funlen // function is easy to follow
func (c *Client) do(method, urlStr string, reqBody, respBody interface{}) error {
	var lastErr error

	// Try each endpoint until one succeeds
	for _, endpoint := range c.endpoints {
		fullURL := fmt.Sprintf("https://%s%s", endpoint, urlStr)

		// Ensure we have a valid session token for this endpoint
		err := c.ensureSessionToken(endpoint)
		if err != nil {
			// Check if this is a network error that warrants trying the next endpoint
			if c.isNetworkError(err) {
				log.Warn().
					Str("endpoint", endpoint).
					Err(err).
					Msg("Network error during login, trying next endpoint")
				lastErr = err

				continue
			}
			// For non-network errors during login, don't try other endpoints
			return fmt.Errorf("failed to obtain session token: %w", err)
		}

		err = c.doRequest(method, fullURL, reqBody, respBody)
		if err == nil {
			return nil // Success
		}

		// If we get an authentication error, clear the session token and retry once
		if c.isAuthError(err) {
			log.Warn().
				Str("endpoint", endpoint).
				Msg("Authentication error, clearing session token and retrying")

			// Clear session token (thread-safe)
			c.sessionMutex.Lock()
			c.sessionToken = ""
			c.sessionMutex.Unlock()

			// Retry with fresh session token
			if retryErr := c.ensureSessionToken(endpoint); retryErr == nil {
				if retryErr := c.doRequest(method, fullURL, reqBody, respBody); retryErr == nil {
					return nil // Success on retry
				}
			}
		}

		// Check if this is a network error that warrants trying the next endpoint
		if c.isNetworkError(err) {
			log.Warn().
				Str("endpoint", endpoint).
				Str("method", method).
				Str("url", urlStr).
				Err(err).
				Msg("Network error, trying next endpoint")
			lastErr = err

			continue
		}

		// For non-network errors (HTTP 4xx/5xx, API errors), don't try other endpoints
		log.Error().
			Str("endpoint", endpoint).
			Str("method", method).
			Str("url", urlStr).
			Err(err).
			Msg("Request failed with non-network error, not trying other endpoints")

		return fmt.Errorf("request failed: %w", err)
	}

	return fmt.Errorf("all endpoints failed with network errors, last error: %w", lastErr)
}

// doRequest performs a single HTTP request to a specific endpoint.
//
//nolint:cyclop // function is easy to follow.
func (c *Client) doRequest(method, urlStr string, reqBody, respBody interface{}) error {
	var reqBodyReader io.Reader
	if reqBody != nil {
		reqBodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBodyReader = bytes.NewBuffer(reqBodyBytes)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*RequestTimeoutSeconds)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, urlStr, reqBodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for FlashArray API
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Use session token for authentication (obtained from login) - thread-safe read
	c.sessionMutex.RLock()
	sessionToken := c.sessionToken
	c.sessionMutex.RUnlock()

	if sessionToken != "" {
		req.Header.Set("x-auth-token", sessionToken)
	}

	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle HTTP error status codes
	if res.StatusCode >= 400 {
		return c.handleErrorResponse(res, body)
	}

	// Parse successful response
	if respBody != nil && len(body) > 0 {
		if err := json.Unmarshal(body, respBody); err != nil {
			return fmt.Errorf("failed to unmarshal response body: %w", err)
		}
	}

	return nil
}

// handleErrorResponse handles HTTP error responses from FlashArray.
func (c *Client) handleErrorResponse(res *http.Response, body []byte) error {
	// Try to parse FlashArray error response
	var errorResp struct {
		Errors []struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &errorResp); err == nil && len(errorResp.Errors) > 0 {
		errorMsg := errorResp.Errors[0].Message
		if res.StatusCode == http.StatusNotFound {
			return fmt.Errorf("%s: %w", errorMsg, ErrNotFound)
		}

		return fmt.Errorf("%s: %w", errorMsg, ErrServer)
	}

	// Fallback to generic error
	if res.StatusCode == http.StatusNotFound {
		return fmt.Errorf("resource not found: %w", ErrNotFound)
	}

	return fmt.Errorf("HTTP %d: %s: %w", res.StatusCode, string(body), ErrServer)
}

// HTTP helper methods.

func (c *Client) get(urlStr string, respBody interface{}) error {
	return c.doFunc(http.MethodGet, urlStr, nil, respBody)
}

func (c *Client) post(urlStr string, reqBody, respBody interface{}) error {
	return c.doFunc(http.MethodPost, urlStr, reqBody, respBody)
}

//nolint:unused // may be used in the future.
func (c *Client) put(urlStr string, reqBody, respBody interface{}) error {
	return c.doFunc(http.MethodPut, urlStr, reqBody, respBody)
}

func (c *Client) patch(urlStr string, reqBody, respBody interface{}) error {
	return c.doFunc(http.MethodPatch, urlStr, reqBody, respBody)
}

func (c *Client) delete(urlStr string, reqBody, respBody interface{}) error {
	return c.doFunc(http.MethodDelete, urlStr, reqBody, respBody)
}

// getVolumes is a shared helper method for retrieving volumes from FlashArray.
func (c *Client) getVolumes(volumeNames []string) ([]*models.Volume, error) {
	// Build the API path
	path := fmt.Sprintf("/api/%s/volumes", c.apiVersion)

	// Add names query parameter if volume names are provided
	if len(volumeNames) > 0 {
		params := url.Values{}
		for _, name := range volumeNames {
			params.Add("names", name)
		}
		path += "?" + params.Encode()
	}

	// Make the API call
	var response map[string]interface{}
	err := c.get(path, &response)
	if err != nil {
		return nil, err
	}

	// Parse the response using the same pattern as CreateVolume
	volumes, err := c.parseVolumesResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volumes response: %w", err)
	}

	return volumes, nil
}

func (c *Client) GetVolume(ctx context.Context, req *models.GetVolumeRequest) (*models.GetVolumeResponse, error) {
	// Validate input
	if req.UUID == "" {
		return nil, fmt.Errorf("GetVolume cannot be called with empty value")
	}

	// Use the shared helper with the specific volume name
	volumes, err := c.getVolumes([]string{req.UUID})
	if err != nil {
		return nil, fmt.Errorf("failed to get volume %s: %w", req.UUID, err)
	}

	if len(volumes) == 0 {
		return nil, fmt.Errorf("volume %s not found", req.UUID)
	}

	if len(volumes) != 1 {
		return nil, fmt.Errorf("expected exactly 1 volume, got %d", len(volumes))
	}

	log.Info().
		Str("volume_uuid", req.UUID).
		Msg("Successfully retrieved volume")

	return &models.GetVolumeResponse{
		Volume: volumes[0],
	}, nil
}

func (c *Client) GetVolumes(ctx context.Context, req *models.GetVolumesRequest) (*models.GetVolumesResponse, error) {
	// Use the shared helper with empty list to get all volumes
	volumes, err := c.getVolumes([]string{})
	if err != nil {
		return nil, fmt.Errorf("failed to get volumes: %w", err)
	}

	log.Info().
		Int("volume_count", len(volumes)).
		Msg("Successfully retrieved volumes")

	return &models.GetVolumesResponse{
		Volumes: volumes,
	}, nil
}

/**
 * CreateVolume creates a new volume on the FlashArray.
 *
 *  - If the request contains a NewVolumeSpec, create an empty volume with the specified size and sector size.
 *  - If the request contains a SnapshotSource, create a volume from the specified snapshot.
 */
func (c *Client) CreateVolume(ctx context.Context, req *models.CreateVolumeRequest) (*models.CreateVolumeResponse, error) {
	switch source := req.Source.(type) {
	case *models.NewVolumeSpec:
		return c.createNewVolume(ctx, req.UUID, source.Size, source.SectorSize)
	case *models.SnapshotSource:
		return c.createVolumeFromSnapshot(ctx, req.UUID, source.SnapshotUUID)
	default:
		return nil, fmt.Errorf("unsupported volume source type: %T", req.Source)
	}
}

/**
 * createNewVolume creates a new empty volume on the FlashArray.
 * description:
 *  - Make a POST request to /api/{version}/volume/{volume_name} with the specified size.
 *  - Parse the response to create the volume model.
 *  - Set the sector size on the volume model.
 */
func (c *Client) createNewVolume(_ context.Context, volumeName string, sizeBytes uint64, sectorSize uint32,
) (*models.CreateVolumeResponse, error) {
	// FlashArray REST API: POST /api/{version}/volume/{volume_name}
	path := fmt.Sprintf("/api/%s/volumes?names=%s", c.apiVersion, volumeName)

	requestBody := map[string]interface{}{
		"provisioned": sizeBytes,
	}

	var response map[string]interface{}
	err := c.post(path, requestBody, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume %s: %w", volumeName, err)
	}

	// Parse the response to create the volume model
	volume, err := c.parseCreateVolumeResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume response: %w", err)
	}

	volume.SectorSize = sectorSize

	return &models.CreateVolumeResponse{
		Volume: volume,
	}, nil
}

// createVolumeFromSnapshot creates a volume from an existing snapshot.
func (c *Client) createVolumeFromSnapshot(_ context.Context, volumeName, snapshotName string) (*models.CreateVolumeResponse, error) {
	// FlashArray REST API: POST /api/{version}/volume?names={volume_name}
	// In this case snapshotName is the snapshot suffix
	// 1. Search for snapshot by suffix
	// 1.1. if did not found return error
	// 2. Create volume from snapshot

	if c.apiVersion < DefaultAPIVersion {
		return nil, fmt.Errorf("create volume from snapshot not supported in API version %s", c.apiVersion)
	}

	if volumeName == "" || snapshotName == "" {
		return nil, fmt.Errorf("volumeName and snapshotName are required")
	}

	// Step 1: Search for snapshot by suffix
	snapshotPath := fmt.Sprintf("/api/%s/volume-snapshots?filter=suffix='%s'", c.apiVersion, snapshotName)
	var snapshotResponse map[string]interface{}
	err := c.get(snapshotPath, &snapshotResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot with suffix %s: %w", snapshotName, err)
	}

	// Parse snapshot response to get the full snapshot name
	snapshot, err := c.parseGetSnapshotResponse(snapshotResponse)
	if err != nil {
		return nil, fmt.Errorf("snapshot with suffix %s not found: %w", snapshotName, err)
	}

	// Get the full snapshot name (VOLUME_NAME.SUFFIX format)
	fullSnapshotName := snapshot.SourceVolumeUUID + "." + snapshot.UUID

	// Step 2: Create volume from snapshot using the full snapshot name
	path := fmt.Sprintf("/api/%s/volumes?names=%s", c.apiVersion, volumeName)
	requestBody := map[string]interface{}{
		"source": map[string]string{
			"name": fullSnapshotName,
		},
	}

	var response map[string]interface{}
	err = c.post(path, requestBody, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume %s from snapshot %s: %w", volumeName, fullSnapshotName, err)
	}

	volume, err := c.parseCreateVolumeResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse volume response: %w", err)
	}

	return &models.CreateVolumeResponse{
		Volume: volume,
	}, nil
}

func (c *Client) ResizeVolume(ctx context.Context, req *models.ResizeVolumeRequest) (*models.ResizeVolumeResponse, error) {
	// Validate input
	if req.UUID == "" {
		return nil, fmt.Errorf("ResizeVolume cannot be called with empty value")
	}
	if req.Size == 0 {
		return nil, fmt.Errorf("ResizeVolume cannot be called with zero size")
	}

	// FlashArray REST API: PATCH /api/{version}/volumes?names={volume_name}
	path := fmt.Sprintf("/api/%s/volumes?names=%s", c.apiVersion, req.UUID)
	body := map[string]interface{}{
		"provisioned": req.Size,
	}

	var response map[string]interface{}
	err := c.patch(path, body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to resize volume %s: %w", req.UUID, err)
	}

	log.Info().
		Str("volume_uuid", req.UUID).
		Uint64("new_size", req.Size).
		Msg("Successfully resized volume")

	return &models.ResizeVolumeResponse{}, nil
}

func (c *Client) DeleteVolume(ctx context.Context, req *models.DeleteVolumeRequest) (*models.DeleteVolumeResponse, error) {
	// Validate input
	if req.UUID == "" {
		return nil, fmt.Errorf("DeleteVolume cannot be called with empty value")
	}

	// FlashArray requires a two-step process to delete volumes:
	// 1. First, destroy the volume (PATCH with destroyed=true)
	// 2. Then, eradicate the volume (DELETE)

	// Step 1: Destroy the volume
	destroyPath := fmt.Sprintf("/api/%s/volumes?names=%s", c.apiVersion, req.UUID)
	destroyBody := map[string]interface{}{
		"destroyed": true,
	}

	var destroyResponse map[string]interface{}
	err := c.patch(destroyPath, destroyBody, &destroyResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to destroy volume %s: %w", req.UUID, err)
	}

	// Step 2: Eradicate the volume
	eradicatePath := fmt.Sprintf("/api/%s/volumes?names=%s", c.apiVersion, req.UUID)
	var eradicateResponse map[string]interface{}
	err = c.delete(eradicatePath, nil, &eradicateResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to eradicate volume %s: %w", req.UUID, err)
	}

	log.Info().
		Str("volume_uuid", req.UUID).
		Msg("Successfully deleted volume")

	return &models.DeleteVolumeResponse{}, nil
}

//nolint:dupl // Detach and attach methods are intentionally similar.
func (c *Client) AttachVolume(ctx context.Context, req *models.AttachVolumeRequest) (*models.AttachVolumeResponse, error) {
	// FlashArray REST API: Attach volume to host
	// Steps:
	// 1. Validate request parameters
	// 2. Get or create host with the UUID (translated to NQN internally)
	// 3. Create connection between host and volume

	if len(req.ACL) != 1 {
		return nil, fmt.Errorf("exactly one ACL value (host UUID) is required, got %d", len(req.ACL))
	}

	if req.UUID == "" {
		return nil, fmt.Errorf("volume UUID is required")
	}

	// ACL contains the host UUID, which will be translated to NQN format in getOrCreateHost
	// Example: "077f1a5f-3240-45c8-a996-4ee013c3f418" -> "nqn.2014-08.org.nvmexpress:uuid:077f1a5f-3240-45c8-a996-4ee013c3f418"
	hostUUID := req.ACL[0]
	volumeName := req.UUID

	// Get or create host (UUID is translated to NQN format internally)
	host, err := c.getOrCreateHost(hostUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create host: %w", err)
	}

	log.Info().
		Str("host_name", host.Name).
		Str("volume_name", volumeName).
		Msg("Attaching volume to host")

	// Create connection between host and volume (using NQN as host name)
	err = c.createConnection(host.Name, volumeName)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	log.Info().
		Str("host_name", host.Name).
		Str("volume_name", volumeName).
		Msg("Successfully attached volume to host")

	return &models.AttachVolumeResponse{}, nil
}

// getOrCreateHost gets an existing host or creates a new one with the given UUID
// The UUID is translated to NQN format: "nqn.2014-08.org.nvmexpress:uuid:<UUID>"
// The NQN is used as both the host name and the NQN value (1:1 mapping).
func (c *Client) getOrCreateHost(uuid string) (*Host, error) {
	// Translate UUID to NQN format

	nqn := fmt.Sprintf("nqn.2014-08.org.nvmexpress:uuid:%s", uuid)

	// try to get this host
	host, err := c.getHost(uuid)

	if err == nil {
		// Host already exists
		log.Info().
			Str("nqn", nqn).
			Msg("Host already exists")

		return host, nil
	}

	// Host doesn't exist, create it
	log.Info().
		Str("nqn", nqn).
		Msg("Creating new host")

	// FlashArray REST API: POST /api/2.20/hosts?names={uuid}
	// Request body: {"nqns": ["nqn..."]}
	createPath := fmt.Sprintf("/api/%s/hosts?names=%s", c.apiVersion, uuid)
	requestBody := map[string]interface{}{
		"nqns": []string{nqn},
	}

	var createResp CreateHostsResponse
	err = c.post(createPath, requestBody, &createResp)

	if err != nil {
		return nil, fmt.Errorf("failed to create host: %w", err)
	}

	if len(createResp.Items) == 0 {
		return nil, fmt.Errorf("no host returned in create response")
	}

	log.Info().
		Str("host_name", createResp.Items[0].Name).
		Msg("Successfully created host")

	return &createResp.Items[0], nil
}

//nolint:dupl // Detach and attach methods are intentionally similar.
func (c *Client) DetachVolume(ctx context.Context, req *models.DetachVolumeRequest) (*models.DetachVolumeResponse, error) {
	// FlashArray REST API: Detach volume from host
	// Steps:
	// 1. Validate request parameters
	// 2. Get host by UUID (translated to NQN internally)
	// 3. Delete connection between host and volume

	if len(req.ACL) != 1 {
		return nil, fmt.Errorf("exactly one ACL (host UUID) is required, got %d", len(req.ACL))
	}

	if req.UUID == "" {
		return nil, fmt.Errorf("volume UUID is required")
	}

	// ACL contains the host UUID, which will be translated to NQN format in getHost
	// Example: "077f1a5f-3240-45c8-a996-4ee013c3f418" -> "nqn.2014-08.org.nvmexpress:uuid:077f1a5f-3240-45c8-a996-4ee013c3f418"
	hostUUID := req.ACL[0]
	volumeName := req.UUID

	// Get host (UUID is translated to NQN format internally)
	host, err := c.getHost(hostUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	log.Info().
		Str("host_name", host.Name).
		Str("volume_name", volumeName).
		Msg("Detaching volume from host")

	// Delete connection between host and volume (using host name)
	err = c.deleteConnection(host.Name, volumeName)
	if err != nil {
		return nil, fmt.Errorf("failed to delete connection: %w", err)
	}

	log.Info().
		Str("host_name", host.Name).
		Str("volume_name", volumeName).
		Msg("Successfully detached volume from host")

	return &models.DetachVolumeResponse{}, nil
}

// getHost gets an existing host by NQN identifier
// The uuid parameter is the raw UUID
// Searches for a host where the nqns field contains "nqn.2014-08.org.nvmexpress:uuid:<uuid>"
// Returns an error if the host doesn't exist.
func (c *Client) getHost(uuid string) (*Host, error) {
	// Build the full NQN identifier
	nqn := fmt.Sprintf("\"nqn.2014-08.org.nvmexpress:uuid:%s\"", uuid)

	q := url.Values{}
	q.Set("is_local", "true")
	q.Set("filter", fmt.Sprintf("nqns=%s", nqn))

	// Search for host by NQN using filter
	// Filter: nqns contains the specific NQN
	path := fmt.Sprintf("/api/%s/hosts?%s", c.apiVersion, q.Encode())

	var getResp GetHostsResponse
	err := c.get(path, &getResp)

	if err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	if len(getResp.Items) == 0 {
		return nil, fmt.Errorf("host with NQN %s not found", nqn)
	}

	// Host found
	log.Info().
		Str("nqn", nqn).
		Str("host_name", getResp.Items[0].Name).
		Msg("Found existing host")

	return &getResp.Items[0], nil
}

// createConnection creates a connection between a host and a volume.
func (c *Client) createConnection(hostName, volumeName string) error {
	path := fmt.Sprintf("/api/%s/connections?host_names=%s&volume_names=%s", c.apiVersion, hostName, volumeName)

	var resp CreateConnectionsResponse
	err := c.post(path, nil, &resp)
	if err != nil {
		return fmt.Errorf("failed to create connection: %w", err)
	}

	log.Info().
		Str("host_name", hostName).
		Str("volume_name", volumeName).
		Msg("Successfully created connection")

	return nil
}

// deleteConnection deletes a connection between a host and a volume.
func (c *Client) deleteConnection(hostName, volumeName string) error {
	path := fmt.Sprintf("/api/%s/connections?host_names=%s&volume_names=%s", c.apiVersion, hostName, volumeName)

	var resp interface{}
	err := c.delete(path, nil, &resp)
	if err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}

	log.Info().
		Str("host_name", hostName).
		Str("volume_name", volumeName).
		Msg("Successfully deleted connection")

	return nil
}

// get snapshot.
func (c *Client) getSnapshotBySuffix(suffix string) (*models.Snapshot, error) {
	if suffix == "" {
		return nil, fmt.Errorf("snapshot suffix is required")
	}

	// FlashArray REST API: GET /api/2.20/volume-snapshots?filter=suffix='UUID' ABC.UUID
	path := fmt.Sprintf("/api/%s/volume-snapshots?destroyed=false&filter=suffix='%s'", c.apiVersion, suffix)

	var response map[string]interface{}
	err := c.get(path, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot %s: %w", suffix, err)
	}

	// Parse the response to create the snapshot model
	snapshot, err := c.parseGetSnapshotResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse snapshot response: %w", err)
	}

	return snapshot, nil
}

func (c *Client) GetSnapshot(ctx context.Context, req *models.GetSnapshotRequest) (*models.GetSnapshotResponse, error) {
	if c.apiVersion < DefaultAPIVersion {
		return nil, fmt.Errorf("get snapshot not supported in API version %s", c.apiVersion)
	}

	snapshot, err := c.getSnapshotBySuffix(req.UUID)
	if err != nil {
		return nil, err
	}

	log.Info().
		Str("snapshot_uuid", req.UUID).
		Msg("Successfully retrieved snapshot")

	return &models.GetSnapshotResponse{
		Snapshot: snapshot,
	}, nil
}

func (c *Client) GetSnapshots(ctx context.Context, req *models.GetSnapshotsRequest) (*models.GetSnapshotsResponse, error) {
	// FlashArray REST API: GET /api/2.20/volume-snapshots
	if c.apiVersion < DefaultAPIVersion {
		return nil, fmt.Errorf("get snapshot not supported in API version %s", c.apiVersion)
	}

	filter := "name='*.*' and not(name='*.*.*') and contains(suffix,'-')"

	q := url.Values{}
	q.Set("destroyed", "false")
	q.Set("filter", filter)

	path := fmt.Sprintf("/api/%s/volume-snapshots?%s", c.apiVersion, q.Encode())

	var response map[string]interface{}
	err := c.get(path, &response)

	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots: %w", err)
	}

	// Parse the response to create the snapshot model
	snapshots, err := c.parseGetSnapshotsResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse snapshot response: %w", err)
	}

	log.Info().
		Int("snapshot_count", len(snapshots)).
		Msg("Successfully retrieved snapshots")

	return &models.GetSnapshotsResponse{
		Snapshots: snapshots,
	}, nil
}

/**
 * CreateSnapshot creates a new snapshot on the FlashArray.
 *
 *  - Make a POST request to /api/{version}/volume-snapshots.
 *  - The snapshot name is specified in the query parameter.
 *  - Parse the response to create the snapshot model.
 */
/**
 * CreateSnapshot creates a new snapshot on the FlashArray.
 *
 *  - Make a POST request to /api/{version}/volume-snapshots.
 *  - The snapshot name is specified in the query parameter.
 *  - Parse the response to create the snapshot model.
 */
func (c *Client) CreateSnapshot(ctx context.Context, req *models.CreateSnapshotRequest,
) (*models.CreateSnapshotResponse, error) {
	// FlashArray REST API: POST /api/{version}/volume-snapshots?names={snapshot-name}
	// Creates a snapshot of a volume by specifying source.name
	if c.apiVersion < DefaultAPIVersion {
		return nil, fmt.Errorf("create snapshot not supported in API version %s", c.apiVersion)
	}

	if req.SourceVolumeUUID == "" || req.UUID == "" {
		return nil, fmt.Errorf("sourceVolumeUUID and UUID are required")
	}

	path := fmt.Sprintf("/api/%s/volume-snapshots?source_names=%s", c.apiVersion, req.SourceVolumeUUID)

	requestBody := map[string]interface{}{
		"suffix": req.UUID,
	}

	log.Info().Msgf("Creating snapshot [name=%s] from volume [id=%s] request body: %+v", req.UUID, req.SourceVolumeUUID, requestBody)

	var response map[string]interface{}
	err := c.post(path, requestBody, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot from volume %s: %w", req.SourceVolumeUUID, err)
	}

	// Parse the response to create the snapshot model
	snapshot, err := c.parseCreateSnapshotResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse snapshot response: %w but the snapshot was created", err)
	}

	return &models.CreateSnapshotResponse{
		Snapshot: snapshot,
	}, nil
}

/*
* DeleteSnapshot deletes a snapshot on the FlashArray.
*  - Get Snapshot Full Name from GetSnapshotBySuffix
*  - Snapshot Full Name is in format: VOLUME_NAME.SUFFIX
*  - Make a DELETE request to /api/{version}/volume-snapshots?names={snapshot-name}.
 */
func (c *Client) DeleteSnapshot(ctx context.Context, req *models.DeleteSnapshotRequest,
) (*models.DeleteSnapshotResponse, error) {
	if c.apiVersion < DefaultAPIVersion {
		return nil, fmt.Errorf("delete snapshot not supported in API version %s", c.apiVersion)
	}

	if req.UUID == "" {
		return nil, fmt.Errorf("delelte snapshotUUID is required")
	}

	snapshot, err := c.getSnapshotBySuffix(req.UUID)
	if err != nil {
		return nil, err
	}
	log.Info().Msgf("Deleting snapshot set destroyed to troe: %+v", snapshot)

	path := fmt.Sprintf("/api/%s/volume-snapshots?names=%s", c.apiVersion, snapshot.SourceVolumeUUID+"."+snapshot.UUID)

	requestBody := map[string]interface{}{
		"destroyed": true,
	}

	var resp interface{}
	err = c.patch(path, requestBody, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to delete snapshot: %w", err)
	}

	log.Info().
		Str("snapshot_uuid", req.UUID).
		Msg("Successfully deleted snapshot")

	return &models.DeleteSnapshotResponse{}, nil
}

// parseVolumeResponse parses FlashArray volume response into models.Volume.
func (c *Client) parseCreateVolumeResponse(response map[string]interface{}) (*models.Volume, error) {
	// marshal response to sjon bytes
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal volume response: %w", err)
	}
	// unmarshal json bytes to createVolumeResponse
	var resp CreateVolumesResponse
	err = json.Unmarshal(jsonBytes, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal volume response: %w", err)
	}

	if len(resp.Items) != 1 {
		return nil, fmt.Errorf("expected exactly 1 volume in response, got %d", len(resp.Items))
	}

	pureVol := resp.Items[0]
	// log purevol
	log.Info().Msgf("purevol: %+v", pureVol)

	sourceSnapshotUUID := ""
	if pureVol.Source != nil {
		sourceSnapshotUUID = pureVol.Source.Name
	}

	return &models.Volume{
		UUID:               pureVol.Name, // FlashArray uses volume name as identifier
		Size:               pureVol.Provisioned,
		SectorSize:         0,
		ACL:                []string{},
		IsAvailable:        true, // Newly created volumes are available
		SourceSnapshotUUID: sourceSnapshotUUID,
	}, nil
}

// parseVolumesResponse parses FlashArray volumes response into []*models.Volume.
func (c *Client) parseVolumesResponse(response map[string]interface{}) ([]*models.Volume, error) {
	// marshal response to json bytes
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal volumes response: %w", err)
	}

	// unmarshal json bytes to CreateVolumesResponse (same structure for GET)
	var resp CreateVolumesResponse
	err = json.Unmarshal(jsonBytes, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal volumes response: %w", err)
	}

	// Convert FlashArray volumes to models.Volume
	volumes := make([]*models.Volume, len(resp.Items))
	for i, pureVol := range resp.Items {
		sourceSnapshotUUID := ""
		if pureVol.Source != nil {
			sourceSnapshotUUID = pureVol.Source.Name
		}

		var createdAt time.Time
		if pureVol.Created != 0 {
			createdAt = time.Unix(0, int64(pureVol.Created)*int64(time.Millisecond))
		}

		volumes[i] = &models.Volume{
			UUID:               pureVol.Name, // FlashArray uses volume name as identifier
			Size:               pureVol.Provisioned,
			SectorSize:         0,
			ACL:                []string{},
			IsAvailable:        true, // Assume volumes are available
			SourceSnapshotUUID: sourceSnapshotUUID,
			CreatedAt:          createdAt,
		}
	}

	return volumes, nil
}

// parseCreateSnapshotResponse parses FlashArray snapshot response into models.Snapshot.
func (c *Client) parseCreateSnapshotResponse(response map[string]interface{}) (*models.Snapshot, error) {
	// marshal response to json bytes
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal snapshot response: %w", err)
	}
	// unmarshal json bytes to createSnapshotResponse
	var resp CreateSnapshotsResponse
	err = json.Unmarshal(jsonBytes, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot response: %w", err)
	}

	if len(resp.Items) != 1 {
		return nil, fmt.Errorf("expected exactly 1 snapshot in response, got %d", len(resp.Items))
	}

	pureSnap := resp.Items[0]
	// log snapshot
	log.Info().Msgf("pure snapshot: %+v", pureSnap)

	var createdAt time.Time
	if pureSnap.Created != 0 {
		createdAt = time.Unix(0, int64(pureSnap.Created)*int64(time.Millisecond))
	}

	return &models.Snapshot{
		UUID:             pureSnap.Name, // FlashArray uses snapshot name as identifier
		Size:             pureSnap.Provisioned,
		SectorSize:       0, // Inherited from source volume, not returned in snapshot response
		IsAvailable:      true,
		SourceVolumeUUID: pureSnap.Source.Name,
		CreatedAt:        createdAt,
	}, nil
}

/*
*
  - parseGetSnapshotResponse parses FlashArray snapshot response into models.Snapshot
  - FlashArray REST API: GET /api/2.20/volume-snapshots?filter=suffix='c0a7d370-c125-439c-9795-f83aec41aca5'
  - - The response is an array of snapshots, but we only expect one.
  - - The snapshot name is specified in the query parameter.
  - - The source volume name is returned in the response.
*/
func (c *Client) parseGetSnapshotResponse(response map[string]interface{}) (*models.Snapshot, error) {
	// marshal response to json bytes
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal snapshot response: %w", err)
	}
	// unmarshal json bytes to CreateSnapshotsResponse (same structure for GET)
	var resp CreateSnapshotsResponse
	err = json.Unmarshal(jsonBytes, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot response: %w", err)
	}

	if len(resp.Items) != 1 {
		return nil, fmt.Errorf("expected exactly 1 snapshot in response, got %d", len(resp.Items))
	}

	pureSnap := resp.Items[0]
	// log snapshot
	log.Info().Msgf("pure snapshot: %+v", pureSnap)

	// FlashArray uses snapshot name format VOLUME_NAME.SUFFIX where SUFFIX is the UUID
	// Extract the suffix after the last dot
	snapshotUUID := pureSnap.Suffix

	sourceVolumeUUID := pureSnap.Name
	if idx := strings.Index(pureSnap.Name, "."); idx != -1 {
		sourceVolumeUUID = pureSnap.Name[0:idx]
	}

	if sourceVolumeUUID == "" || snapshotUUID == "" {
		return nil, fmt.Errorf("failed to parse snapshot response: source volume UUID is empty")
	}

	var createdAt time.Time
	if pureSnap.Created != 0 {
		createdAt = time.Unix(0, int64(pureSnap.Created)*int64(time.Millisecond))
	}

	return &models.Snapshot{
		UUID:             snapshotUUID,
		Size:             pureSnap.Provisioned,
		SectorSize:       0, // Inherited from source volume, not returned in snapshot response
		IsAvailable:      true,
		SourceVolumeUUID: sourceVolumeUUID,
		CreatedAt:        createdAt,
	}, nil
}

func (c *Client) parseGetSnapshotsResponse(response map[string]interface{}) ([]*models.Snapshot, error) {
	// marshal response to json bytes
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal snapshot response: %w", err)
	}
	// unmarshal json bytes to CreateSnapshotsResponse (same structure for GET)
	var resp CreateSnapshotsResponse
	err = json.Unmarshal(jsonBytes, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot response: %w", err)
	}

	if resp.Items == nil {
		return nil, fmt.Errorf("failed to unmarshal snapshot response: items is nil")
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no snapshots found")
	}

	snapshots := make([]*models.Snapshot, len(resp.Items))
	for i, pureSnap := range resp.Items {
		sourceVolumeUUID := pureSnap.Name
		if idx := strings.Index(pureSnap.Name, "."); idx != -1 {
			sourceVolumeUUID = pureSnap.Name[0:idx]
		}

		var createdAt time.Time
		if pureSnap.Created != 0 {
			createdAt = time.Unix(0, int64(pureSnap.Created)*int64(time.Millisecond))
		}

		snapshots[i] = &models.Snapshot{
			UUID:             pureSnap.Suffix,
			Size:             pureSnap.Provisioned,
			SectorSize:       0, // Inherited from source volume, not returned in snapshot response
			IsAvailable:      true,
			SourceVolumeUUID: sourceVolumeUUID,
			CreatedAt:        createdAt,
		}
	}

	return snapshots, nil
}
