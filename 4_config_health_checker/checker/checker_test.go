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

	t.Run("successful validation", func(t *testing.T) {
		maxDiff := big.NewInt(2)
		results := TestEVMMethodWithCaller(ctx, EVMMethodTestConfig{
			ReferenceProvider: referenceProvider,
			Providers:         []rpcprovider.RpcProvider{validProvider},
			Method:            "eth_blockNumber",
			Params:            nil,
			MaxDiff:           maxDiff,
			Timeout:           500 * time.Millisecond,
		}, mockCaller)

		assert.Len(t, results, 1)
		assert.True(t, results["valid"].Valid)
		assert.Equal(t, big.NewInt(1), results["valid"].Diff)
	})

	t.Run("invalid result", func(t *testing.T) {
		maxDiff := big.NewInt(2)
		results := TestEVMMethodWithCaller(ctx, EVMMethodTestConfig{
			ReferenceProvider: referenceProvider,
			Providers:         []rpcprovider.RpcProvider{invalidProvider},
			Method:            "eth_blockNumber",
			Params:            nil,
			MaxDiff:           maxDiff,
			Timeout:           500 * time.Millisecond,
		}, mockCaller)

		assert.Len(t, results, 1)
		assert.False(t, results["invalid"].Valid)
		assert.Equal(t, big.NewInt(10), results["invalid"].Diff)
	})

	t.Run("provider error", func(t *testing.T) {
		maxDiff := big.NewInt(2)
		results := TestEVMMethodWithCaller(ctx, EVMMethodTestConfig{
			ReferenceProvider: referenceProvider,
			Providers:         []rpcprovider.RpcProvider{errorProvider},
			Method:            "eth_blockNumber",
			Params:            nil,
			MaxDiff:           maxDiff,
			Timeout:           500 * time.Millisecond,
		}, mockCaller)

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

		maxDiff := big.NewInt(2)
		results := TestEVMMethodWithCaller(ctx, EVMMethodTestConfig{
			ReferenceProvider: referenceProvider,
			Providers:         []rpcprovider.RpcProvider{validProvider},
			Method:            "eth_blockNumber",
			Params:            nil,
			MaxDiff:           maxDiff,
			Timeout:           500 * time.Millisecond,
		}, failingMock)

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
