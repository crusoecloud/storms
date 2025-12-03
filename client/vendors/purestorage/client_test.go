package purestorage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/crusoeenergy/island/storage/storms/client/models"
)

func Test_NewClient(t *testing.T) {
	cfg := &ClientConfig{
		Endpoints:  []string{"10.0.0.1", "10.0.0.2"},
		AuthToken:  "test-token",
		Username:   "testuser",
		Password:   "testpass",
		APIVersion: DefaultAPIVersion,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.Equal(t, cfg.Endpoints, client.endpoints)
	require.Equal(t, cfg.AuthToken, client.authToken)
	require.Equal(t, cfg.Username, client.username)
	require.Equal(t, cfg.Password, client.password)
	require.Equal(t, DefaultAPIVersion, client.apiVersion)
	require.NotNil(t, client.Client)
	require.NotNil(t, client.doFunc)
}

func Test_NewClient_DefaultAPIVersion(t *testing.T) {
	cfg := &ClientConfig{
		Endpoints: []string{"10.0.0.1"},
		AuthToken: "test-token",
		Username:  "testuser",
		Password:  "testpass",
		// APIVersion not set - should use default
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.Equal(t, DefaultAPIVersion, client.apiVersion)
}

func Test_NewClient_NoEndpoints(t *testing.T) {
	cfg := &ClientConfig{
		Endpoints: []string{}, // Empty endpoints
		AuthToken: "test-token",
		Username:  "testuser",
		Password:  "testpass",
	}

	client, err := NewClient(cfg)
	require.Error(t, err)
	require.Nil(t, client)
	require.Contains(t, err.Error(), "no endpoints")
}

func Test_NewClient_HTTPClientConfiguration(t *testing.T) {
	cfg := &ClientConfig{
		Endpoints: []string{"10.0.0.1"},
		AuthToken: "test-token",
		Username:  "testuser",
		Password:  "testpass",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	// Verify HTTP client configuration
	transport, ok := client.Client.Transport.(*http.Transport)
	require.True(t, ok)
	require.NotNil(t, transport.TLSClientConfig)
	require.True(t, transport.TLSClientConfig.InsecureSkipVerify)
	require.Equal(t, time.Second*RequestTimeoutSeconds, client.Client.Timeout)
}

func Test_maskToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "normal token",
			token:    "abcdef123456",
			expected: "abcd***3456",
		},
		{
			name:     "short token",
			token:    "abc",
			expected: "***",
		},
		{
			name:     "empty token",
			token:    "",
			expected: "***",
		},
		{
			name:     "very long token",
			token:    "abcdefghijklmnopqrstuvwxyz123456789",
			expected: "abcd***6789",
		},
		{
			name:     "exactly 8 chars",
			token:    "12345678",
			expected: "***",
		},
		{
			name:     "9 chars",
			token:    "123456789",
			expected: "1234***6789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskToken(tt.token)
			require.Equal(t, tt.expected, result)
		})
	}
}

// Test login functionality with mock HTTP server
func Test_Client_Login_Success(t *testing.T) {
	// Create mock HTTPS server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/"+DefaultAPIVersion+"/login", r.URL.Path)
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "test-api-token", r.Header.Get("api-token"))

		// Return session token in header
		w.Header().Set("x-auth-token", "session-token-123")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Extract host from server URL
	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints:  []string{endpoint},
		AuthToken:  "test-api-token",
		APIVersion: DefaultAPIVersion,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	sessionToken, err := client.login(endpoint)
	require.NoError(t, err)
	require.Equal(t, "session-token-123", sessionToken)
}

func Test_Client_Login_AuthError(t *testing.T) {
	// Create mock server that returns 401
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "Invalid API token"}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "invalid-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	sessionToken, err := client.login(endpoint)
	require.Error(t, err)
	require.Empty(t, sessionToken)
	require.Contains(t, err.Error(), "401")
}

func Test_Client_Login_NoSessionToken(t *testing.T) {
	// Create mock server that doesn't return session token
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// No x-auth-token header
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	sessionToken, err := client.login(endpoint)
	require.Error(t, err)
	require.Empty(t, sessionToken)
	require.Contains(t, err.Error(), "session token")
}

// ========== HTTP Tests (from http_test.go) ==========

func Test_Client_DoRequest_HTTP_Success(t *testing.T) {
	expectedResponse := map[string]interface{}{
		"result": "success",
		"data":   "test-data",
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.Equal(t, "test-session-token", r.Header.Get("x-auth-token"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token" // Set session token directly

	var response map[string]interface{}
	testURL := fmt.Sprintf("https://%s/api/test", endpoint)

	err = client.doRequest("GET", testURL, nil, &response)
	require.NoError(t, err)
	require.Equal(t, expectedResponse, response)
}

func Test_Client_DoRequest_HTTP_WithRequestBody(t *testing.T) {
	requestBody := map[string]interface{}{
		"name": "test-volume",
		"size": 1024,
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify request body
		var receivedBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&receivedBody)
		require.NoError(t, err)
		// JSON unmarshaling converts numbers to float64, so we need to handle that
		require.Equal(t, "test-volume", receivedBody["name"])
		require.Equal(t, float64(1024), receivedBody["size"])

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"result": "created"}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	var response map[string]interface{}
	testURL := fmt.Sprintf("https://%s/api/test", endpoint)

	err = client.doRequest("POST", testURL, requestBody, &response)
	require.NoError(t, err)
	require.Equal(t, "created", response["result"])
}

func Test_Client_DoRequest_HTTP_HTTPError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "Bad request", "code": 400}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	var response map[string]interface{}
	testURL := fmt.Sprintf("https://%s/api/test", endpoint)

	err = client.doRequest("GET", testURL, nil, &response)
	require.Error(t, err)
	require.Contains(t, err.Error(), "400")
	require.Contains(t, err.Error(), "Bad request")
}

