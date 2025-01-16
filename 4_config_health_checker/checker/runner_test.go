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
	)

	// Run tests
	results := runner.Run(context.Background())

	// Verify results
	assert.NotNil(t, results)
	assert.Contains(t, results, int64(1))

	chainResults := results[1]
	assert.Len(t, chainResults, 2)

	// Verify provider1 results
	provider1Result := chainResults["provider1"]
	assert.True(t, provider1Result.Valid)
	assert.Empty(t, provider1Result.FailedMethods)

	// Verify provider2 results
	provider2Result := chainResults["provider2"]
	assert.False(t, provider2Result.Valid)
	assert.Contains(t, provider2Result.FailedMethods, "eth_blockNumber")
}

func TestChainValidationRunner_RunForChain(t *testing.T) {
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

	// Create mock caller
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
		},
	}

	// Create runner
	runner := NewChainValidationRunner(
		chainCfgs,
		referenceCfgs,
		methodConfigs,
		mockCaller,
		10*time.Second,
	)

	// Test valid chain
	t.Run("valid chain", func(t *testing.T) {
		results, err := runner.RunForChain(context.Background(), 1)
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.Contains(t, results, "provider1")
		assert.True(t, results["provider1"].Valid)
	})

	// Test invalid chain
	t.Run("invalid chain", func(t *testing.T) {
		_, err := runner.RunForChain(context.Background(), 2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "chain config not found")
	})
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
	)

	// Run tests
	results := runner.Run(context.Background())

	// Verify results
	assert.NotNil(t, results)
	assert.Contains(t, results, int64(1))

	chainResults := results[1]
	assert.Contains(t, chainResults, "provider1")
	assert.False(t, chainResults["provider1"].Valid)
	assert.Contains(t, chainResults["provider1"].FailedMethods, "eth_blockNumber")
}
