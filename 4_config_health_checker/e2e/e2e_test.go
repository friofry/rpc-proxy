package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/friofry/config-health-checker/checker"
	"github.com/friofry/config-health-checker/confighttpserver"
	"github.com/friofry/config-health-checker/configreader"
	"github.com/friofry/config-health-checker/periodictask"
	requestsrunner "github.com/friofry/config-health-checker/requests-runner"
	"github.com/stretchr/testify/require"
)

const (
	testPort        = "8081"
	testConfigFile  = "test_config.json"
	testOutputFile  = "test_output.json"
	testTempDir     = "testdata"
	shutdownTimeout = 5 * time.Second
)

func TestMain(m *testing.M) {
	// Setup
	err := os.MkdirAll(testTempDir, 0755)
	if err != nil {
		fmt.Printf("failed to create test directory: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	os.RemoveAll(testTempDir)
	os.Exit(code)
}

func TestE2E(t *testing.T) {
	// Create test configuration
	cfg := configreader.CheckerConfig{
		IntervalSeconds:        1,
		DefaultProvidersPath:   filepath.Join(testTempDir, "default_providers.json"),
		ReferenceProvidersPath: filepath.Join(testTempDir, "reference_providers.json"),
		OutputProvidersPath:    filepath.Join(testTempDir, "output_providers.json"),
	}

	// Write config to file
	configPath := filepath.Join(testTempDir, testConfigFile)
	configBytes, err := json.Marshal(cfg)
	require.NoError(t, err)
	err = os.WriteFile(configPath, configBytes, 0644)
	require.NoError(t, err)

	// Create test providers file with sample data
	providersPath := filepath.Join(testTempDir, "providers.json")
	testProviders := map[string]interface{}{
		"test-provider": map[string]interface{}{
			"name":     "Test Provider",
			"url":      "http://localhost:8545",
			"authType": "no-auth",
		},
	}
	providersBytes, err := json.Marshal(testProviders)
	require.NoError(t, err)
	err = os.WriteFile(providersPath, providersBytes, 0644)
	require.NoError(t, err)

	// Start application
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create EVM method caller
	caller := requestsrunner.NewRequestsRunner()

	// Create periodic task
	validationTask := periodictask.New(
		time.Duration(cfg.IntervalSeconds)*time.Second,
		func() {
			runner, err := checker.NewRunnerFromConfig(cfg, caller)
			if err != nil {
				fmt.Printf("failed to create runner: %v\n", err)
				return
			}
			runner.Run(context.Background())
		},
	)

	// Start HTTP server
	server := confighttpserver.New(testPort, providersPath)
	serverDone := make(chan error)
	go func() {
		serverDone <- server.Start()
	}()

	// Start periodic task
	validationTask.Start()
	defer validationTask.Stop()

	// Wait for first run to complete
	time.Sleep(2 * time.Second)

	// Test HTTP endpoint
	t.Run("HTTP API returns providers", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%s/providers", testPort))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var providers map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&providers)
		require.NoError(t, err)
		require.NotEmpty(t, providers)
	})

	// Cleanup server
	cancel()
	select {
	case err := <-serverDone:
		if err != nil {
			t.Logf("server stopped with error: %v", err)
		}
	case <-time.After(shutdownTimeout):
		t.Log("server shutdown timeout")
	}
}