func Test_Client_HandleErrorResponse_HTTP(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		responseBody  string
		expectedError string
	}{
		{
			name:          "JSON error response",
			statusCode:    400,
			responseBody:  `{"errors": [{"message": "Invalid parameter", "code": "400"}]}`,
			expectedError: "Invalid parameter",
		},
		{
			name:          "Plain text error response",
			statusCode:    500,
			responseBody:  "Internal Server Error",
			expectedError: "HTTP 500: Internal Server Error",
		},
		{
			name:          "Empty error response",
			statusCode:    404,
			responseBody:  "",
			expectedError: "resource not found",
		},
		{
			name:          "Invalid JSON error response",
			statusCode:    422,
			responseBody:  `{"message": "Invalid JSON"`,
			expectedError: "HTTP 422:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			serverURL, err := url.Parse(server.URL)
			require.NoError(t, err)
			endpoint := serverURL.Host

			cfg := &ClientConfig{
				Endpoints: []string{endpoint},
				AuthToken: "test-token",
			}

			client, err := NewClient(cfg)
			require.NoError(t, err)
			client.sessionToken = "test-session-token"

			var response map[string]interface{}
			testURL := fmt.Sprintf("https://%s/api/test", endpoint)

			err = client.doRequest("GET", testURL, nil, &response)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func Test_Client_HTTPHelperMethods_HTTP(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back the method and any request body
		response := map[string]interface{}{
			"method": r.Method,
			"path":   r.URL.Path,
		}

		if r.Method != "GET" && r.Method != "DELETE" {
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
				response["body"] = body
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	// Use URL path only - the helper methods use the do method internally
	testPath := "/api/test"
	requestBody := map[string]interface{}{"test": "data"}

	// Test GET
	var getResp map[string]interface{}
	err = client.get(testPath, &getResp)
	require.NoError(t, err)
	require.Equal(t, "GET", getResp["method"])

	// Test POST
	var postResp map[string]interface{}
	err = client.post(testPath, requestBody, &postResp)
	require.NoError(t, err)
	require.Equal(t, "POST", postResp["method"])
	// JSON unmarshaling converts to map[string]interface{} with float64 for numbers
	bodyMap, ok := postResp["body"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "data", bodyMap["test"])

	// Test PUT
	var putResp map[string]interface{}
	err = client.put(testPath, requestBody, &putResp)
	require.NoError(t, err)
	require.Equal(t, "PUT", putResp["method"])

	// Test PATCH
	var patchResp map[string]interface{}
	err = client.patch(testPath, requestBody, &patchResp)
	require.NoError(t, err)
	require.Equal(t, "PATCH", patchResp["method"])

	// Test DELETE
	var deleteResp map[string]interface{}
	err = client.delete(testPath, nil, &deleteResp)
	require.NoError(t, err)
	require.Equal(t, "DELETE", deleteResp["method"])
}

// ========== Failover Tests (from failover_test.go) ==========

func Test_Client_Do_Failover_SingleEndpoint_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/"+DefaultAPIVersion+"/login" {
			w.Header().Set("x-auth-token", "session-token")
			w.WriteHeader(http.StatusOK)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints:  []string{endpoint},
		AuthToken:  "test-token",
		APIVersion: DefaultAPIVersion,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	var response map[string]interface{}
	// Use URL path only - the do method adds the endpoint
	err = client.do("GET", "/api/test", nil, &response)
	require.NoError(t, err)
	require.Equal(t, "success", response["result"])
}

func Test_Client_Do_Failover_MultipleEndpoints_FirstFails_NetworkError(t *testing.T) {
	// First endpoint: use invalid host to simulate network error
	endpoint1 := "nonexistent-host:443"

	// Second server succeeds
	server2 := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/"+DefaultAPIVersion+"/login" {
			w.Header().Set("x-auth-token", "session-token-2")
			w.WriteHeader(http.StatusOK)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"result": "success-from-server2"})
	}))
	defer server2.Close()

	serverURL2, err := url.Parse(server2.URL)
	require.NoError(t, err)
	endpoint2 := serverURL2.Host

	cfg := &ClientConfig{
		Endpoints:  []string{endpoint1, endpoint2},
		AuthToken:  "test-token",
		APIVersion: DefaultAPIVersion,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	var response map[string]interface{}
	// Network error on first endpoint should trigger failover to second endpoint
	err = client.do("GET", "/api/test", nil, &response)
	require.NoError(t, err)
	require.Equal(t, "success-from-server2", response["result"])
}

func Test_Client_Do_Failover_MultipleEndpoints_FirstFails_HTTPError(t *testing.T) {
	// First server returns HTTP error (not network error)
	server1 := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/"+DefaultAPIVersion+"/login" {
			w.Header().Set("x-auth-token", "session-token-1")
			w.WriteHeader(http.StatusOK)
			return
		}
		// Return HTTP error - this should NOT trigger failover
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}))
	defer server1.Close()

	// Second server (should not be called)
	server2 := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Second server should not be called for HTTP errors")
		w.WriteHeader(http.StatusOK)
	}))
	defer server2.Close()

	serverURL1, err := url.Parse(server1.URL)
	require.NoError(t, err)
	endpoint1 := serverURL1.Host

	serverURL2, err := url.Parse(server2.URL)
	require.NoError(t, err)
	endpoint2 := serverURL2.Host

	cfg := &ClientConfig{
		Endpoints:  []string{endpoint1, endpoint2},
		AuthToken:  "test-token",
		APIVersion: DefaultAPIVersion,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	var response map[string]interface{}
	// HTTP error should NOT trigger failover
	err = client.do("GET", "/api/test", nil, &response)
	require.Error(t, err)
	require.Contains(t, err.Error(), "500")                     // Should contain the HTTP error
	require.NotContains(t, err.Error(), "all endpoints failed") // Should not try other endpoints
}

func Test_Client_Do_Failover_AllEndpointsFail_NetworkErrors(t *testing.T) {
	// Both endpoints have network connectivity issues
	cfg := &ClientConfig{
		Endpoints: []string{"nonexistent-host1:443", "nonexistent-host2:443"},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	var response map[string]interface{}
	err = client.do("GET", "/api/test", nil, &response)
	require.Error(t, err)
	// Should contain error indicating all endpoints failed with network errors
	require.Contains(t, err.Error(), "all endpoints failed with network errors")
}

func Test_Client_Do_Failover_AuthErrorTriggersRefresh(t *testing.T) {
	loginCallCount := 0
	requestCount := 0

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/"+DefaultAPIVersion+"/login" {
			loginCallCount++
			w.Header().Set("x-auth-token", fmt.Sprintf("session-token-%d", loginCallCount))
			w.WriteHeader(http.StatusOK)
			return
		}

		requestCount++
		if requestCount == 1 {
			// First request fails with auth error
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message": "Session expired"}`))
		} else {
			// Subsequent requests succeed
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"result": "success-after-refresh"})
		}
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints:  []string{endpoint},
		AuthToken:  "test-token",
		APIVersion: DefaultAPIVersion,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	var response map[string]interface{}
	err = client.do("GET", "/api/test", nil, &response)
	require.NoError(t, err)
	require.Equal(t, "success-after-refresh", response["result"])

	// Should have logged in twice: once initially, once for refresh
	require.Equal(t, 2, loginCallCount)
	require.Equal(t, 2, requestCount) // First failed, second succeeded
}

