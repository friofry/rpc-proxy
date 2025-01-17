package requestsrunner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	rpcprovider "github.com/friofry/config-health-checker/rpcprovider"
)

// RequestsRunner implements EVMMethodCaller interface
type RequestsRunner struct{}

// NewRequestsRunner creates a new instance of RequestsRunner
func NewRequestsRunner() *RequestsRunner {
	return &RequestsRunner{}
}

// CallEVMMethod makes an HTTP POST request to an RPC provider for a specific EVM method
// Implements the EVMMethodCaller interface
func (r *RequestsRunner) CallEVMMethod(
	ctx context.Context,
	provider rpcprovider.RpcProvider,
	method string,
	params []interface{},
	timeout time.Duration,
) ProviderResult {
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
