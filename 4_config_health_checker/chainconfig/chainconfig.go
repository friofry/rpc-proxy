package chainconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	rpcprovider "github.com/friofry/config-health-checker/rpcprovider"
	"github.com/go-playground/validator/v10"
)

// ChainsConfig represents a collection of chain configurations
type ChainsConfig struct {
	Chains []ChainConfig `json:"chains" validate:"required,dive"`
}

// ReferenceChainsConfig represents a collection of reference chain configurations
type ReferenceChainsConfig struct {
	Chains []ReferenceChainConfig `json:"chains" validate:"required,dive"`
}

// ChainConfig represents configuration for a blockchain network
type ChainConfig struct {
	Name      string                    `json:"name" validate:"required,lowercase"`
	Network   string                    `json:"network" validate:"required,lowercase"`
	ChainId   int                       `json:"chainId" validate:"required"`
	Providers []rpcprovider.RpcProvider `json:"providers" validate:"required,dive"`
}

// ReferenceChainConfig represents configuration for reference providers
type ReferenceChainConfig struct {
	Name     string                  `json:"name" validate:"required,lowercase"`
	Network  string                  `json:"network" validate:"required,lowercase"`
	ChainId  int                     `json:"chainId" validate:"required"`
	Provider rpcprovider.RpcProvider `json:"provider" validate:"required"`
}

// LoadChains loads chain configurations from a JSON file
func LoadChains(filePath string) (ChainsConfig, error) {
	chains, err := loadConfig[ChainConfig](filePath, "chains")
	return ChainsConfig{Chains: chains}, err
}

// LoadReferenceChains loads reference provider configurations from a JSON file
func LoadReferenceChains(filePath string) (ReferenceChainsConfig, error) {
	chains, err := loadConfig[ReferenceChainConfig](filePath, "chains")
	if err != nil {
		return ReferenceChainsConfig{}, err
	}

	for _, chain := range chains {
		if err := validateReferenceChainConfig(chain); err != nil {
			return ReferenceChainsConfig{}, err
		}
	}

	return ReferenceChainsConfig{Chains: chains}, nil
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

// WriteChains writes chain configurations to a JSON file
func WriteChains(filePath string, config ChainsConfig) error {
	// Validate each chain configuration
	for _, chain := range config.Chains {
		if err := validateChainConfig(chain); err != nil {
			return fmt.Errorf("invalid chain configuration: %w", err)
		}
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal chains: %w", err)
	}

	// Write to file with proper permissions
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write chains file: %w", err)
	}

	return nil
}

// WriteReferenceChains writes reference chain configurations to a JSON file
func WriteReferenceChains(filePath string, config ReferenceChainsConfig) error {
	// Validate each reference chain configuration
	for _, chain := range config.Chains {
		if err := validateReferenceChainConfig(chain); err != nil {
			return fmt.Errorf("invalid reference chain configuration: %w", err)
		}
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal reference chains: %w", err)
	}

	// Write to file with proper permissions
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write reference chains file: %w", err)
	}

	return nil
}

var validate = validator.New()

// validateChainConfig validates required fields in chain configuration
func validateChainConfig(chain ChainConfig) error {
	// Validate struct fields
	if err := validate.Struct(chain); err != nil {
		return err
	}

	// Additional custom validation
	if len(chain.Providers) == 0 {
		return errors.New("at least one provider is required")
	}

	return nil
}
