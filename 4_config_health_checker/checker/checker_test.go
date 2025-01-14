package checker

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	requestsrunner "github.com/friofry/config-health-checker/requests-runner"
	"github.com/friofry/config-health-checker/rpcprovider"
	"github.com/stretchr/testify/assert"
)

func TestTestEVMMethod(t *testing.T) {
	// Create test context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Create mock providers
	referenceProvider := rpcprovider.RpcProvider{
		Name:     "reference",
		URL:      "http://reference.com",
		AuthType: rpcprovider.NoAuth,
	}

	validProvider := rpcprovider.RpcProvider{
		Name:     "valid",
		URL:      "http://valid.com",
		AuthType: rpcprovider.NoAuth,
	}

	invalidProvider := rpcprovider.RpcProvider{
		Name:     "invalid",
		URL:      "http://invalid.com",
		AuthType: rpcprovider.NoAuth,
	}

	errorProvider := rpcprovider.RpcProvider{
		Name:     "error",
		URL:      "http://error.com",
		AuthType: rpcprovider.NoAuth,
	}

	// Create mock EVMMethodCaller
	mockCaller := &mockEVMMethodCaller{
		responses: map[string]requestsrunner.ProviderResult{
			"reference": {
				Success:  true,
				Response: `{"result":"0x64"}`,
			},
			"valid": {
				Success:  true,
				Response: `{"result":"0x65"}`,
			},
			"invalid": {
				Success:  true,
				Response: `{"result":"0x6e"}`,
			},
			"error": {
				Success: false,
				Error:   errors.New("connection error"),
			},
		},
	}

	// Create comparison function
	compareFunc := func(reference, result *big.Int) bool {
		diff := new(big.Int).Abs(new(big.Int).Sub(result, reference))
		return diff.Cmp(big.NewInt(2)) <= 0
	}

	t.Run("successful validation", func(t *testing.T) {
		results := TestEVMMethodWithCaller(ctx, EVMMethodTestConfig{
			Method:      "eth_blockNumber",
			Params:      nil,
			CompareFunc: compareFunc,
		}, mockCaller, []rpcprovider.RpcProvider{validProvider}, referenceProvider, 500*time.Millisecond)

		assert.Len(t, results, 1)
		assert.True(t, results["valid"].Valid)
	})

	t.Run("invalid result", func(t *testing.T) {
		results := TestEVMMethodWithCaller(ctx, EVMMethodTestConfig{
			Method:      "eth_blockNumber",
			Params:      nil,
			CompareFunc: compareFunc,
		}, mockCaller, []rpcprovider.RpcProvider{invalidProvider}, referenceProvider, 500*time.Millisecond)

		assert.Len(t, results, 1)
		assert.False(t, results["invalid"].Valid)
	})

	t.Run("provider error", func(t *testing.T) {
		results := TestEVMMethodWithCaller(ctx, EVMMethodTestConfig{
			Method:      "eth_blockNumber",
			Params:      nil,
			CompareFunc: compareFunc,
		}, mockCaller, []rpcprovider.RpcProvider{errorProvider}, referenceProvider, 500*time.Millisecond)

		assert.Len(t, results, 1)
		assert.False(t, results["error"].Valid)
		assert.Error(t, results["error"].Error)
	})

	t.Run("reference provider failure", func(t *testing.T) {
		// Create new mock with failing reference provider
		failingMock := &mockEVMMethodCaller{
			responses: map[string]requestsrunner.ProviderResult{
				"reference": {
					Success: false,
					Error:   errors.New("reference failed"),
				},
				"valid": {
					Success:  true,
					Response: `{"result":"0x65"}`,
				},
			},
		}

		results := TestEVMMethodWithCaller(ctx, EVMMethodTestConfig{
			Method:      "eth_blockNumber",
			Params:      nil,
			CompareFunc: compareFunc,
		}, failingMock, []rpcprovider.RpcProvider{validProvider}, referenceProvider, 500*time.Millisecond)

		assert.Len(t, results, 1)
		assert.False(t, results["valid"].Valid)
		assert.Error(t, results["valid"].Error)
	})
}

// mockEVMMethodCaller implements the EVMMethodCaller interface for testing
type mockEVMMethodCaller struct {
	responses map[string]requestsrunner.ProviderResult
}

func (m *mockEVMMethodCaller) CallEVMMethod(
	ctx context.Context,
	provider rpcprovider.RpcProvider,
	method string,
	params []interface{},
	timeout time.Duration,
) requestsrunner.ProviderResult {
	return m.responses[provider.Name]
}
