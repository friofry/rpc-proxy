package checker

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/friofry/config-health-checker/chainconfig"
	requestsrunner "github.com/friofry/config-health-checker/requests-runner"
	"github.com/friofry/config-health-checker/rpcprovider"
	"github.com/stretchr/testify/assert"
)

// MockEVMMethodCaller implements EVMMethodCaller for testing
type MockEVMMethodCaller struct {
	results map[string]requestsrunner.ProviderResult
}

func (m *MockEVMMethodCaller) CallEVMMethod(
	ctx context.Context,
	provider rpcprovider.RpcProvider,
	method string,
	params []interface{},
	timeout time.Duration,
) requestsrunner.ProviderResult {
	return m.results[provider.Name]
}

func TestChainValidationRunner_Run(t *testing.T) {
	// Setup test data
	chainCfgs := map[int64]chainconfig.ChainConfig{
		1: {
			Providers: []rpcprovider.RpcProvider{
				{Name: "provider1"},
				{Name: "provider2"},
			},
		},
	}

	referenceCfgs := map[int64]chainconfig.ReferenceChainConfig{
		1: {
			Provider: rpcprovider.RpcProvider{Name: "reference"},
		},
	}

	methodConfigs := []EVMMethodTestConfig{
		{
			Method: "eth_blockNumber",
			CompareFunc: func(ref, res *big.Int) bool {
				return ref.Cmp(res) == 0
			},
		},
	}

	// Create mock caller with predefined results
	mockCaller := &MockEVMMethodCaller{
		results: map[string]requestsrunner.ProviderResult{
			"reference": {
				Success:  true,
				Response: `{"result":"0x1234"}`,
			},
			"provider1": {
				Success:  true,
				Response: `{"result":"0x1234"}`,
			},
			"provider2": {
				Success:  true,
				Response: `{"result":"0x5678"}`,
			},
		},
	}

	// Create runner
	runner := NewChainValidationRunner(
		chainCfgs,
		referenceCfgs,
		methodConfigs,
		mockCaller,
		10*time.Second,
		"", // Empty output path for tests
	)

	// Run tests
	runner.Run(context.Background())

	// Since Run no longer returns results, we need to verify the output file
	// or other side effects instead. For now, just verify it runs without errors.
}

func TestChainValidationRunner_ReferenceProviderFailure(t *testing.T) {
	// Setup test data
	chainCfgs := map[int64]chainconfig.ChainConfig{
		1: {
			Providers: []rpcprovider.RpcProvider{
				{Name: "provider1"},
			},
		},
	}

	referenceCfgs := map[int64]chainconfig.ReferenceChainConfig{
		1: {
			Provider: rpcprovider.RpcProvider{Name: "reference"},
		},
	}

	methodConfigs := []EVMMethodTestConfig{
		{
			Method: "eth_blockNumber",
			CompareFunc: func(ref, res *big.Int) bool {
				return ref.Cmp(res) == 0
			},
		},
	}

	// Create mock caller with failing reference provider
	mockCaller := &MockEVMMethodCaller{
		results: map[string]requestsrunner.ProviderResult{
			"reference": {
				Success: false,
				Error:   errors.New("reference failed"),
			},
			"provider1": {
				Success:  true,
				Response: `{"result":"0x1234"}`,
			},
		},
	}

	// Create runner
	runner := NewChainValidationRunner(
		chainCfgs,
		referenceCfgs,
		methodConfigs,
		mockCaller,
		10*time.Second,
		"", // Empty output path for tests
	)

	// Run tests
	runner.Run(context.Background())

	// Since Run no longer returns results, we need to verify the output file
	// or other side effects instead. For now, just verify it runs without errors.
}

func TestChainValidationRunner_ValidateChains(t *testing.T) {
	// Setup test data
	chainCfgs := map[int64]chainconfig.ChainConfig{
		1: {
			Providers: []rpcprovider.RpcProvider{
				{Name: "provider1"},
				{Name: "provider2"},
			},
		},
	}

	referenceCfgs := map[int64]chainconfig.ReferenceChainConfig{
		1: {
			Provider: rpcprovider.RpcProvider{Name: "reference"},
		},
	}

	methodConfigs := []EVMMethodTestConfig{
		{
			Method: "eth_blockNumber",
			CompareFunc: func(ref, res *big.Int) bool {
				return ref.Cmp(res) == 0
			},
		},
	}

	// Create mock caller with predefined results
	mockCaller := &MockEVMMethodCaller{
		results: map[string]requestsrunner.ProviderResult{
			"reference": {
				Success:  true,
				Response: `{"result":"0x1234"}`,
			},
			"provider1": {
				Success:  true,
				Response: `{"result":"0x1234"}`,
			},
			"provider2": {
				Success:  true,
				Response: `{"result":"0x5678"}`,
			},
		},
	}

	// Create runner
	runner := NewChainValidationRunner(
		chainCfgs,
		referenceCfgs,
		methodConfigs,
		mockCaller,
		10*time.Second,
		"", // Empty output path for tests
	)

	// Test validateChains
	t.Run("valid chains", func(t *testing.T) {
		results := make(map[int64]map[string]ProviderValidationResult)
		runner.validateChains(context.Background(), results)

		assert.Contains(t, results, int64(1))

		chainResults := results[1]
		assert.Contains(t, chainResults, "provider1")
		assert.True(t, chainResults["provider1"].Valid)
		assert.Contains(t, chainResults, "provider2")
		assert.False(t, chainResults["provider2"].Valid)
	})

	t.Run("failed methods tracking", func(t *testing.T) {
		results := make(map[int64]map[string]ProviderValidationResult)
		runner.validateChains(context.Background(), results)

		assert.Contains(t, results, int64(1))

		chainResults := results[1]
		assert.Contains(t, chainResults, "provider1")
		assert.True(t, chainResults["provider1"].Valid)
		assert.Contains(t, chainResults, "provider2")
		assert.False(t, chainResults["provider2"].Valid)
		assert.Contains(t, chainResults["provider2"].FailedMethods, "eth_blockNumber")
	})
}
