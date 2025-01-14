package requestsrunner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	rpcprovider "github.com/friofry/config-health-checker/rpcprovider"
)

// ProviderResult contains information about the result of a provider check.
type ProviderResult struct {
	Success     bool          // Indicates if the request was successful
	Error       error         // Error if the request failed
	Response    string        // Response from the provider (if successful)
	ElapsedTime time.Duration // Duration taken to perform the request
}

// CallEVMMethod makes an HTTP POST request to an RPC provider for a specific EVM method
func CallEVMMethod(ctx context.Context, provider rpcprovider.RpcProvider, method string, params []interface{}) ProviderResult {
	startTime := time.Now()

	// Create JSON-RPC 2.0 request body
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      1,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return ProviderResult{
			Success:     false,
			Error:       fmt.Errorf("failed to marshal request body: %w", err),
			Response:    "",
			ElapsedTime: time.Since(startTime),
		}
	}

	// Create HTTP client with timeout from context
	client := &http.Client{}
	req, err := http.NewRequest("POST", provider.URL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return ProviderResult{
			Success:     false,
			Error:       fmt.Errorf("failed to create request: %w", err),
			Response:    "",
			ElapsedTime: time.Since(startTime),
		}
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Set authentication based on provider type
	switch provider.AuthType {
	case rpcprovider.BasicAuth:
		req.SetBasicAuth(provider.AuthLogin, provider.AuthPassword)
	case rpcprovider.TokenAuth:
		req.URL.RawQuery = provider.AuthToken
	}

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return ProviderResult{
			Success:     false,
			Error:       fmt.Errorf("request failed: %w", err),
			Response:    "",
			ElapsedTime: time.Since(startTime),
		}
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusOK {
		return ProviderResult{
			Success:     false,
			Error:       fmt.Errorf("unexpected status code: %d", resp.StatusCode),
			Response:    "",
			ElapsedTime: time.Since(startTime),
		}
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ProviderResult{
			Success:     false,
			Error:       fmt.Errorf("failed to read response: %w", err),
			Response:    "",
			ElapsedTime: time.Since(startTime),
		}
	}

	return ProviderResult{
		Success:     true,
		Error:       nil,
		Response:    string(body),
		ElapsedTime: time.Since(startTime),
	}
}

// RequestFunc defines the type of function used to check a provider.
type RequestFunc func(ctx context.Context, provider rpcprovider.RpcProvider) ProviderResult

// ParallelCheckProviders performs concurrent checks on multiple RPC providers using the provided checker function.
// It does not limit the number of concurrent goroutines.
func ParallelCheckProviders(ctx context.Context, providers []rpcprovider.RpcProvider, timeout time.Duration, checker RequestFunc) map[string]ProviderResult {
	results := make(map[string]ProviderResult)
	resultsChan := make(chan struct {
		name   string
		result ProviderResult
	}, len(providers)) // Buffered channel to collect results

	// Create a child context with the specified timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var wg sync.WaitGroup

	for _, provider := range providers {
		// Increment the WaitGroup counter
		wg.Add(1)

		// Launch a goroutine for each provider
		go func(p rpcprovider.RpcProvider) {
			defer wg.Done()

			// Create a temporary channel to receive checker result
			tempChan := make(chan ProviderResult, 1)

			// Run the checker function in a separate goroutine
			go func() {
				result := checker(ctx, p)
				tempChan <- result
			}()

			// Wait for either the checker function to finish or the context to be done
			select {
			case res := <-tempChan:
				// Checker function completed
				resultsChan <- struct {
					name   string
					result ProviderResult
				}{name: p.Name, result: res}
			case <-ctx.Done():
				// Context canceled or timed out
				fmt.Println("Context canceled", p.Name)
				resultsChan <- struct {
					name   string
					result ProviderResult
				}{name: p.Name, result: ProviderResult{Success: false, Error: ctx.Err(), Response: ""}}
			}
		}(provider)
	}

	// Launch a goroutine to close the results channel once all checks are done
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results from the channel
	for entry := range resultsChan {
		results[entry.name] = entry.result
	}

	return results
}

// ParallelCallEVMMethods executes EVM methods in parallel across multiple providers
func ParallelCallEVMMethods(ctx context.Context, providers []rpcprovider.RpcProvider, method string, params []interface{}, timeout time.Duration) map[string]ProviderResult {
	// Create a RequestFunc that wraps CallEVMMethod with the given method and params
	checker := func(ctx context.Context, provider rpcprovider.RpcProvider) ProviderResult {
		return CallEVMMethod(ctx, provider, method, params)
	}

	// Use ParallelCheckProviders to execute the calls in parallel
	return ParallelCheckProviders(ctx, providers, timeout, checker)
}
