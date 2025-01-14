package checker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	requestsrunner "github.com/friofry/config-health-checker/requests-runner"
	"github.com/friofry/config-health-checker/rpcprovider"
)

// EVMMethodCaller defines the interface for calling EVM methods
type EVMMethodCaller interface {
	CallEVMMethod(
		ctx context.Context,
		provider rpcprovider.RpcProvider,
		method string,
		params []interface{},
		timeout time.Duration,
	) requestsrunner.ProviderResult
}

type Test struct {
	Name string
	Func requestsrunner.RequestFunc
}

// TestEVMMethodWithCaller is a version of TestEVMMethod that accepts an EVMMethodCaller
// interface for dependency injection, primarily for testing
func TestEVMMethodWithCaller(
	ctx context.Context,
	config EVMMethodTestConfig,
	caller EVMMethodCaller,
) map[string]CheckResult {
	// Combine reference provider with test providers
	allProviders := append([]rpcprovider.RpcProvider{config.ReferenceProvider}, config.Providers...)

	// Execute the EVM method using the provided caller
	results := make(map[string]requestsrunner.ProviderResult)
	for _, provider := range allProviders {
		results[provider.Name] = caller.CallEVMMethod(ctx, provider, config.Method, config.Params, config.Timeout)
	}

	// Extract reference result
	refResult, refExists := results[config.ReferenceProvider.Name]
	if !refExists || !refResult.Success {
		return handleReferenceFailure(results, config.ReferenceProvider.Name)
	}

	// Parse reference value
	refValue, err := parseJSONRPCResult(refResult.Response)
	if err != nil {
		return handleReferenceParseError(results, config.ReferenceProvider.Name, err)
	}

	// Compare each provider's result to reference
	checkResults := make(map[string]CheckResult)
	for _, provider := range config.Providers {
		result, exists := results[provider.Name]
		if !exists {
			checkResults[provider.Name] = CheckResult{
				Valid: false,
				Error: errors.New("provider result not found"),
			}
			continue
		}

		// Handle failed requests
		if !result.Success {
			checkResults[provider.Name] = CheckResult{
				Valid:  false,
				Result: result,
				Error:  result.Error,
			}
			continue
		}

		// Parse provider's result
		providerValue, err := parseJSONRPCResult(result.Response)
		if err != nil {
			checkResults[provider.Name] = CheckResult{
				Valid:  false,
				Result: result,
				Error:  fmt.Errorf("failed to parse provider response: %w", err),
			}
			continue
		}

		// Calculate difference from reference
		diff := new(big.Int).Abs(new(big.Int).Sub(providerValue, refValue))
		valid := diff.Cmp(config.MaxDiff) <= 0

		checkResults[provider.Name] = CheckResult{
			Valid:  valid,
			Result: result,
			Diff:   diff,
		}
	}

	return checkResults
}

// CheckResult contains the validation result for a provider
type CheckResult struct {
	Valid  bool
	Result requestsrunner.ProviderResult
	Diff   *big.Int // Difference from reference value
	Error  error
}

// EVMMethodTestConfig contains configuration for testing an EVM method
type EVMMethodTestConfig struct {
	ReferenceProvider rpcprovider.RpcProvider
	Providers         []rpcprovider.RpcProvider
	Method            string
	Params            []interface{}
	MaxDiff           *big.Int
	Timeout           time.Duration
}

// handleReferenceFailure handles cases where reference provider fails
func handleReferenceFailure(results map[string]requestsrunner.ProviderResult, refName string) map[string]CheckResult {
	checkResults := make(map[string]CheckResult)

	// Mark all non-reference providers as invalid due to reference failure
	for name, result := range results {
		if name != refName {
			checkResults[name] = CheckResult{
				Valid:  false,
				Result: result,
				Error:  fmt.Errorf("validation failed: reference provider %s failed", refName),
			}
		}
	}

	return checkResults
}

// handleReferenceParseError handles cases where reference result cannot be parsed
func handleReferenceParseError(results map[string]requestsrunner.ProviderResult, refName string, err error) map[string]CheckResult {
	checkResults := make(map[string]CheckResult)
	for name, result := range results {
		checkResults[name] = CheckResult{
			Valid:  false,
			Result: result,
			Error:  fmt.Errorf("failed to parse reference provider %s response: %w", refName, err),
		}
	}
	return checkResults
}

// parseJSONRPCResult extracts the numeric result from a JSON-RPC response
func parseJSONRPCResult(response string) (*big.Int, error) {
	var jsonResponse struct {
		Result string `json:"result"`
	}

	if err := json.Unmarshal([]byte(response), &jsonResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON-RPC response: %w", err)
	}

	// Remove 0x prefix if present
	resultStr := jsonResponse.Result
	if len(resultStr) > 2 && resultStr[0:2] == "0x" {
		resultStr = resultStr[2:]
	}

	value, ok := new(big.Int).SetString(resultStr, 16)
	if !ok {
		return nil, errors.New("failed to parse result as hex number")
	}

	return value, nil
}
