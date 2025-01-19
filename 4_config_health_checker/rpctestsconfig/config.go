package rpctestsconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/friofry/config-health-checker/checker"
)

// EVMMethodTestJSON represents the JSON structure for EVM method test configuration
type EVMMethodTestJSON struct {
	Method        string        `json:"method"`
	Params        []interface{} `json:"params"`
	MaxDifference string        `json:"maxDifference"`
}

// ReadConfig reads and parses the EVM method test configuration from a JSON file
func ReadConfig(path string) ([]checker.EVMMethodTestConfig, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var testConfigs []EVMMethodTestJSON
	if err := json.Unmarshal(data, &testConfigs); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert to EVMMethodTestConfig
	var configs []checker.EVMMethodTestConfig
	for _, cfg := range testConfigs {
		// Parse max difference
		maxDiff, ok := new(big.Int).SetString(cfg.MaxDifference, 10)
		if !ok {
			return nil, fmt.Errorf("invalid maxDifference value: %s", cfg.MaxDifference)
		}

		// Create comparison function
		compareFunc := func(reference, result *big.Int) bool {
			diff := new(big.Int).Abs(new(big.Int).Sub(reference, result))
			return diff.Cmp(maxDiff) <= 0
		}

		configs = append(configs, checker.EVMMethodTestConfig{
			Method:      cfg.Method,
			Params:      cfg.Params,
			CompareFunc: compareFunc,
		})
	}

	return configs, nil
}

// ValidateConfig validates the test configuration
func ValidateConfig(configs []checker.EVMMethodTestConfig) error {
	if len(configs) == 0 {
		return errors.New("empty test configuration")
	}

	for _, cfg := range configs {
		if cfg.Method == "" {
			return errors.New("method name cannot be empty")
		}
	}

	return nil
}
