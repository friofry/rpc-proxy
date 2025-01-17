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

	"github.com/friofry/config-health-checker/chainconfig"
	"github.com/friofry/config-health-checker/checker"
	"github.com/friofry/config-health-checker/confighttpserver"
	"github.com/friofry/config-health-checker/configreader"
	"github.com/friofry/config-health-checker/periodictask"
	requestsrunner "github.com/friofry/config-health-checker/requests-runner"
	rpcprovider "github.com/friofry/config-health-checker/rpcprovider"
	"github.com/stretchr/testify/suite"
)

type E2ETestSuite struct {
	suite.Suite
	cfg configreader.CheckerConfig
}

const (
	testPort        = "8081"
	testConfigFile  = "test_config.json"
	testOutputFile  = "test_output.json"
	testTempDir     = "testdata"
	shutdownTimeout = 5 * time.Second
)

func (s *E2ETestSuite) SetupSuite() {
	// Create test directory
	err := os.MkdirAll(testTempDir, 0755)
	if err != nil {
		s.FailNow("failed to create test directory", err)
	}

	// Create test config files
	s.cfg = configreader.CheckerConfig{
		IntervalSeconds:        1,
		DefaultProvidersPath:   filepath.Join(testTempDir, "default_providers.json"),
		ReferenceProvidersPath: filepath.Join(testTempDir, "reference_providers.json"),
		OutputProvidersPath:    filepath.Join(testTempDir, "output_providers.json"),
	}

	// Write default providers using ChainsConfig
	defaultChains := chainconfig.ChainsConfig{
		Chains: []chainconfig.ChainConfig{
			{
				Name:    "test-chain",
				Network: "testnet",
				ChainId: 1,
				Providers: []rpcprovider.RpcProvider{
					{
						Name:     "test-provider",
						URL:      "http://localhost:8545",
						AuthType: "no-auth",
					},
				},
			},
		},
	}
	err = chainconfig.WriteChains(s.cfg.DefaultProvidersPath, defaultChains)
	if err != nil {
		s.FailNow("failed to write default providers", err)
	}

	// Write reference providers using ReferenceChainsConfig
	referenceChains := chainconfig.ReferenceChainsConfig{
		Chains: []chainconfig.ReferenceChainConfig{
			{
				Name:    "test-chain",
				Network: "testnet",
				ChainId: 1,
				Provider: rpcprovider.RpcProvider{
					Name:     "test-provider",
					URL:      "http://localhost:8545",
					AuthType: "no-auth",
				},
			},
		},
	}
	err = chainconfig.WriteReferenceChains(s.cfg.ReferenceProvidersPath, referenceChains)
	if err != nil {
		s.FailNow("failed to write reference providers", err)
	}

	// Write checker config
	s.writeJSONFile(filepath.Join(testTempDir, testConfigFile), s.cfg)
}

func (s *E2ETestSuite) TearDownSuite() {
	os.RemoveAll(testTempDir)
}

func (s *E2ETestSuite) writeJSONFile(path string, data interface{}) {
	bytes, err := json.Marshal(data)
	if err != nil {
		s.FailNow("failed to marshal JSON", err)
	}
	err = os.WriteFile(path, bytes, 0644)
	if err != nil {
		s.FailNow("failed to write file", err, path)
	}
}

func TestE2E(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}

func (s *E2ETestSuite) TestE2E() {
	// Start application
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create EVM method caller
	caller := requestsrunner.NewRequestsRunner()

	// Create periodic task
	validationTask := periodictask.New(
		time.Duration(s.cfg.IntervalSeconds)*time.Second,
		func() {
			runner, err := checker.NewRunnerFromConfig(s.cfg, caller)
			if err != nil {
				fmt.Printf("failed to create runner: %v\n", err)
				return
			}
			runner.Run(context.Background())
		},
	)

	// Start HTTP server
	server := confighttpserver.New(testPort, s.cfg.OutputProvidersPath)
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
	s.Run("HTTP API returns providers", func() {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%s/providers", testPort))
		s.NoError(err)
		defer resp.Body.Close()

		s.Equal(http.StatusOK, resp.StatusCode)

		var providers map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&providers)
		s.NoError(err)
		s.NotEmpty(providers)
	})

	// Cleanup server
	cancel()
	select {
	case err := <-serverDone:
		if err != nil {
			s.T().Logf("server stopped with error: %v", err)
		}
	case <-time.After(shutdownTimeout):
		s.T().Log("server shutdown timeout")
	}
}
