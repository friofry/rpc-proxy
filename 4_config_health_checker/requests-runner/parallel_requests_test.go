package requests_runner_test

import (
	"context"
	"errors"
	"fmt"
	requestsrunner "github.com/friofry/config-health-checker/requests-runner"
	rpcprovider "github.com/friofry/config-health-checker/rpc-provider"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ParallelCheckProvidersTestSuite defines the test suite for ParallelCheckProviders.
type ParallelCheckProvidersTestSuite struct {
	suite.Suite
	mockClock *clock.Mock
}

// SetupSuite runs once before all tests in the suite.
func (suite *ParallelCheckProvidersTestSuite) SetupSuite() {
	// Initialize the mock clock
	suite.mockClock = clock.NewMock()
}

// TearDownSuite runs once after all tests in the suite.
func (suite *ParallelCheckProvidersTestSuite) TearDownSuite() {
	// No teardown required for mock clock
}

// TestParallelCheckProvidersSuccess tests that all providers are checked successfully.
func (suite *ParallelCheckProvidersTestSuite) TestParallelCheckProvidersSuccess() {
	// Define a list of providers to test
	providers := []rpcprovider.RpcProvider{
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

	// Define a checker function that simulates successful checks
	checker := func(ctx context.Context, provider rpcprovider.RpcProvider) requestsrunner.ProviderResult {
		return requestsrunner.ProviderResult{
			Success:     true,
			Error:       nil,
			Response:    "OK",
			ElapsedTime: 10 * time.Millisecond,
		}
	}

	// Define context and timeout
	ctx := context.Background()

	// Perform parallel checks
	resultsChan := make(chan map[string]requestsrunner.ProviderResult)

	go func() {
		results := requestsrunner.ParallelCheckProviders(ctx, providers, 1*time.Second, checker)
		resultsChan <- results
	}()

	// Receive the results
	results := <-resultsChan

	// Assert that results are as expected
	assert.Len(suite.T(), results, len(providers), "Expected results for all providers")

	for _, provider := range providers {
		result, exists := results[provider.Name]
		assert.True(suite.T(), exists, "Result for %s should exist", provider.Name)
		assert.True(suite.T(), result.Success, "Provider %s should be successful", provider.Name)
		assert.Nil(suite.T(), result.Error, "Provider %s should have no error", provider.Name)
		assert.Equal(suite.T(), "OK", result.Response, "Provider %s should respond with 'OK'", provider.Name)
	}
}

// TestParallelCheckProvidersWithFailures tests handling of providers that fail checks.
func (suite *ParallelCheckProvidersTestSuite) TestParallelCheckProvidersWithFailures() {
	// Define a list of providers to test
	providers := []rpcprovider.RpcProvider{
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

	// Define a checker function that simulates a failure for Provider2
	checker := func(ctx context.Context, provider rpcprovider.RpcProvider) requestsrunner.ProviderResult {
		// Simulate processing time using the mock clock
		if provider.Name == "Provider2" {
			// Simulate a failure
			return requestsrunner.ProviderResult{
				Success:     false,
				Error:       errors.New("connection timeout"),
				Response:    "",
				ElapsedTime: 10 * time.Millisecond,
			}
		}
		// Simulate successful checks for other providers
		return requestsrunner.ProviderResult{
			Success:     true,
			Error:       nil,
			Response:    "OK",
			ElapsedTime: 10 * time.Millisecond,
		}
	}

	// Define context and timeout
	ctx := context.Background()
	timeout := 1 * time.Second

	// Perform parallel checks
	resultsChan := make(chan map[string]requestsrunner.ProviderResult)
	go func() {
		results := requestsrunner.ParallelCheckProviders(ctx, providers, timeout, checker)
		resultsChan <- results
	}()

	// Receive the results
	results := <-resultsChan

	// Assert that results are as expected
	assert.Len(suite.T(), results, len(providers), "Expected results for all providers")

	for _, provider := range providers {
		result, exists := results[provider.Name]
		assert.True(suite.T(), exists, "Result for %s should exist", provider.Name)
		if provider.Name == "Provider2" {
			assert.False(suite.T(), result.Success, "Provider %s should have failed", provider.Name)
			assert.NotNil(suite.T(), result.Error, "Provider %s should have an error", provider.Name)
			assert.Equal(suite.T(), "", result.Response, "Provider %s should have empty response", provider.Name)
			assert.Equal(suite.T(), "connection timeout", result.Error.Error(), "Provider %s should have 'connection timeout' error", provider.Name)
		} else {
			assert.True(suite.T(), result.Success, "Provider %s should be successful", provider.Name)
			assert.Nil(suite.T(), result.Error, "Provider %s should have no error", provider.Name)
			assert.Equal(suite.T(), "OK", result.Response, "Provider %s should respond with 'OK'", provider.Name)
		}
	}
}

// TestParallelCheckProvidersTimeout tests handling of overall timeout.
func (suite *ParallelCheckProvidersTestSuite) TestParallelCheckProvidersTimeout() {
	// Define a list of providers to test
	providers := []rpcprovider.RpcProvider{
		{
			Name:      "Provider1",
			URL:       "https://provider1.example.com",
			Enabled:   true,
			AuthType:  rpcprovider.NoAuth,
			AuthToken: "",
		},
		{
			Name:      "Provider2",
			URL:       "https://provider2.example.com",
			Enabled:   true,
			AuthType:  rpcprovider.TokenAuth,
			AuthToken: "dummy_token",
		},
	}

	// Define a checker function that simulates work
	checker := func(ctx context.Context, provider rpcprovider.RpcProvider) requestsrunner.ProviderResult {
		select {
		case <-time.After(1 * time.Second):
			// Simulate successful response if not canceled
			return requestsrunner.ProviderResult{
				Success:  true,
				Error:    nil,
				Response: "OK",
			}
		case <-ctx.Done():
			// Simulate failure due to context cancellation
			return requestsrunner.ProviderResult{
				Success:  false,
				Error:    ctx.Err(),
				Response: "",
			}
		}
	}
	// Define context and timeout shorter than the simulated request duration
	ctx := context.Background()
	timeout := 1 * time.Millisecond

	// Perform parallel checks
	resultsChan := make(chan map[string]requestsrunner.ProviderResult)
	go func() {
		results := requestsrunner.ParallelCheckProviders(ctx, providers, timeout, checker)
		resultsChan <- results
	}()
	// Receive the results
	results := <-resultsChan

	// Assert that results are as expected
	assert.Len(suite.T(), results, len(providers), "Expected results for all providers")

	for _, provider := range providers {
		result, exists := results[provider.Name]
		assert.True(suite.T(), exists, "Result for %s should exist", provider.Name)
		assert.False(suite.T(), result.Success, "Provider %s should have failed due to timeout", provider.Name)
		assert.NotNil(suite.T(), result.Error, "Provider %s should have an error", provider.Name)
		assert.Contains(suite.T(), result.Error.Error(), "context deadline exceeded", "Provider %s should have timeout error", provider.Name)
		fmt.Println(result.Error.Error())
		assert.Equal(suite.T(), "", result.Response, "Provider %s should have empty response", provider.Name)
	}
}

// TestParallelCheckProvidersPartialSuccess tests a mix of successful and failed provider checks.
func (suite *ParallelCheckProvidersTestSuite) TestParallelCheckProvidersPartialSuccess() {
	// Define a list of providers to test
	providers := []rpcprovider.RpcProvider{
		{
			Name:      "Provider1",
			URL:       "https://provider1.example.com",
			Enabled:   true,
			AuthType:  rpcprovider.NoAuth,
			AuthToken: "",
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

	// Define a checker function that simulates partial successes
	checker := func(ctx context.Context, provider rpcprovider.RpcProvider) requestsrunner.ProviderResult {
		if provider.Name == "Provider2" {
			// Simulate a failure for Provider2
			return requestsrunner.ProviderResult{
				Success:     false,
				Error:       errors.New("authentication failed"),
				Response:    "",
				ElapsedTime: 20 * time.Millisecond,
			}
		}
		// Simulate successful checks for other providers
		return requestsrunner.ProviderResult{
			Success:     true,
			Error:       nil,
			Response:    "OK",
			ElapsedTime: 20 * time.Millisecond,
		}
	}

	// Define context and timeout
	ctx := context.Background()
	timeout := 2 * time.Second

	// Perform parallel checks
	resultsChan := make(chan map[string]requestsrunner.ProviderResult)
	go func() {
		results := requestsrunner.ParallelCheckProviders(ctx, providers, timeout, checker)
		resultsChan <- results
	}()

	// Receive the results
	results := <-resultsChan

	// Assert that results are as expected
	assert.Len(suite.T(), results, len(providers), "Expected results for all providers")

	for _, provider := range providers {
		result, exists := results[provider.Name]
		assert.True(suite.T(), exists, "Result for %s should exist", provider.Name)
		if provider.Name == "Provider2" {
			assert.False(suite.T(), result.Success, "Provider %s should have failed", provider.Name)
			assert.NotNil(suite.T(), result.Error, "Provider %s should have an error", provider.Name)
			assert.Equal(suite.T(), "", result.Response, "Provider %s should have empty response", provider.Name)
			assert.Equal(suite.T(), "authentication failed", result.Error.Error(), "Provider %s should have 'authentication failed' error", provider.Name)
		} else {
			assert.True(suite.T(), result.Success, "Provider %s should be successful", provider.Name)
			assert.Nil(suite.T(), result.Error, "Provider %s should have no error", provider.Name)
			assert.Equal(suite.T(), "OK", result.Response, "Provider %s should respond with 'OK'", provider.Name)
		}
	}
}

// TestParallelCheckProvidersNoProviders tests the function with an empty provider list.
func (suite *ParallelCheckProvidersTestSuite) TestParallelCheckProvidersNoProviders() {
	// Define an empty list of providers
	var providers []rpcprovider.RpcProvider

	// Define a checker function (won't be called)
	checker := func(ctx context.Context, provider rpcprovider.RpcProvider) requestsrunner.ProviderResult {
		return requestsrunner.ProviderResult{
			Success:     true,
			Error:       nil,
			Response:    "OK",
			ElapsedTime: 10 * time.Millisecond,
		}
	}

	// Define context and timeout
	ctx := context.Background()
	timeout := 1 * time.Second

	// Perform parallel checks
	results := requestsrunner.ParallelCheckProviders(ctx, providers, timeout, checker)

	// Assert that results are empty
	assert.Len(suite.T(), results, 0, "Expected no results for empty provider list")
}

// TestParallelCheckProvidersContextCancellation tests handling of context cancellation.
func (suite *ParallelCheckProvidersTestSuite) TestParallelCheckProvidersContextCancellation() {
	// Define a list of providers to test
	providers := []rpcprovider.RpcProvider{
		{
			Name:      "Provider1",
			URL:       "https://provider1.example.com",
			Enabled:   true,
			AuthType:  rpcprovider.NoAuth,
			AuthToken: "",
		},
		{
			Name:      "Provider2",
			URL:       "https://provider2.example.com",
			Enabled:   true,
			AuthType:  rpcprovider.TokenAuth,
			AuthToken: "dummy_token",
		},
	}

	// Define a checker function that simulates work
	checker := func(ctx context.Context, provider rpcprovider.RpcProvider) requestsrunner.ProviderResult {
		select {
		case <-time.After(1 * time.Second):
			// Simulate successful response if not canceled
			return requestsrunner.ProviderResult{
				Success:  true,
				Error:    nil,
				Response: "OK",
			}
		case <-ctx.Done():
			// Simulate failure due to context cancellation
			return requestsrunner.ProviderResult{
				Success:  false,
				Error:    ctx.Err(),
				Response: "",
			}
		}
	}

	// Create a context that will be canceled
	ctx, cancel := context.WithCancel(context.Background())
	timeout := 1 * time.Second

	// Perform parallel checks in a separate goroutine
	resultsChan := make(chan map[string]requestsrunner.ProviderResult)
	go func() {
		results := requestsrunner.ParallelCheckProviders(ctx, providers, timeout, checker)
		resultsChan <- results
	}()

	// Cancel the context before all checks complete
	cancel()

	// Receive the results
	results := <-resultsChan

	// Assert that both providers failed due to context cancellation
	assert.Len(suite.T(), results, len(providers), "Expected results for all providers")

	for _, provider := range providers {
		result, exists := results[provider.Name]
		assert.True(suite.T(), exists, "Result for %s should exist", provider.Name)
		assert.False(suite.T(), result.Success, "Provider %s should have failed due to context cancellation", provider.Name)
		assert.NotNil(suite.T(), result.Error, "Provider %s should have an error", provider.Name)
		assert.Equal(suite.T(), "context canceled", result.Error.Error(), "Provider %s should have 'context canceled' error", provider.Name)
		assert.Equal(suite.T(), "", result.Response, "Provider %s should have empty response", provider.Name)
	}
}

// Run the test suite
func TestParallelCheckProvidersTestSuite(t *testing.T) {
	suite.Run(t, new(ParallelCheckProvidersTestSuite))
}
