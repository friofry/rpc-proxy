// config/config_utils_test.go
package config

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/friofry/config-health-checker/rpcprovider"
)

func TestLoadConfig(t *testing.T) {
	// Create a test configuration file.
	config := Config{
		IntervalSeconds:   10,
		ReferenceProvider: rpcprovider.RpcProvider{},
	}

	// Marshal the configuration to JSON and write it to a file.
	jsonBytes, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile("config.json", jsonBytes, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Load the configuration from the file.
	cfg, err := LoadConfig("config.json")
	if err != nil {
		t.Fatal(err)
	}

	// Check that the loaded configuration matches the original.
	if cfg.IntervalSeconds != config.IntervalSeconds {
		t.Errorf("IntervalSeconds mismatch: expected %d, got %d", config.IntervalSeconds, cfg.IntervalSeconds)
	}
	if cfg.ReferenceProvider != config.ReferenceProvider {
		t.Errorf("ReferenceProvider mismatch: expected %v, got %v", config.ReferenceProvider, cfg.ReferenceProvider)
	}

	// Clean up.
	err = os.Remove("config.json")
	if err != nil {
		t.Fatal(err)
	}
}