func Test_Client_IsNetworkError_Failover(t *testing.T) {
	client := &Client{}

	testCases := []struct {
		name        string
		err         error
		expectTrue  bool
		description string
	}{
		{
			name:        "nil error",
			err:         nil,
			expectTrue:  false,
			description: "nil error should not be considered network error",
		},
		{
			name:        "connection refused",
			err:         errors.New("dial tcp 127.0.0.1:443: connect: connection refused"),
			expectTrue:  true,
			description: "connection refused should be network error",
		},
		{
			name:        "no such host",
			err:         errors.New("dial tcp: lookup nonexistent-host: no such host"),
			expectTrue:  true,
			description: "DNS resolution failure should be network error",
		},
		{
			name:        "timeout",
			err:         errors.New("dial tcp 10.0.0.1:443: i/o timeout"),
			expectTrue:  true,
			description: "timeout should be network error",
		},
		{
			name:        "context deadline exceeded",
			err:         errors.New("context deadline exceeded"),
			expectTrue:  true,
			description: "context deadline exceeded should be network error",
		},
		{
			name:        "HTTP 500 error",
			err:         errors.New("HTTP 500: Internal Server Error"),
			expectTrue:  false,
			description: "HTTP errors should not be network errors",
		},
		{
			name:        "HTTP 404 error",
			err:         errors.New("HTTP 404: Not Found"),
			expectTrue:  false,
			description: "HTTP errors should not be network errors",
		},
		{
			name:        "authentication error",
			err:         errors.New("HTTP 401: Unauthorized"),
			expectTrue:  false,
			description: "authentication errors should not be network errors",
		},
		{
			name:        "API error",
			err:         errors.New("Volume already exists"),
			expectTrue:  false,
			description: "API errors should not be network errors",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := client.isNetworkError(tc.err)
			require.Equal(t, tc.expectTrue, result, tc.description)
		})
	}
}

// ========== Auth Tests (from auth_test.go) ==========

func Test_isAuthError_Auth(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "401 error",
			err:      errors.New("HTTP 401: Unauthorized"),
			expected: true,
		},
		{
			name:     "403 error",
			err:      errors.New("HTTP 403: Forbidden"),
			expected: true,
		},
		{
			name:     "authentication error",
			err:      errors.New("authentication failed"),
			expected: true,
		},
		{
			name:     "unauthorized error",
			err:      errors.New("Unauthorized access"),
			expected: true,
		},
		{
			name:     "forbidden error",
			err:      errors.New("Forbidden resource"),
			expected: true,
		},
		{
			name:     "network error",
			err:      errors.New("connection refused"),
			expected: false,
		},
		{
			name:     "500 error",
			err:      errors.New("HTTP 500: Internal Server Error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	client := &Client{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.isAuthError(tt.err)
			require.Equal(t, tt.expected, result)
		})
	}
}

func Test_Client_EnsureSessionToken_Auth_Success(t *testing.T) {
	loginCallCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loginCallCount++
		w.Header().Set("x-auth-token", "new-session-token")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	// First call should login
	err = client.ensureSessionToken(endpoint)
	require.NoError(t, err)
	require.Equal(t, 1, loginCallCount)
	require.Equal(t, "new-session-token", client.sessionToken)

	// Second call should not login again (token already exists)
	err = client.ensureSessionToken(endpoint)
	require.NoError(t, err)
	require.Equal(t, 1, loginCallCount) // Should not increment
}

func Test_Client_ConcurrentAuth_Auth(t *testing.T) {
	loginCallCount := 0
	var loginMutex sync.Mutex

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loginMutex.Lock()
		loginCallCount++
		currentCount := loginCallCount
		loginMutex.Unlock()

		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)

		w.Header().Set("x-auth-token", fmt.Sprintf("session-token-%d", currentCount))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	// Launch multiple concurrent authentication attempts
	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := client.ensureSessionToken(endpoint)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		require.NoError(t, err)
	}

	// Should only have logged in once despite concurrent requests
	require.Equal(t, 1, loginCallCount)
	require.NotEmpty(t, client.sessionToken)
}

// ========== CreateVolume Tests ==========

func Test_Client_CreateVolume_NewVolumeSpec_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volumes", DefaultAPIVersion), r.URL.Path)
		require.Equal(t, "test-volume", r.URL.Query().Get("names"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var requestBody map[string]interface{}
		err = json.Unmarshal(body, &requestBody)
		require.NoError(t, err)
		require.Equal(t, float64(1073741824), requestBody["provisioned"])

		// Return successful response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"items": [
				{
					"id": "12345678-1234-1234-1234-123456789012",
					"name": "test-volume",
					"provisioned": 1073741824,
					"created": 1234567890,
					"serial": "TEST123456789",
					"source": null
				}
			]
		}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.CreateVolumeRequest{
		UUID: "test-volume",
		Source: &models.NewVolumeSpec{
			Size:       1073741824, // 1 GB
			SectorSize: 512,
		},
	}

	resp, err := client.CreateVolume(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Volume)
	require.Equal(t, "test-volume", resp.Volume.UUID)
	require.Equal(t, uint64(1073741824), resp.Volume.Size)
	require.Equal(t, uint32(512), resp.Volume.SectorSize)
	require.True(t, resp.Volume.IsAvailable)
	require.Empty(t, resp.Volume.SourceSnapshotUUID)
}

