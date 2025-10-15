package dms

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
	"time"

	"gitlab.com/crusoeenergy/island/storage/storms/client/vendors/lightbits/loadbalancer"
)

var (
	errNoDMSClients     = errors.New("error no dms clients")
	errMissingAuthToken = errors.New("error missing auth token")
	ErrDMSDisabled      = errors.New("dms is disabled")
)

const (
	RequestTimeoutSeconds = 60
	placeholderHost       = "dms-load-balanced"
)

type loadBalancer interface {
	Dial() (*loadbalancer.Conn, error)
}

// dmsTransport is a custom http.RoundTripper that handles endpoint selection
// via a load balancer and sets the appropriate auth token for each request.
type dmsTransport struct {
	lb           loadBalancer
	addrTokenMap map[string]string
}

// RoundTrip executes a single HTTP transaction. It's the core of the custom transport.
func (t *dmsTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Dial via the load balancer to get a connection to a specific endpoint.
	conn, err := t.lb.Dial()
	if err != nil {
		return nil, fmt.Errorf("failed to dial via load balancer: %w", err)
	}
	// The http.Transport that uses this connection will be responsible for closing it.

	addr := conn.RemoteAddr().String()
	token, ok := t.addrTokenMap[addr]
	if !ok {
		conn.Close() // Must close the connection if we are not using it.

		return nil, fmt.Errorf("no auth token found for dms endpoint %s: %w", addr, errMissingAuthToken)
	}

	// Create a new, single-use transport that is forced to use the connection we just dialed.
	transport := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return conn, nil
		},
		TLSClientConfig: &tls.Config{
			//nolint:gosec // Lightbits may use non-standard TLS certificates.
			InsecureSkipVerify: true,
		},
	}

	// Clone the original request so we can modify it without side effects.
	reqClone := req.Clone(req.Context())
	reqClone.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// Execute the request using our single-use transport.
	httpResp, err := transport.RoundTrip(reqClone)
	if err != nil {
		return nil, fmt.Errorf("failed to execute transport round trip: %w", err)
	}

	return httpResp, nil
}

//nolint:tagliatelle // using using snake case for YAML
type Config struct {
	Enabled   bool       `yaml:"enabled"`
	ClusterID string     `yaml:"cluster_id"`
	Endpoints []Endpoint `yaml:"endpoints"`
}

//nolint:tagliatelle // using using snake case for YAML
type Endpoint struct {
	Addr      string `yaml:"addr"`
	AuthToken string `yaml:"auth_token"`
}

type Client struct {
	*http.Client
	doFunc func(ctx context.Context, method string, url string, reqBody interface{}, resBody interface{}) error

	ClusterID string
}

func NewClientWithLoadBalancer(cfg Config) (*Client, error) {
	if !cfg.Enabled {
		return nil, ErrDMSDisabled
	}

	if cfg.Enabled && len(cfg.Endpoints) < 1 {
		return nil, errNoDMSClients
	}

	addrs := make([]net.Addr, len(cfg.Endpoints))
	addrTokenMap := make(map[string]string, len(cfg.Endpoints))

	for i, endpoint := range cfg.Endpoints {
		addr, err := net.ResolveTCPAddr("tcp", endpoint.Addr)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve tcp addr '%s': %w", endpoint.Addr, err)
		}

		addrs[i] = addr
		// Map the resolved address string (e.g., "1.2.3.4:443") to the token.
		addrTokenMap[addr.String()] = endpoint.AuthToken
	}

	lb := loadbalancer.NewLoadBalancer(loadbalancer.AlgorithmRoundRobin, addrs...)

	client := newClient(lb, addrTokenMap, cfg.ClusterID)

	return client, nil
}

func newClient(lb loadBalancer, addrTokenMap map[string]string, clusterID string) *Client {
	// Create an instance of our custom transport.
	transport := &dmsTransport{
		lb:           lb,
		addrTokenMap: addrTokenMap,
	}

	c := &Client{
		Client: &http.Client{
			Transport: transport,
			Timeout:   time.Second * RequestTimeoutSeconds,
		},
		ClusterID: clusterID,
	}
	c.doFunc = c.do

	return c
}

func (c *Client) do(ctx context.Context, method, url string, reqBody, respBody interface{}) error {
	var reqBodyReader io.Reader
	if reqBody != nil {
		reqBodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBodyReader = bytes.NewBuffer(reqBodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Auth header is now set in the custom dmsTransport.
	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read http response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return returnLightbitsError(res, body)
	}

	if respBody != nil {
		err := json.Unmarshal(body, respBody)
		if err != nil {
			return fmt.Errorf("failed to unmarshal http response body: %w", err)
		}
	}

	return nil
}

var (
	ErrServer   = errors.New("server error")
	ErrNotFound = errors.New("not found")
)

type errorBody struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func returnLightbitsError(res *http.Response, body []byte) error {
	var errBody errorBody
	err := json.Unmarshal(body, &errBody)
	if err != nil {
		return fmt.Errorf("failed to unmarshal http error response: %w", err)
	}

	if res.StatusCode == http.StatusNotFound {
		return fmt.Errorf("%s: %w", errBody.Message, ErrNotFound)
	}

	return fmt.Errorf("%v: %w", errBody.Message, ErrServer)
}

func (c *Client) get(ctx context.Context, url string, respBody interface{}) error {
	return c.doFunc(ctx, http.MethodGet, url, nil, respBody)
}

func (c *Client) post(ctx context.Context, url string, reqBody, respBody interface{}) error {
	return c.doFunc(ctx, http.MethodPost, url, reqBody, respBody)
}

// CloneVolume creates a thick clone volume from a source snapshot.
func (c *Client) CloneVolume(ctx context.Context, req *ThickCloneVolumeRequest) (*ThickCloneVolumeResponse, error) {
	url := fmt.Sprintf("https://%s/api/v1/volumes/thickclone", placeholderHost)
	var resp ThickCloneVolumeResponse
	if err := c.post(ctx, url, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to clone volume: %w", err)
	}

	return &resp, nil
}

// CloneSnapshot creates a thick clone snapshot from a source snapshot.
func (c *Client) CloneSnapshot(ctx context.Context, req *ThickCloneSnapshotRequest,
) (*ThickCloneSnapshotResponse, error) {
	url := fmt.Sprintf("https://%s/api/v1/snapshots/thickclone", placeholderHost)
	var resp ThickCloneSnapshotResponse
	if err := c.post(ctx, url, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to clone snapshot: %w", err)
	}

	return &resp, nil
}

// GetWorkflow retrieves the status and details of a specific workflow.
func (c *Client) GetWorkflow(ctx context.Context, workflowID string) (*GetWorkflowResponse, error) {
	url := fmt.Sprintf("https://%s/api/v1/workflows/%s", placeholderHost, workflowID)
	var resp GetWorkflowResponse
	if err := c.get(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	return &resp, nil
}
