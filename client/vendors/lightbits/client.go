package lightbits

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

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"gitlab.com/crusoeenergy/island/storage/storms/client/vendors/lightbits/loadbalancer"
)

//nolint:tagliatelle // using snake case for YAML
type ClientConfig struct {
	AddrsStrs         []string `yaml:"addr_strs"`
	AuthToken         string   `yaml:"auth_token"`
	ProjectName       string   `yaml:"project_name"`
	ReplicationFactor int      `yaml:"replication_factor"`
}

var errNoLightbitsClients = errors.New("error no lightbits clients")

func (ClientConfig) IsClientConfig() {}

// String constants representing the compression state.
const (
	CompressionEnabled    string = "enabled"
	CompressionDisabled   string = "disabled"
	RequestTimeoutSeconds        = 60
)

var (
	ErrServer   = errors.New("server error")
	ErrNotFound = errors.New("not found")
)

type Client struct {
	*http.Client
	addr              string
	token             string
	doFunc            func(method string, url string, reqBody interface{}, resBody interface{}) error
	projectName       string
	replicationFactor int
}

type errorBody struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type loadBalancer interface {
	Dial() (*loadbalancer.Conn, error)
}

func NewClientV2(cfg ClientConfig) (*Client, error) {
	addrStrs := cfg.AddrsStrs
	if len(addrStrs) < 1 {
		return nil, errNoLightbitsClients
	}

	addrs := make([]net.Addr, len(addrStrs))
	for i, addrStr := range addrStrs {
		addr, err := net.ResolveTCPAddr("tcp", addrStr)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve tcp addr: %w", err)
		}

		addrs[i] = addr
	}

	lb := loadbalancer.NewLoadBalancer(loadbalancer.AlgorithmRoundRobin, addrs...)
	client := NewClient(lb, cfg)

	return client, nil
}

func NewClient(lb loadBalancer, cfg ClientConfig) *Client {
	c := &Client{
		Client: &http.Client{
			Transport: &http.Transport{
				Dial: func(_, _ string) (net.Conn, error) {
					conn, err := lb.Dial()
					if err != nil {
						return nil, fmt.Errorf("failed to dial lightbits http server: %w", err)
					}

					return conn, nil
				},
				TLSClientConfig: &tls.Config{
					//nolint:gosec // Lightbits uses non-standard TLS certificates.
					InsecureSkipVerify: true,
				},
			},
		},
		addr:              "lightbits.crusoecloud.io",
		token:             cfg.AuthToken,
		projectName:       cfg.ProjectName,
		replicationFactor: cfg.ReplicationFactor,
	}
	c.doFunc = c.do

	return c
}

func (c *Client) GetVolume(name string) (*Volume, error) {
	url := fmt.Sprintf("https://%s/api/v2/projects/%s/volumes/?name=%s", c.addr, c.projectName, name)
	var resp Volume
	if err := c.get(url, &resp); err != nil {
		return nil, fmt.Errorf("failed to get volume: %w", err)
	}

	return &resp, nil
}

func (c *Client) GetVolumes() ([]*Volume, error) {
	const pageSize = 1000 // The number of results to return per request.
	var volumes []*Volume
	url := fmt.Sprintf(
		"https://%s/api/v2/projects/%s/volumes?limit=%d",
		c.addr, c.projectName, pageSize)
	// maximum duration for the api call
	maxWaitTimeSeconds := 10
	maxDuration := time.Duration(maxWaitTimeSeconds) * time.Second

	// Calculate the timeout time
	timeoutTime := time.Now().Add(maxDuration)
	for {
		if time.Now().After(timeoutTime) {
			log.Info().Msgf("GetVolumes in client.go failed with timeout error.")

			break
		}

		// Make a request to the Lightbits API, using the `offset_uuid`
		// parameter to specify the index of the first result to be returned.
		var resp GetVolumeResponse
		if err := c.get(url, &resp); err != nil {
			return volumes, fmt.Errorf("failed to list volumes: %w", err)
		}

		// If there are no more results to be retrieved, then break out of the loop.
		if len(resp.Volumes) < 1 {
			break
		}
		// Add the results from the current request to the `volumes` slice.
		volumes = append(volumes, resp.Volumes...)

		url = fmt.Sprintf(
			"https://%s/api/v2/projects/%s/volumes?offsetUUID=%s&limit=%d",
			c.addr, c.projectName, resp.Volumes[len(resp.Volumes)-1].UUID.String(), pageSize)
	}

	return volumes, nil
}

func (c *Client) CreateVolume(v *Volume) (*Volume, error) {
	url := fmt.Sprintf("https://%s/api/v2/projects/%s/volumes", c.addr, c.projectName)
	var resp Volume
	if err := c.post(url, v, &resp); err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	return &resp, nil
}

func (c *Client) UpdateVolume(id uuid.UUID, req *UpdateVolumeRequest) error {
	url := fmt.Sprintf("https://%s/api/v2/projects/%s/volumes/%s", c.addr, c.projectName, id)
	if err := c.put(url, req, nil); err != nil {
		return fmt.Errorf("failed to update volume: %w", err)
	}

	return nil
}

func (c *Client) DeleteVolume(name string) error {
	url := fmt.Sprintf("https://%s/api/v2/projects/%s/volumes/?name=%s", c.addr, c.projectName, name)
	if err := c.delete(url, nil, nil); err != nil {
		return fmt.Errorf("failed to delete volume: %w", err)
	}

	return nil
}

func (c *Client) GetSnapshots() ([]*Snapshot, error) {
	url := fmt.Sprintf("https://%s/api/v2/projects/%s/snapshots", c.addr, c.projectName)
	var resp GetSnapshotResponse
	if err := c.get(url, &resp); err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	return resp.Snapshots, nil
}

func (c *Client) GetSnapshot(name string) (*Snapshot, error) {
	url := fmt.Sprintf("https://%s/api/v2/projects/%s/snapshots/?Name=%s", c.addr, c.projectName, name)
	var resp Snapshot
	if err := c.get(url, &resp); err != nil {
		return nil, fmt.Errorf("failed to get snapshot %s: %w", name, err)
	}

	return &resp, nil
}

func (c *Client) CreateSnapshot(req *CreateSnapshotRequest) (*Snapshot, error) {
	url := fmt.Sprintf("https://%s/api/v2/projects/%s/snapshots", c.addr, c.projectName)
	var resp Snapshot
	if err := c.post(url, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	return &resp, nil
}

func (c *Client) DeleteSnapshot(name string) error {
	url := fmt.Sprintf("https://%s/api/v2/projects/%s/snapshots/?name=%s", c.addr, c.projectName, name)
	if err := c.delete(url, nil, nil); err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	return nil
}

func (c *Client) do(method, url string, reqBody, respBody interface{}) error {
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
	req, err := http.NewRequestWithContext(ctx, method, url, reqBodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.token))
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

func (c *Client) get(url string, respBody interface{}) error {
	return c.doFunc(http.MethodGet, url, nil, respBody)
}

func (c *Client) post(url string, reqBody, respBody interface{}) error {
	return c.doFunc(http.MethodPost, url, reqBody, respBody)
}

func (c *Client) put(url string, reqBody, respBody interface{}) error {
	return c.doFunc(http.MethodPut, url, reqBody, respBody)
}

func (c *Client) delete(url string, reqBody, respBody interface{}) error {
	return c.doFunc(http.MethodDelete, url, reqBody, respBody)
}
