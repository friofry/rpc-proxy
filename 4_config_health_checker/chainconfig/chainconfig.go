package chainconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	rpcprovider "github.com/friofry/config-health-checker/rpcprovider"
)

// ChainConfig represents configuration for a blockchain network
type ChainConfig struct {
	Name      string           `json:"name" validate:"required,lowercase"`
	Network   string           `json:"network" validate:"required,lowercase"`
	Providers []ProviderConfig `json:"providers" validate:"required,dive"`
}

// ProviderConfig represents configuration for an RPC provider
type ProviderConfig struct {
	rpcprovider.RpcProvider
	Enabled bool `json:"enabled"`
}

// LoadChains loads chain configurations from a JSON file
func LoadChains(filePath string) ([]ChainConfig, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config struct {
		Chains []ChainConfig `json:"chains"`
	}
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Normalize names and networks to lowercase
	for i := range config.Chains {
		config.Chains[i].Name = strings.ToLower(config.Chains[i].Name)
		config.Chains[i].Network = strings.ToLower(config.Chains[i].Network)
	}

	if len(config.Chains) == 0 {
		return nil, errors.New("no chains configured")
	}

	for i, chain := range config.Chains {
		if err := validateChainConfig(chain); err != nil {
			return nil, fmt.Errorf("invalid chain config at index %d: %w", i, err)
		}
	}

	return config.Chains, nil
}

// GetChainByNameAndNetwork finds a chain by name and network
func GetChainByNameAndNetwork(chains []ChainConfig, name, network string) (*ChainConfig, error) {
	for _, chain := range chains {
		if chain.Name == name && chain.Network == network {
			return &chain, nil
		}
	}
	return nil, fmt.Errorf("chain %s (%s) not found", name, network)
}

// GetEnabledProviders returns only enabled providers for a chain
func (c *ChainConfig) GetEnabledProviders() []ProviderConfig {
	var enabled []ProviderConfig
	for _, provider := range c.Providers {
		if provider.Enabled {
			enabled = append(enabled, provider)
		}
	}
	return enabled
}

// validateChainConfig validates required fields in chain configuration
func validateChainConfig(chain ChainConfig) error {
	if chain.Name == "" {
		return errors.New("chain name is required")
	}
	if chain.Network == "" {
		return errors.New("network is required")
	}
	if len(chain.Providers) == 0 {
		return errors.New("at least one provider is required")
	}

	// Ensure values are lowercase
	if chain.Name != strings.ToLower(chain.Name) {
		return errors.New("chain name must be lowercase")
	}
	if chain.Network != strings.ToLower(chain.Network) {
		return errors.New("network must be lowercase")
	}

	return nil
}
