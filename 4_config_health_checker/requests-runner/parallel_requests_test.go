package requestsrunner_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	requestsrunner "github.com/friofry/config-health-checker/requests-runner"
	"github.com/friofry/config-health-checker/rpcprovider"
)

// ParallelCheckProvidersTestSuite defines the test suite for ParallelCheckProviders.
type ParallelCheckProvidersTestSuite struct {
	suite.Suite
}

// getSampleProviders returns a predefined list of RPC providers.
func getSampleProviders() []rpcprovider.RpcProvider {
	return []rpcprovider.RpcProvider{
		{
			Name:     "Provider1",
			URL:      "https://provider1.example.com",
			AuthType: rpcprovider.NoAuth,
		},
		{
			Name:      "Provider2",
			URL:       "https://provider2.example.com",
			AuthType:  rpcprovider.TokenAuth,
			AuthToken: "dummy_token",
		},
		{
			Name:         "Provider3",
			URL:          "https://provider3.example.com",
			AuthType:     rpcprovider.BasicAuth,
			AuthLogin:    "user",
			AuthPassword: "pass",
		},
	}
}

// createChecker is a factory function that returns a checker function based on the test case configuration.
// failProviders maps provider names to the errors they should return.
// delay specifies the simulated processing time for each provider.
func createChecker(failProviders map[string]error, delay time.Duration) requestsrunner.RequestFunc {
	return func(ctx context.Context, provider rpcprovider.RpcProvider) requestsrunner.ProviderResult {
		var result requestsrunner.ProviderResult

		// Determine the expected result based on whether the provider should fail.
		if err, shouldFail := failProviders[provider.Name]; shouldFail {
			result = requestsrunner.ProviderResult{
				Success:     false,
				Error:       err,
				Response:    "",
				ElapsedTime: delay,
			}
		} else {
			result = requestsrunner.ProviderResult{
				Success:     true,
				Error:       nil,
				Response:    "OK",
				ElapsedTime: delay,
			}
		}

		select {
		case <-time.After(delay):
			return result
		case <-ctx.Done():
			return requestsrunner.ProviderResult{
				Success:     false,
				Error:       ctx.Err(),
				Response:    "",
				ElapsedTime: 0,
			}
		}
	}
}

// runParallelChecks executes ParallelCheckProviders and returns the results.
func runParallelChecks(ctx context.Context, providers []rpcprovider.RpcProvider, timeout time.Duration, checker requestsrunner.RequestFunc) map[string]requestsrunner.ProviderResult {
	resultsChan := make(chan map[string]requestsrunner.ProviderResult)

	go func() {
		results := requestsrunner.ParallelCheckProviders(ctx, providers, timeout, checker)
		resultsChan <- results
	}()

	return <-resultsChan
}

// assertProviderResults verifies that the actual results match the expected outcomes.
func assertProviderResults(suite *ParallelCheckProvidersTestSuite, providers []rpcprovider.RpcProvider, results map[string]requestsrunner.ProviderResult, expectedResults map[string]requestsrunner.ProviderResult) {
	assert.Len(suite.T(), results, len(providers), "Expected results for all providers")

	for _, provider := range providers {
		result, exists := results[provider.Name]
		assert.True(suite.T(), exists, "Result for %s should exist", provider.Name)

		expected, ok := expectedResults[provider.Name]
		if ok {
			assert.Equal(suite.T(), expected.Success, result.Success, "Provider %s Success status mismatch", provider.Name)
			if expected.Error != nil {
				assert.NotNil(suite.T(), result.Error, "Provider %s should have an error", provider.Name)
				assert.Equal(suite.T(), expected.Error.Error(), result.Error.Error(), "Provider %s Error mismatch", provider.Name)
			} else {
				assert.Nil(suite.T(), result.Error, "Provider %s should have no error", provider.Name)
			}
			assert.Equal(suite.T(), expected.Response, result.Response, "Provider %s Response mismatch", provider.Name)
			assert.Equal(suite.T(), expected.ElapsedTime, result.ElapsedTime, "Provider %s ElapsedTime mismatch", provider.Name)
		}
	}
}