func Test_Client_CreateVolume_UnsupportedSourceType(t *testing.T) {
	cfg := &ClientConfig{
		Endpoints: []string{"10.0.0.1"},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	// Create a request with an unsupported source type (nil)
	req := &models.CreateVolumeRequest{
		UUID:   "test-volume",
		Source: nil,
	}

	resp, err := client.CreateVolume(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "unsupported volume source type")
}

func Test_Client_CreateVolume_NewVolumeSpec_HTTPError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volumes", DefaultAPIVersion), r.URL.Path)

		// Return HTTP error
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errors": [{"message": "Internal server error", "code": "500"}]}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.CreateVolumeRequest{
		UUID: "test-volume",
		Source: &models.NewVolumeSpec{
			Size:       1073741824,
			SectorSize: 512,
		},
	}

	resp, err := client.CreateVolume(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to create volume")
}

func Test_Client_CreateVolume_NewVolumeSpec_ParseError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volumes", DefaultAPIVersion), r.URL.Path)

		// Return invalid JSON
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.CreateVolumeRequest{
		UUID: "test-volume",
		Source: &models.NewVolumeSpec{
			Size:       1073741824,
			SectorSize: 512,
		},
	}

	resp, err := client.CreateVolume(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to create volume")
}

func Test_Client_CreateVolume_SnapshotSource_UnsupportedAPIVersion(t *testing.T) {
	cfg := &ClientConfig{
		Endpoints:  []string{"10.0.0.1"},
		AuthToken:  "test-token",
		APIVersion: "2.19", // Unsupported version (requires 2.20+)
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	req := &models.CreateVolumeRequest{
		UUID: "test-volume",
		Source: &models.SnapshotSource{
			SnapshotUUID: "snapshot-123",
		},
	}

	resp, err := client.CreateVolume(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "not supported in API version")
}

func Test_Client_CreateVolume_SnapshotSource_EmptySnapshotUUID(t *testing.T) {
	cfg := &ClientConfig{
		Endpoints: []string{"10.0.0.1"},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	req := &models.CreateVolumeRequest{
		UUID: "test-volume",
		Source: &models.SnapshotSource{
			SnapshotUUID: "", // Empty snapshot UUID
		},
	}

	resp, err := client.CreateVolume(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "volumeName and snapshotName are required")
}

func Test_Client_CreateVolume_SnapshotSource_EmptyVolumeName(t *testing.T) {
	cfg := &ClientConfig{
		Endpoints: []string{"10.0.0.1"},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	req := &models.CreateVolumeRequest{
		UUID: "", // Empty volume name
		Source: &models.SnapshotSource{
			SnapshotUUID: "snapshot-123",
		},
	}

	resp, err := client.CreateVolume(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "volumeName and snapshotName are required")
}

// Test to change
func Test_Client_CreateVolume_SnapshotSource_HTTPError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First request is GET to search for snapshot by suffix
		if r.Method == "GET" && strings.Contains(r.URL.Path, "volume-snapshots") {
			// Return HTTP error when trying to get the snapshot
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"errors": [{"message": "Snapshot not found", "code": "404"}]}`))
			return
		}

		// Should not reach POST request since GET failed
		t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.CreateVolumeRequest{
		UUID: "test-volume",
		Source: &models.SnapshotSource{
			SnapshotUUID: "nonexistent-snapshot",
		},
	}

	resp, err := client.CreateVolume(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to get snapshot")
}

func Test_Client_CreateVolume_NewVolumeSpec_EmptyItems(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "POST", r.Method)

		// Return empty items array
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items": []}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.CreateVolumeRequest{
		UUID: "test-volume",
		Source: &models.NewVolumeSpec{
			Size:       1073741824,
			SectorSize: 512,
		},
	}

	resp, err := client.CreateVolume(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to parse volume response")
}

func Test_Client_CreateVolume_NewVolumeSpec_VariousSizes(t *testing.T) {
	testCases := []struct {
		name       string
		size       uint64
		sectorSize uint32
	}{
		{
			name:       "1GB_512_sector",
			size:       1073741824,
			sectorSize: 512,
		},
		{
			name:       "10GB_4096_sector",
			size:       10737418240,
			sectorSize: 4096,
		},
		{
			name:       "100GB_512_sector",
			size:       107374182400,
			sectorSize: 512,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "POST", r.Method)

				// Verify request body
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				var requestBody map[string]interface{}
				err = json.Unmarshal(body, &requestBody)
				require.NoError(t, err)
				require.Equal(t, float64(tc.size), requestBody["provisioned"])

				// Return successful response
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf(`{
					"items": [
						{
							"id": "12345678-1234-1234-1234-123456789012",
							"name": "test-volume-%s",
							"provisioned": %d,
							"created": 1234567890,
							"serial": "TEST123456789",
							"source": null
						}
					]
				}`, tc.name, tc.size)))
			}))
			defer server.Close()

			serverURL, err := url.Parse(server.URL)
			require.NoError(t, err)
			endpoint := serverURL.Host

			cfg := &ClientConfig{
				Endpoints: []string{endpoint},
				AuthToken: "test-token",
			}

			client, err := NewClient(cfg)
			require.NoError(t, err)
			client.sessionToken = "test-session-token"

			req := &models.CreateVolumeRequest{
				UUID: fmt.Sprintf("test-volume-%s", tc.name),
				Source: &models.NewVolumeSpec{
					Size:       tc.size,
					SectorSize: tc.sectorSize,
				},
			}

			resp, err := client.CreateVolume(context.Background(), req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Volume)
			require.Equal(t, tc.size, resp.Volume.Size)
			require.Equal(t, tc.sectorSize, resp.Volume.SectorSize)
		})
	}
}

// ========== DeleteVolume Tests ==========

func Test_Client_DeleteVolume_Success(t *testing.T) {
	callCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		require.Equal(t, fmt.Sprintf("/api/%s/volumes", DefaultAPIVersion), r.URL.Path)
		require.Equal(t, "test-volume-uuid", r.URL.Query().Get("names"))

		// Verify headers
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.Equal(t, "test-session-token", r.Header.Get("x-auth-token"))

		if callCount == 1 {
			// First call should be PATCH to destroy the volume
			require.Equal(t, "PATCH", r.Method)

			// Verify request body for destroy operation
			var body map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)
			require.Equal(t, true, body["destroyed"])
		} else if callCount == 2 {
			// Second call should be DELETE to eradicate the volume
			require.Equal(t, "DELETE", r.Method)
		} else {
			t.Fatalf("Unexpected call count: %d", callCount)
		}

		// Return successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items": []}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	// Test DeleteVolume
	req := &models.DeleteVolumeRequest{
		UUID: "test-volume-uuid",
	}

	resp, err := client.DeleteVolume(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 2, callCount) // Should have made exactly 2 calls
}

func Test_Client_DeleteVolume_EradicateError(t *testing.T) {
	callCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		require.Equal(t, fmt.Sprintf("/api/%s/volumes", DefaultAPIVersion), r.URL.Path)
		require.Equal(t, "test-volume-uuid", r.URL.Query().Get("names"))

		if callCount == 1 {
			// First call (destroy) succeeds
			require.Equal(t, "PATCH", r.Method)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"items": []}`))
		} else if callCount == 2 {
			// Second call (eradicate) fails
			require.Equal(t, "DELETE", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"errors": [{"message": "Cannot eradicate volume", "code": "400"}]}`))
		}
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.DeleteVolumeRequest{
		UUID: "test-volume-uuid",
	}

	resp, err := client.DeleteVolume(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to eradicate volume test-volume-uuid")
	require.Contains(t, err.Error(), "Cannot eradicate volume")
	require.Equal(t, 2, callCount) // Should have made both calls
}

func Test_Client_DeleteVolume_NotFound(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "PATCH", r.Method) // First call should be PATCH to destroy
		require.Equal(t, fmt.Sprintf("/api/%s/volumes", DefaultAPIVersion), r.URL.Path)
		require.Equal(t, "nonexistent-volume", r.URL.Query().Get("names"))

		// Return 404 error
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errors": [{"message": "Volume not found", "code": "404"}]}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.DeleteVolumeRequest{
		UUID: "nonexistent-volume",
	}

	resp, err := client.DeleteVolume(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to destroy volume nonexistent-volume")
}

func Test_Client_DeleteVolume_EmptyUUID(t *testing.T) {
	cfg := &ClientConfig{
		Endpoints: []string{"test-endpoint"},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.DeleteVolumeRequest{
		UUID: "",
	}

	resp, err := client.DeleteVolume(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, "DeleteVolume cannot be called with empty value", err.Error())
}

func Test_Client_ResizeVolume_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "PATCH", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volumes", DefaultAPIVersion), r.URL.Path)
		require.Equal(t, "test-volume-uuid", r.URL.Query().Get("names"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var requestBody map[string]interface{}
		err = json.Unmarshal(body, &requestBody)
		require.NoError(t, err)
		require.Equal(t, float64(2147483648), requestBody["provisioned"])

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items": []}`))
	}))
	defer server.Close()

	cfg := &ClientConfig{
		Endpoints: []string{server.URL[8:]}, // Remove https://
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.ResizeVolumeRequest{
		UUID: "test-volume-uuid",
		Size: 2147483648, // 2GB
	}

	resp, err := client.ResizeVolume(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func Test_Client_ResizeVolume_NotFound(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "PATCH", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volumes", DefaultAPIVersion), r.URL.Path)
		require.Equal(t, "nonexistent-volume", r.URL.Query().Get("names"))

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errors": [{"message": "Volume not found", "code": "404"}]}`))
	}))
	defer server.Close()

	cfg := &ClientConfig{
		Endpoints: []string{server.URL[8:]}, // Remove https://
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.ResizeVolumeRequest{
		UUID: "nonexistent-volume",
		Size: 2147483648,
	}

	resp, err := client.ResizeVolume(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to resize volume nonexistent-volume")
}

func Test_Client_ResizeVolume_EmptyUUID(t *testing.T) {
	cfg := &ClientConfig{
		Endpoints: []string{"test-endpoint"},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.ResizeVolumeRequest{
		UUID: "",
		Size: 2147483648,
	}

	resp, err := client.ResizeVolume(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, "ResizeVolume cannot be called with empty value", err.Error())
}

func Test_Client_ResizeVolume_ZeroSize(t *testing.T) {
	cfg := &ClientConfig{
		Endpoints: []string{"test-endpoint"},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.ResizeVolumeRequest{
		UUID: "test-volume-uuid",
		Size: 0,
	}

	resp, err := client.ResizeVolume(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, "ResizeVolume cannot be called with zero size", err.Error())
}

func Test_Client_GetVolume_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volumes", DefaultAPIVersion), r.URL.Path)
		require.Equal(t, "test-volume-uuid", r.URL.Query().Get("names"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"items": [
				{
					"id": "12345678-1234-1234-1234-123456789012",
					"name": "test-volume-uuid",
					"provisioned": 1073741824,
					"created": 1234567890,
					"serial": "TEST123456789",
					"source": null
				}
			]
		}`))
	}))
	defer server.Close()

	cfg := &ClientConfig{
		Endpoints: []string{server.URL[8:]}, // Remove https://
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetVolumeRequest{
		UUID: "test-volume-uuid",
	}

	resp, err := client.GetVolume(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Volume)
	require.Equal(t, "test-volume-uuid", resp.Volume.UUID)
	require.Equal(t, uint64(1073741824), resp.Volume.Size)
	require.Equal(t, "", resp.Volume.SourceSnapshotUUID)
	require.True(t, resp.Volume.IsAvailable)
}

func Test_Client_GetVolume_NotFound(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volumes", DefaultAPIVersion), r.URL.Path)
		require.Equal(t, "nonexistent-volume", r.URL.Query().Get("names"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items": []}`))
	}))
	defer server.Close()

	cfg := &ClientConfig{
		Endpoints: []string{server.URL[8:]}, // Remove https://
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetVolumeRequest{
		UUID: "nonexistent-volume",
	}

	resp, err := client.GetVolume(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "volume nonexistent-volume not found")
}

