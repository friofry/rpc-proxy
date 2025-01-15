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

func TestValidateMultipleEVMMethods(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Create mock providers
	referenceProvider := rpcprovider.RpcProvider{
		Name:     "reference",
		URL:      "http://reference.com",
		AuthType: rpcprovider.NoAuth,
	}

	providerA := rpcprovider.RpcProvider{
		Name:     "providerA",
		URL:      "http://providerA.com",
		AuthType: rpcprovider.NoAuth,
	}

	providerB := rpcprovider.RpcProvider{
		Name:     "providerB",
		URL:      "http://providerB.com",
		AuthType: rpcprovider.NoAuth,
	}

	// Create mock EVMMethodCaller
	mockCaller := &mockEVMMethodCaller{
		responses: map[string]requestsrunner.ProviderResult{
			"reference": {
				Success:  true,
				Response: `{"result":"0x64"}`,
			},
			"providerA": {
				Success:  true,
				Response: `{"result":"0x65"}`,
			},
			"providerB": {
				Success:  true,
				Response: `{"result":"0x6e"}`,
			},
		},
	}

	// Create comparison function
	compareFunc := func(reference, result *big.Int) bool {
		diff := new(big.Int).Abs(new(big.Int).Sub(result, reference))
		return diff.Cmp(big.NewInt(2)) <= 0
	}

	// Define multiple method tests
	methodConfigs := []EVMMethodTestConfig{
		{
			Method:      "eth_blockNumber",
			Params:      nil,
			CompareFunc: compareFunc,
		},
		{
			Method:      "eth_chainId",
			Params:      nil,
			CompareFunc: compareFunc,
		},
	}

	t.Run("successful validation with some failures", func(t *testing.T) {
		results := ValidateMultipleEVMMethods(
			ctx,
			methodConfigs,
			mockCaller,
			[]rpcprovider.RpcProvider{providerA, providerB},
			referenceProvider,
			500*time.Millisecond,
		)

		// Verify results structure
		assert.Len(t, results, 2)
		assert.Contains(t, results, "providerA")
		assert.Contains(t, results, "providerB")

		// Verify providerA results
		providerAResults := results["providerA"]
		assert.True(t, providerAResults.Valid)
		assert.Len(t, providerAResults.FailedMethods, 0)

		// Verify providerB results
		providerBResults := results["providerB"]
		assert.False(t, providerBResults.Valid)
		assert.Len(t, providerBResults.FailedMethods, 2)
		assert.Contains(t, providerBResults.FailedMethods, "eth_blockNumber")
		assert.Contains(t, providerBResults.FailedMethods, "eth_chainId")
	})

	t.Run("reference provider failure", func(t *testing.T) {
		// Create failing reference mock
		failingMock := &mockEVMMethodCaller{
			responses: map[string]requestsrunner.ProviderResult{
				"reference": {
					Success: false,
					Error:   errors.New("reference failed"),
				},
				"providerA": {
					Success:  true,
					Response: `{"result":"0x65"}`,
				},
			},
		}

		results := ValidateMultipleEVMMethods(
			ctx,
			methodConfigs,
			failingMock,
			[]rpcprovider.RpcProvider{providerA},
			referenceProvider,
			500*time.Millisecond,
		)

		// Verify all results are invalid due to reference failure
		providerAResults := results["providerA"]
		assert.False(t, providerAResults.Valid)
		assert.Len(t, providerAResults.FailedMethods, 2)
	})

	t.Run("partial provider failures", func(t *testing.T) {
		partialMock := &mockEVMMethodCaller{
			responses: map[string]requestsrunner.ProviderResult{
				"reference": {
					Success:  true,
					Response: `{"result":"0x64"}`,
				},
				"providerA": {
					Success:  true,
					Response: `{"result":"0x65"}`,
				},
			},
			methodResponses: map[string]map[string]requestsrunner.ProviderResult{
				"providerA": {
					"eth_blockNumber": {
						Success:  true,
						Response: `{"result":"0x65"}`,
					},
					"eth_chainId": {
						Success: false,
						Error:   errors.New("method failed"),
					},
				},
			},
		}

		results := ValidateMultipleEVMMethods(
			ctx,
			methodConfigs,
			partialMock,
			[]rpcprovider.RpcProvider{providerA},
			referenceProvider,
			500*time.Millisecond,
		)

		// Verify partial failure results
		providerAResults := results["providerA"]
		assert.False(t, providerAResults.Valid)
		assert.Len(t, providerAResults.FailedMethods, 1)
		assert.Contains(t, providerAResults.FailedMethods, "eth_chainId")
	})
}

// mockEVMMethodCaller implements the EVMMethodCaller interface for testing
type mockEVMMethodCaller struct {
	responses       map[string]requestsrunner.ProviderResult
	methodResponses map[string]map[string]requestsrunner.ProviderResult
}

func (m *mockEVMMethodCaller) CallEVMMethod(
	ctx context.Context,
	provider rpcprovider.RpcProvider,
	method string,
	params []interface{},
	timeout time.Duration,
) requestsrunner.ProviderResult {
	// Check if there are method-specific responses
	if methodResponses, ok := m.methodResponses[provider.Name]; ok {
		if response, ok := methodResponses[method]; ok {
			return response
		}
	}
	// Fall back to general provider response
	return m.responses[provider.Name]
}
