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

// EVMMethodTestConfig contains configuration for testing an EVM method
type EVMMethodTestConfig struct {
	Method      string
	Params      []interface{}
	CompareFunc func(reference, result *big.Int) bool
}

// MultiMethodTestResult contains results for multiple method tests
type MultiMethodTestResult struct {
	Results map[string]CheckResult // method -> result
}

// TestEVMMethodWithCaller tests a single EVM method
func TestEVMMethodWithCaller(
	ctx context.Context,
	config EVMMethodTestConfig,
	caller EVMMethodCaller,
	providers []rpcprovider.RpcProvider,
	referenceProvider rpcprovider.RpcProvider,
	timeout time.Duration,
) map[string]CheckResult {
	// Combine reference provider with test providers
	allProviders := append([]rpcprovider.RpcProvider{referenceProvider}, providers...)

	// Execute the EVM method using the provided caller
	results := make(map[string]requestsrunner.ProviderResult)
	for _, provider := range allProviders {
		results[provider.Name] = caller.CallEVMMethod(ctx, provider, config.Method, config.Params, timeout)
	}

	// Extract reference result
	refResult, refExists := results[referenceProvider.Name]
	if !refExists || !refResult.Success {
		return handleReferenceFailure(results, referenceProvider.Name)
	}

	// Parse reference value
	refValue, err := parseJSONRPCResult(refResult.Response)
	if err != nil {
		return handleReferenceParseError(results, referenceProvider.Name, err)
	}

	// Compare each provider's result to reference
	checkResults := make(map[string]CheckResult)
	for _, provider := range providers {
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

		// Use provided comparison function
		valid := config.CompareFunc(refValue, providerValue)

		checkResults[provider.Name] = CheckResult{
			Valid:  valid,
			Result: result,
		}
	}

	return checkResults
}

// CheckResult contains the validation result for a provider
type CheckResult struct {
	Valid  bool
	Result requestsrunner.ProviderResult
	Error  error
}

// TestMultipleEVMMethods runs multiple EVM method tests and returns results per provider per method
func TestMultipleEVMMethods(
	ctx context.Context,
	methodConfigs []EVMMethodTestConfig, // list of method configs
	caller EVMMethodCaller,
	providers []rpcprovider.RpcProvider,
	referenceProvider rpcprovider.RpcProvider,
	timeout time.Duration,
) map[string]map[string]CheckResult { // provider -> method -> result
	results := make(map[string]map[string]CheckResult)

	// Initialize result structure
	for _, provider := range providers {
		results[provider.Name] = make(map[string]CheckResult)
	}

	// Run tests for each method
	for _, config := range methodConfigs {
		methodResults := TestEVMMethodWithCaller(ctx, config, caller, providers, referenceProvider, timeout)

		// Store results per provider using method name from config
		for providerName, result := range methodResults {
			results[providerName][config.Method] = result
		}
	}

	return results
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

// ValidateMultipleEVMMethods runs multiple EVM method tests and returns validation summary
func ValidateMultipleEVMMethods(
	ctx context.Context,
	methodConfigs []EVMMethodTestConfig,
	caller EVMMethodCaller,
	providers []rpcprovider.RpcProvider,
	referenceProvider rpcprovider.RpcProvider,
	timeout time.Duration,
) map[string]ProviderValidationResult {
	// Run all method tests
	methodResults := TestMultipleEVMMethods(ctx, methodConfigs, caller, providers, referenceProvider, timeout)

	// Prepare validation results
	validationResults := make(map[string]ProviderValidationResult)

	for providerName, results := range methodResults {
		// Track failed methods
		failedMethods := make(map[string]FailedMethodResult)
		allValid := true

		for method, result := range results {
			if !result.Valid {
				allValid = false
				// Get reference result for this method
				refResult := methodResults[referenceProvider.Name][method]

				failedMethods[method] = FailedMethodResult{
					Result:          result.Result,
					ReferenceResult: refResult.Result,
				}
			}
		}

		validationResults[providerName] = ProviderValidationResult{
			Valid:         allValid,
			FailedMethods: failedMethods,
		}
	}

	return validationResults
}

// ProviderValidationResult contains aggregated validation results for a provider
type ProviderValidationResult struct {
	Valid         bool                          // Overall validation status
	FailedMethods map[string]FailedMethodResult // Map of failed test methods to their results
}

// FailedMethodResult contains details about a failed method test
type FailedMethodResult struct {
	Result          requestsrunner.ProviderResult // Raw result from the provider
	ReferenceResult requestsrunner.ProviderResult // Raw result from the reference provider
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
