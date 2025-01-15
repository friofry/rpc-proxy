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
	Name      string                    `json:"name" validate:"required,lowercase"`
	Network   string                    `json:"network" validate:"required,lowercase"`
	Providers []rpcprovider.RpcProvider `json:"providers" validate:"required,dive"`
}

// ReferenceChainConfig represents configuration for reference providers
type ReferenceChainConfig struct {
	Name     string                  `json:"name" validate:"required,lowercase"`
	Network  string                  `json:"network" validate:"required,lowercase"`
	Provider rpcprovider.RpcProvider `json:"provider" validate:"required"`
}

// LoadChains loads chain configurations from a JSON file
func LoadChains(filePath string) ([]ChainConfig, error) {
	return loadConfig[ChainConfig](filePath, "chains")
}

// LoadReferenceChains loads reference provider configurations from a JSON file
func LoadReferenceChains(filePath string) ([]ReferenceChainConfig, error) {
	chains, err := loadConfig[ReferenceChainConfig](filePath, "chains")
	if err != nil {
		return nil, err
	}

	for _, chain := range chains {
		if err := validateReferenceChainConfig(chain); err != nil {
			return nil, err
		}
	}

	return chains, nil
}

// validateReferenceChainConfig validates required fields in reference chain configuration
func validateReferenceChainConfig(chain ReferenceChainConfig) error {
	if chain.Name == "" {
		return errors.New("chain name is required")
	}
	if chain.Network == "" {
		return errors.New("network is required")
	}
	if chain.Provider.Name == "" {
		return errors.New("provider name is required")
	}
	if chain.Provider.URL == "" {
		return errors.New("provider URL is required")
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

// loadConfig is a generic function to load chain configurations
func loadConfig[T any](filePath string, key string) ([]T, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config struct {
		Chains []T `json:"chains"`
	}
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if len(config.Chains) == 0 {
		return nil, errors.New("no chains configured")
	}

	// Normalize names and networks to lowercase
	for i := range config.Chains {
		if chain, ok := any(&config.Chains[i]).(interface{ normalize() }); ok {
			chain.normalize()
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

// GetReferenceProvider finds a reference provider by name and network
func GetReferenceProvider(chains []ReferenceChainConfig, name, network string) (*rpcprovider.RpcProvider, error) {
	for _, chain := range chains {
		if chain.Name == name && chain.Network == network {
			return &chain.Provider, nil
		}
	}
	return nil, fmt.Errorf("reference provider for %s (%s) not found", name, network)
}

// normalize ensures chain name and network are lowercase
func (c *ChainConfig) normalize() {
	c.Name = strings.ToLower(c.Name)
	c.Network = strings.ToLower(c.Network)
}

// normalize ensures reference chain name and network are lowercase
func (c *ReferenceChainConfig) normalize() {
	c.Name = strings.ToLower(c.Name)
	c.Network = strings.ToLower(c.Network)
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
