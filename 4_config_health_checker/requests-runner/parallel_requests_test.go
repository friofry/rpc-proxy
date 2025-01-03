package requests_runner_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	requestsrunner "github.com/friofry/config-health-checker/requests-runner"
	rpcprovider "github.com/friofry/config-health-checker/rpc-provider"
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
			Enabled:  true,
			AuthType: rpcprovider.NoAuth,
		},
		{
			Name:      "Provider2",
			URL:       "https://provider2.example.com",
			Enabled:   true,
			AuthType:  rpcprovider.TokenAuth,
			AuthToken: "dummy_token",
		},
		{
			Name:         "Provider3",
			URL:          "https://provider3.example.com",
			Enabled:      true,
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
			timeout: 1 * time.Second,
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
					Enabled:   true,
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
