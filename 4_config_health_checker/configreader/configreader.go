package configreader

import (
	"encoding/json"
	"os"
)

type CheckerConfig struct {
	IntervalSeconds        int    `json:"interval_seconds"`
	DefaultProvidersPath   string `json:"default_providers_path"`
	ReferenceProvidersPath string `json:"reference_providers_path"`
	OutputProvidersPath    string `json:"output_providers_path"`
}

// ReadConfig reads and validates the configuration from checker_config.json
func ReadConfig(path string) (*CheckerConfig, error) {
	configData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config CheckerConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, err
	}

	// Set default values if not specified or invalid
	if config.IntervalSeconds <= 0 {
		config.IntervalSeconds = 60
	}
	if config.DefaultProvidersPath == "" {
		config.DefaultProvidersPath = "default_providers.json"
	}
	if config.ReferenceProvidersPath == "" {
		config.ReferenceProvidersPath = "reference_providers.json"
	}
	if config.OutputProvidersPath == "" {
		config.OutputProvidersPath = "providers.json"
	}

	return &config, nil
}