// TestParallelCheckProviders runs all table-driven test cases for ParallelCheckProviders.
func (suite *ParallelCheckProvidersTestSuite) TestParallelCheckProviders() {
	testCases := []struct {
		name            string
		providers       []rpcprovider.RpcProvider
		failProviders   map[string]error
		delay           time.Duration
		timeout         time.Duration
		expectedResults map[string]requestsrunner.ProviderResult
	}{
		{
			name:          "AllProvidersSuccess",
			providers:     getSampleProviders(),
			failProviders: map[string]error{
				// No failures
			},
			delay:   10 * time.Millisecond,
			timeout: 1 * time.Second,
			expectedResults: map[string]requestsrunner.ProviderResult{
				"Provider1": {Success: true, Error: nil, Response: "OK", ElapsedTime: 10 * time.Millisecond},
				"Provider2": {Success: true, Error: nil, Response: "OK", ElapsedTime: 10 * time.Millisecond},
				"Provider3": {Success: true, Error: nil, Response: "OK", ElapsedTime: 10 * time.Millisecond},
			},
		},
		{
			name:      "SomeProvidersFail",
			providers: getSampleProviders(),
			failProviders: map[string]error{
				"Provider2": errors.New("connection timeout"),
			},
			delay:   10 * time.Millisecond,
			timeout: 1 * time.Second,
			expectedResults: map[string]requestsrunner.ProviderResult{
				"Provider1": {Success: true, Error: nil, Response: "OK", ElapsedTime: 10 * time.Millisecond},
				"Provider2": {Success: false, Error: errors.New("connection timeout"), Response: "", ElapsedTime: 10 * time.Millisecond},
				"Provider3": {Success: true, Error: nil, Response: "OK", ElapsedTime: 10 * time.Millisecond},
			},
		},
		{
			name:          "OverallTimeout",
			providers:     getSampleProviders(),
			failProviders: map[string]error{
				// No failures, but delay causes timeout
			},
			delay:   2 * time.Second,
			timeout: 50 * time.Millisecond,
			expectedResults: map[string]requestsrunner.ProviderResult{
				"Provider1": {Success: false, Error: errors.New("context deadline exceeded"), Response: "", ElapsedTime: 0},
				"Provider2": {Success: false, Error: errors.New("context deadline exceeded"), Response: "", ElapsedTime: 0},
				"Provider3": {Success: false, Error: errors.New("context deadline exceeded"), Response: "", ElapsedTime: 0},
			},
		},
		{
			name:      "PartialSuccess",
			providers: getSampleProviders(),
			failProviders: map[string]error{
				"Provider2": errors.New("authentication failed"),
			},
			delay:   20 * time.Millisecond,
			timeout: 2 * time.Second,
			expectedResults: map[string]requestsrunner.ProviderResult{
				"Provider1": {Success: true, Error: nil, Response: "OK", ElapsedTime: 20 * time.Millisecond},
				"Provider2": {Success: false, Error: errors.New("authentication failed"), Response: "", ElapsedTime: 20 * time.Millisecond},
				"Provider3": {Success: true, Error: nil, Response: "OK", ElapsedTime: 20 * time.Millisecond},
			},
		},
		{
			name:            "NoProviders",
			providers:       []rpcprovider.RpcProvider{},
			failProviders:   map[string]error{},
			delay:           10 * time.Millisecond,
			timeout:         1 * time.Second,
			expectedResults: map[string]requestsrunner.ProviderResult{},
		},
		{
			name: "InvalidAuthType",
			providers: []rpcprovider.RpcProvider{
				{
					Name:      "Provider1",
					URL:       "https://provider1.example.com",
					AuthType:  rpcprovider.RpcProviderAuthType("invalid-auth"), // Assuming RpcProviderAuthType is a string alias
					AuthToken: "",
				},
			},
			failProviders: map[string]error{
				"Provider1": errors.New("unknown authentication type"),
			},
			delay:   5 * time.Millisecond,
			timeout: 1 * time.Second,
			expectedResults: map[string]requestsrunner.ProviderResult{
				"Provider1": {Success: false, Error: errors.New("unknown authentication type"), Response: "", ElapsedTime: 5 * time.Millisecond},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		suite.Run(tc.name, func() {
			// Create the checker function using the factory
			checker := createChecker(tc.failProviders, tc.delay)

			// Perform parallel checks with specified timeout
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()

			results := runParallelChecks(ctx, tc.providers, tc.timeout, checker)
			assertProviderResults(suite, tc.providers, results, tc.expectedResults)
		})
	}
}

func TestParallelCallEVMMethods(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","result":"0x1"}`))
	}))
	defer server.Close()

	// Create test providers
	providers := []rpcprovider.RpcProvider{
		{
			Name:     "Provider1",
			URL:      server.URL,
			AuthType: rpcprovider.NoAuth,
		},
		{
			Name:     "Provider2",
			URL:      server.URL,
			AuthType: rpcprovider.NoAuth,
		},
	}

	// Test successful parallel execution
	t.Run("SuccessfulExecution", func(t *testing.T) {
		ctx := context.Background()
		results := requestsrunner.ParallelCallEVMMethods(ctx, providers, "eth_blockNumber", nil, 1*time.Second)

		assert.Len(t, results, len(providers))
		for _, provider := range providers {
			result, exists := results[provider.Name]
			assert.True(t, exists)
			assert.True(t, result.Success)
			assert.Equal(t, `{"jsonrpc":"2.0","result":"0x1"}`, result.Response)
			assert.Nil(t, result.Error)
		}
	})

	// Test timeout
	t.Run("Timeout", func(t *testing.T) {
		// Create a slow test server that responds after 100ms
		slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"jsonrpc":"2.0","result":"0x1"}`))
		}))
		defer slowServer.Close()

		// Create providers pointing to the slow server
		slowProviders := []rpcprovider.RpcProvider{
			{
				Name:     "SlowProvider1",
				URL:      slowServer.URL,
				AuthType: rpcprovider.NoAuth,
			},
			{
				Name:     "SlowProvider2",
				URL:      slowServer.URL,
				AuthType: rpcprovider.NoAuth,
			},
		}

		ctx := context.Background()
		results := requestsrunner.ParallelCallEVMMethods(ctx, slowProviders, "eth_blockNumber", nil, 10*time.Millisecond)

		assert.Len(t, results, len(slowProviders))
		for _, provider := range slowProviders {
			result, exists := results[provider.Name]
			assert.True(t, exists)
			assert.False(t, result.Success, "Expected timeout failure for provider %s", provider.Name)
			assert.Contains(t, result.Error.Error(), "context deadline exceeded", "Expected timeout error for provider %s", provider.Name)
		}
	})

	// Test empty providers
	t.Run("EmptyProviders", func(t *testing.T) {
		ctx := context.Background()
		results := requestsrunner.ParallelCallEVMMethods(ctx, []rpcprovider.RpcProvider{}, "eth_blockNumber", nil, 1*time.Second)
		assert.Empty(t, results)
	})
}

// TestParallelCheckProvidersContextCancellation tests handling of context cancellation.
func (suite *ParallelCheckProvidersTestSuite) TestParallelCheckProvidersContextCancellation() {
	// Use getSampleProviders to define providers
	providers := getSampleProviders()

	checker := createChecker(map[string]error{}, 1*time.Second)

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	timeout := 2 * time.Second // Set timeout longer than checker delay to allow manual cancellation

	// Perform parallel checks in a separate goroutine
	resultsChan := make(chan map[string]requestsrunner.ProviderResult)
	go func() {
		results := requestsrunner.ParallelCheckProviders(ctx, providers, timeout, checker)
		resultsChan <- results
	}()

	// Cancel the context before all checks complete
	time.Sleep(10 * time.Millisecond) // Sleep briefly to ensure ParallelCheckProviders has started
	cancel()

	// Receive the results
	results := <-resultsChan

	// Define expected results: All providers fail due to context cancellation
	expectedResults := make(map[string]requestsrunner.ProviderResult)
	for _, provider := range providers {
		expectedResults[provider.Name] = requestsrunner.ProviderResult{
			Success:     false,
			Error:       errors.New("context canceled"),
			Response:    "",
			ElapsedTime: 0,
		}
	}

	// Assert that results are as expected using assertProviderResults
	assertProviderResults(suite, providers, results, expectedResults)
}

// Run the test suite
func TestParallelCheckProvidersTestSuite(t *testing.T) {
	suite.Run(t, new(ParallelCheckProvidersTestSuite))
}

func TestCallEVMMethod(t *testing.T) {
	tests := []struct {
		name         string
		provider     rpcprovider.RpcProvider
		method       string
		params       []interface{}
		handler      func(http.ResponseWriter, *http.Request)
		wantSuccess  bool
		wantResponse string
		wantError    string
	}{
		{
			name: "Successful NoAuth request",
			provider: rpcprovider.RpcProvider{
				Name:     "test",
				URL:      "", // Will be set to test server URL
				AuthType: rpcprovider.NoAuth,
			},
			method: "eth_blockNumber",
			params: []interface{}{},
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var req map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(t, err)
				assert.Equal(t, "2.0", req["jsonrpc"])
				assert.Equal(t, "eth_blockNumber", req["method"])
				assert.Equal(t, []interface{}{}, req["params"])
				assert.Equal(t, float64(1), req["id"])

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"jsonrpc":"2.0","result":"0x1"}`))
			},
			wantSuccess:  true,
			wantResponse: `{"jsonrpc":"2.0","result":"0x1"}`,
		},
		{
			name: "Successful BasicAuth request",
			provider: rpcprovider.RpcProvider{
				Name:         "test",
				URL:          "", // Will be set to test server URL
				AuthType:     rpcprovider.BasicAuth,
				AuthLogin:    "user",
				AuthPassword: "pass",
			},
			method: "eth_getBalance",
			params: []interface{}{"0x123", "latest"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				user, pass, ok := r.BasicAuth()
				assert.True(t, ok)
				assert.Equal(t, "user", user)
				assert.Equal(t, "pass", pass)

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"jsonrpc":"2.0","result":"0x100"}`))
			},
			wantSuccess:  true,
			wantResponse: `{"jsonrpc":"2.0","result":"0x100"}`,
		},
		{
			name: "Successful TokenAuth request",
			provider: rpcprovider.RpcProvider{
				Name:      "test",
				URL:       "", // Will be set to test server URL
				AuthType:  rpcprovider.TokenAuth,
				AuthToken: "test-token",
			},
			method: "eth_chainId",
			params: []interface{}{},
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Verify token is in URL path
				assert.Contains(t, r.URL.String(), "test-token")

				// Verify no Authorization header
				assert.Empty(t, r.Header.Get("Authorization"))

				// Verify JSON-RPC request
				var req map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(t, err)
				assert.Equal(t, "2.0", req["jsonrpc"])
				assert.Equal(t, "eth_chainId", req["method"])
				assert.Equal(t, []interface{}{}, req["params"])
				assert.Equal(t, float64(1), req["id"])

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"jsonrpc":"2.0","result":"0x1"}`))
			},
			wantSuccess:  true,
			wantResponse: `{"jsonrpc":"2.0","result":"0x1"}`,
		},
		{
			name: "Server error response",
			provider: rpcprovider.RpcProvider{
				Name:     "test",
				URL:      "", // Will be set to test server URL
				AuthType: rpcprovider.NoAuth,
			},
			method: "eth_blockNumber",
			params: []interface{}{},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantSuccess: false,
			wantError:   "500",
		},
		{
			name: "Invalid JSON response",
			provider: rpcprovider.RpcProvider{
				Name:     "test",
				URL:      "", // Will be set to test server URL
				AuthType: rpcprovider.NoAuth,
			},
			method: "eth_blockNumber",
			params: []interface{}{},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`invalid json`))
			},
			wantSuccess:  true,
			wantResponse: "invalid json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(tt.handler))
			defer server.Close()

			// Update provider URL
			tt.provider.URL = server.URL

			// Call the method
			result := requestsrunner.CallEVMMethod(context.Background(), tt.provider, tt.method, tt.params)

			// Verify results
			assert.Equal(t, tt.wantSuccess, result.Success)
			if tt.wantResponse != "" {
				assert.Equal(t, tt.wantResponse, result.Response)
			}
			if tt.wantError != "" {
				assert.Contains(t, result.Error.Error(), tt.wantError)
			}
		})
	}
}