func Test_Client_GetVolume_EmptyUUID(t *testing.T) {
	cfg := &ClientConfig{
		Endpoints: []string{"test-endpoint"},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetVolumeRequest{
		UUID: "",
	}

	resp, err := client.GetVolume(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, "GetVolume cannot be called with empty value", err.Error())
}

func Test_Client_GetVolume_HTTPError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volumes", DefaultAPIVersion), r.URL.Path)
		require.Equal(t, "test-volume-uuid", r.URL.Query().Get("names"))

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errors": [{"message": "Internal server error", "code": "500"}]}`))
	}))
	defer server.Close()

	cfg := &ClientConfig{
		Endpoints: []string{server.URL[8:]}, // Remove https://
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetVolumeRequest{
		UUID: "test-volume-uuid",
	}

	resp, err := client.GetVolume(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to get volume test-volume-uuid")
}

func Test_Client_GetVolumes_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volumes", DefaultAPIVersion), r.URL.Path)
		require.Empty(t, r.URL.Query().Get("names")) // No names filter for GetVolumes

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"items": [
				{
					"id": "12345678-1234-1234-1234-123456789012",
					"name": "volume-1",
					"provisioned": 1073741824,
					"created": 1234567890,
					"serial": "TEST123456789",
					"source": null
				},
				{
					"id": "87654321-4321-4321-4321-210987654321",
					"name": "volume-2",
					"provisioned": 2147483648,
					"created": 1234567891,
					"serial": "TEST987654321",
					"source": {
						"name": "snapshot-1"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	cfg := &ClientConfig{
		Endpoints: []string{server.URL[8:]}, // Remove https://
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetVolumesRequest{}

	resp, err := client.GetVolumes(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Volumes, 2)

	// Check first volume
	vol1 := resp.Volumes[0]
	require.Equal(t, "volume-1", vol1.UUID)
	require.Equal(t, uint64(1073741824), vol1.Size)
	require.Equal(t, "", vol1.SourceSnapshotUUID)
	require.True(t, vol1.IsAvailable)

	// Check second volume
	vol2 := resp.Volumes[1]
	require.Equal(t, "volume-2", vol2.UUID)
	require.Equal(t, uint64(2147483648), vol2.Size)
	require.Equal(t, "snapshot-1", vol2.SourceSnapshotUUID)
	require.True(t, vol2.IsAvailable)
}

func Test_Client_GetVolumes_EmptyList(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volumes", DefaultAPIVersion), r.URL.Path)
		require.Empty(t, r.URL.Query().Get("names"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items": []}`))
	}))
	defer server.Close()

	cfg := &ClientConfig{
		Endpoints: []string{server.URL[8:]}, // Remove https://
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetVolumesRequest{}

	resp, err := client.GetVolumes(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Volumes, 0)
}

func Test_Client_GetVolumes_HTTPError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volumes", DefaultAPIVersion), r.URL.Path)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errors": [{"message": "Internal server error", "code": "500"}]}`))
	}))
	defer server.Close()

	cfg := &ClientConfig{
		Endpoints: []string{server.URL[8:]}, // Remove https://
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetVolumesRequest{}

	resp, err := client.GetVolumes(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to get volumes")
}

// ========== CreateSnapshot Tests ==========

func Test_Client_CreateSnapshot_APIVersionCheck(t *testing.T) {
	tests := []struct {
		name          string
		apiVersion    string
		expectError   bool
		errorContains string
	}{
		{
			name:          "API version 2.20 - supported",
			apiVersion:    "2.20",
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "API version 2.30 - supported",
			apiVersion:    "2.30",
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "API version 2.19 - not supported",
			apiVersion:    "2.19",
			expectError:   true,
			errorContains: "create snapshot not supported in API version 2.19",
		},
		{
			name:          "API version 2.10 - not supported",
			apiVersion:    "2.10",
			expectError:   true,
			errorContains: "create snapshot not supported in API version 2.10",
		},
		{
			name:          "API version 1.0 - not supported",
			apiVersion:    "1.0",
			expectError:   true,
			errorContains: "create snapshot not supported in API version 1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server for successful snapshot creation
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/"+tt.apiVersion+"/login" {
					w.Header().Set("x-auth-token", "test-session-token")
					w.WriteHeader(http.StatusOK)
					return
				}

				// Verify the correct API endpoint is being used
				expectedPath := "/api/" + tt.apiVersion + "/volume-snapshots"
				if r.URL.Path == expectedPath {
					// Verify query parameter for source volume
					sourceNames := r.URL.Query().Get("source_names")
					require.Equal(t, "test-volume", sourceNames)

					// Verify request body contains suffix
					var reqBody map[string]interface{}
					err := json.NewDecoder(r.Body).Decode(&reqBody)
					require.NoError(t, err)
					require.Equal(t, "test-snapshot", reqBody["suffix"])

					// Mock successful snapshot creation response
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					response := map[string]interface{}{
						"items": []map[string]interface{}{
							{
								"name":        "test-snapshot",
								"provisioned": 1073741824,
								"source": map[string]interface{}{
									"name": "test-volume",
								},
							},
						},
					}
					json.NewEncoder(w).Encode(response)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			serverURL, err := url.Parse(server.URL)
			require.NoError(t, err)
			endpoint := serverURL.Host

			cfg := &ClientConfig{
				Endpoints:  []string{endpoint},
				AuthToken:  "test-token",
				APIVersion: tt.apiVersion,
			}

			client, err := NewClient(cfg)
			require.NoError(t, err)

			// Create snapshot request
			req := &models.CreateSnapshotRequest{
				UUID:             "test-snapshot",
				SourceVolumeUUID: "test-volume",
			}

			// Call CreateSnapshot
			resp, err := client.CreateSnapshot(nil, req)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, resp)
				require.Contains(t, err.Error(), tt.errorContains)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Snapshot)
				require.Equal(t, "test-snapshot", resp.Snapshot.UUID)
			}
		})
	}
}

// ========== GetSnapshot Tests ==========

func Test_Client_GetSnapshot_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volume-snapshots", DefaultAPIVersion), r.URL.Path)
		require.Equal(t, "destroyed=false&filter=suffix='test-snapshot-uuid'", r.URL.RawQuery)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"items": [
				{
					"id": "12345678-1234-1234-1234-123456789012",
					"name": "test-volume.test-snapshot-uuid",
					"provisioned": 1073741824,
					"created": 1234567890,
					"serial": "TEST123456789",
					"suffix": "test-snapshot-uuid",
					"source": {
						"id": "87654321-4321-4321-4321-210987654321",
						"name": "test-volume"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetSnapshotRequest{
		UUID: "test-snapshot-uuid",
	}

	resp, err := client.GetSnapshot(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Snapshot)
	require.Equal(t, "test-snapshot-uuid", resp.Snapshot.UUID)
	require.Equal(t, uint64(1073741824), resp.Snapshot.Size)
	require.Equal(t, "test-volume", resp.Snapshot.SourceVolumeUUID)
	require.True(t, resp.Snapshot.IsAvailable)
}

func Test_Client_GetSnapshot_NotFound(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volume-snapshots", DefaultAPIVersion), r.URL.Path)
		require.Equal(t, "destroyed=false&filter=suffix='nonexistent-snapshot'", r.URL.RawQuery)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items": []}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetSnapshotRequest{
		UUID: "nonexistent-snapshot",
	}

	resp, err := client.GetSnapshot(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "expected exactly 1 snapshot in response, got 0")
}

func Test_Client_GetSnapshot_HTTPError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volume-snapshots", DefaultAPIVersion), r.URL.Path)
		require.Equal(t, "destroyed=false&filter=suffix='test-snapshot-uuid'", r.URL.RawQuery)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errors": [{"message": "Internal server error", "code": "500"}]}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetSnapshotRequest{
		UUID: "test-snapshot-uuid",
	}

	resp, err := client.GetSnapshot(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to get snapshot test-snapshot-uuid")
}

func Test_Client_GetSnapshot_APIVersionCheck(t *testing.T) {
	tests := []struct {
		name          string
		apiVersion    string
		expectError   bool
		errorContains string
	}{
		{
			name:          "API version 2.20 - supported",
			apiVersion:    "2.20",
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "API version 2.30 - supported",
			apiVersion:    "2.30",
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "API version 2.19 - not supported",
			apiVersion:    "2.19",
			expectError:   true,
			errorContains: "get snapshot not supported in API version 2.19",
		},
		{
			name:          "API version 2.10 - not supported",
			apiVersion:    "2.10",
			expectError:   true,
			errorContains: "get snapshot not supported in API version 2.10",
		},
		{
			name:          "API version 1.0 - not supported",
			apiVersion:    "1.0",
			expectError:   true,
			errorContains: "get snapshot not supported in API version 1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server for successful snapshot retrieval
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/"+tt.apiVersion+"/login" {
					w.Header().Set("x-auth-token", "test-session-token")
					w.WriteHeader(http.StatusOK)
					return
				}

				// Verify the correct API endpoint is being used
				expectedPath := "/api/" + tt.apiVersion + "/volume-snapshots"
				if r.URL.Path == expectedPath {
					require.Equal(t, "destroyed=false&filter=suffix='test-snapshot-uuid'", r.URL.RawQuery)

					// Mock successful snapshot retrieval response
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					response := map[string]interface{}{
						"items": []map[string]interface{}{
							{
								"id":          "12345678-1234-1234-1234-123456789012",
								"name":        "test-volume.test-snapshot-uuid",
								"provisioned": 1073741824,
								"created":     1234567890,
								"serial":      "TEST123456789",
								"suffix":      "test-snapshot-uuid",
								"source": map[string]interface{}{
									"id":   "87654321-4321-4321-4321-210987654321",
									"name": "test-volume",
								},
							},
						},
					}
					json.NewEncoder(w).Encode(response)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()

			serverURL, err := url.Parse(server.URL)
			require.NoError(t, err)
			endpoint := serverURL.Host

			cfg := &ClientConfig{
				Endpoints:  []string{endpoint},
				AuthToken:  "test-token",
				APIVersion: tt.apiVersion,
			}

			client, err := NewClient(cfg)
			require.NoError(t, err)

			// Get snapshot request
			req := &models.GetSnapshotRequest{
				UUID: "test-snapshot-uuid",
			}

			// Call GetSnapshot
			resp, err := client.GetSnapshot(context.Background(), req)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, resp)
				require.Contains(t, err.Error(), tt.errorContains)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Snapshot)
				require.Equal(t, "test-snapshot-uuid", resp.Snapshot.UUID)
				require.Equal(t, "test-volume", resp.Snapshot.SourceVolumeUUID)
			}
		})
	}
}

func Test_Client_GetSnapshot_ParseError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volume-snapshots", DefaultAPIVersion), r.URL.Path)

		// Return invalid JSON structure
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items": "invalid"}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetSnapshotRequest{
		UUID: "test-snapshot-uuid",
	}

	resp, err := client.GetSnapshot(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to parse snapshot response")
}

func Test_Client_GetSnapshot_EmptySuffix(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volume-snapshots", DefaultAPIVersion), r.URL.Path)

		// Return multiple snapshots (should only return 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"items": [
				{
					"id": "12345678-1234-1234-1234-123456789012",
					"name": "test-volume",
					"provisioned": 1073741824,
					"created": 1234567890,
					"serial": "TEST123456789",
					"source": {
						"id": "87654321-4321-4321-4321-210987654321",
						"name": "test-volume"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetSnapshotRequest{
		UUID: "test-volume",
	}

	resp, err := client.GetSnapshot(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to parse snapshot response")

}

func Test_Client_GetSnapshot_EmptyName(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volume-snapshots", DefaultAPIVersion), r.URL.Path)

		// Return multiple snapshots (should only return 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"items": [
				{
					"id": "12345678-1234-1234-1234-123456789012",
					"suffix": "test-volume",
					"provisioned": 1073741824,
					"created": 1234567890,
					"serial": "TEST123456789",
					"source": {
						"id": "87654321-4321-4321-4321-210987654321",
						"name": "test-volume"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetSnapshotRequest{
		UUID: "test-volume",
	}

	resp, err := client.GetSnapshot(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to parse snapshot response")

}

func Test_Client_GetSnapshot_MultipleSnapshots(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volume-snapshots", DefaultAPIVersion), r.URL.Path)

		// Return multiple snapshots (should only return 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"items": [
				{
					"id": "12345678-1234-1234-1234-123456789012",
					"name": "test-volume.test-snapshot-uuid",
					"provisioned": 1073741824,
					"created": 1234567890,
					"serial": "TEST123456789",
					"suffix": "test-snapshot-uuid",
					"source": {
						"id": "87654321-4321-4321-4321-210987654321",
						"name": "test-volume"
					}
				},
				{
					"id": "22345678-1234-1234-1234-123456789012",
					"name": "test-volume2.test-snapshot-uuid",
					"provisioned": 2147483648,
					"created": 1234567891,
					"serial": "TEST223456789",
					"suffix": "test-snapshot-uuid",
					"source": {
						"id": "97654321-4321-4321-4321-210987654321",
						"name": "test-volume2"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetSnapshotRequest{
		UUID: "test-snapshot-uuid",
	}

	resp, err := client.GetSnapshot(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "expected exactly 1 snapshot in response, got 2")
}

func Test_Client_GetSnapshots_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volume-snapshots", DefaultAPIVersion), r.URL.Path)
		// The actual query includes the full filter with name pattern and suffix check (URL-encoded)
		require.Contains(t, r.URL.RawQuery, "destroyed=false")
		require.Contains(t, r.URL.RawQuery, "filter=")
		require.Contains(t, r.URL.RawQuery, "contains%28suffix")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"items": [
				{
					"id": "12345678-1234-1234-1234-123456789012",
					"name": "test-volume.test-snapshot-uuid-1",
					"provisioned": 1073741824,
					"created": 1234567890,
					"serial": "TEST123456789",
					"suffix": "test-snapshot-uuid-1",
					"source": {
						"id": "87654321-4321-4321-4321-210987654321",
						"name": "test-volume"
					}
				},
				{
					"id": "22345678-1234-1234-1234-123456789012",
					"name": "test-volume2.test-snapshot-uuid-2",
					"provisioned": 2147483648,
					"created": 1234567891,
					"serial": "TEST223456789",
					"suffix": "test-snapshot-uuid-2",
					"source": {
						"id": "97654321-4321-4321-4321-210987654321",
						"name": "test-volume2"
					}
				},
				{
					"id": "32345678-1234-1234-1234-123456789012",
					"name": "test-volume3.test-snapshot-uuid-3",
					"provisioned": 3221225472,
					"created": 1234567892,
					"serial": "TEST323456789",
					"suffix": "test-snapshot-uuid-3",
					"source": {
						"id": "07654321-4321-4321-4321-210987654321",
						"name": "test-volume3"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetSnapshotsRequest{}

	resp, err := client.GetSnapshots(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Snapshots)
	require.Len(t, resp.Snapshots, 3)

	// Verify first snapshot
	require.Equal(t, "test-snapshot-uuid-1", resp.Snapshots[0].UUID)
	require.Equal(t, uint64(1073741824), resp.Snapshots[0].Size)
	require.Equal(t, "test-volume", resp.Snapshots[0].SourceVolumeUUID)
	require.True(t, resp.Snapshots[0].IsAvailable)

	// Verify second snapshot
	require.Equal(t, "test-snapshot-uuid-2", resp.Snapshots[1].UUID)
	require.Equal(t, uint64(2147483648), resp.Snapshots[1].Size)
	require.Equal(t, "test-volume2", resp.Snapshots[1].SourceVolumeUUID)
	require.True(t, resp.Snapshots[1].IsAvailable)

	// Verify third snapshot
	require.Equal(t, "test-snapshot-uuid-3", resp.Snapshots[2].UUID)
	require.Equal(t, uint64(3221225472), resp.Snapshots[2].Size)
	require.Equal(t, "test-volume3", resp.Snapshots[2].SourceVolumeUUID)
	require.True(t, resp.Snapshots[2].IsAvailable)
}

func Test_Client_GetSnapshots_EmptyList(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volume-snapshots", DefaultAPIVersion), r.URL.Path)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items": []}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetSnapshotsRequest{}

	resp, err := client.GetSnapshots(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "no snapshots found")
}

func Test_Client_GetSnapshots_HTTPError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volume-snapshots", DefaultAPIVersion), r.URL.Path)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errors": [{"message": "Internal server error", "code": "500"}]}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetSnapshotsRequest{}

	resp, err := client.GetSnapshots(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to get snapshots")
}

func Test_Client_GetSnapshots_ItemsEmptyArray(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volume-snapshots", DefaultAPIVersion), r.URL.Path)

		// Return empty items array
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"items": []
		}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetSnapshotsRequest{}

	resp, err := client.GetSnapshots(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "no snapshots found")
}

func Test_Client_GetSnapshots_ParseError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volume-snapshots", DefaultAPIVersion), r.URL.Path)

		// Return invalid JSON
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetSnapshotsRequest{}

	resp, err := client.GetSnapshots(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to get snapshots")
}

func Test_Client_GetSnapshots_NilItems(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volume-snapshots", DefaultAPIVersion), r.URL.Path)

		// Return response without items field
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetSnapshotsRequest{}

	resp, err := client.GetSnapshots(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to parse snapshot response")
}

func Test_Client_GetSnapshots_SingleSnapshot(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volume-snapshots", DefaultAPIVersion), r.URL.Path)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"items": [
				{
					"id": "12345678-1234-1234-1234-123456789012",
					"name": "test-volume.test-snapshot-uuid",
					"provisioned": 1073741824,
					"created": 1234567890,
					"serial": "TEST123456789",
					"suffix": "test-snapshot-uuid",
					"source": {
						"id": "87654321-4321-4321-4321-210987654321",
						"name": "test-volume"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetSnapshotsRequest{}

	resp, err := client.GetSnapshots(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Snapshots)
	require.Len(t, resp.Snapshots, 1)
	require.Equal(t, "test-snapshot-uuid", resp.Snapshots[0].UUID)
	require.Equal(t, uint64(1073741824), resp.Snapshots[0].Size)
	require.Equal(t, "test-volume", resp.Snapshots[0].SourceVolumeUUID)
	require.True(t, resp.Snapshots[0].IsAvailable)
}

func Test_Client_GetSnapshots_SnapshotWithoutSuffix(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, fmt.Sprintf("/api/%s/volume-snapshots", DefaultAPIVersion), r.URL.Path)

		// Return snapshot without suffix field
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"items": [
				{
					"id": "12345678-1234-1234-1234-123456789012",
					"name": "test-volume",
					"provisioned": 1073741824,
					"created": 1234567890,
					"serial": "TEST123456789",
					"source": {
						"id": "87654321-4321-4321-4321-210987654321",
						"name": "test-volume"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	serverURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	endpoint := serverURL.Host

	cfg := &ClientConfig{
		Endpoints: []string{endpoint},
		AuthToken: "test-token",
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	client.sessionToken = "test-session-token"

	req := &models.GetSnapshotsRequest{}

	resp, err := client.GetSnapshots(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Snapshots)
	require.Len(t, resp.Snapshots, 1)
	require.Equal(t, "", resp.Snapshots[0].UUID) // Empty suffix
	require.Equal(t, "test-volume", resp.Snapshots[0].SourceVolumeUUID)
}
