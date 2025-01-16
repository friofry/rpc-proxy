package checker

import (
	"context"
	"fmt"
	"time"

	"github.com/friofry/config-health-checker/chainconfig"
	"github.com/friofry/config-health-checker/configreader"
	requestsrunner "github.com/friofry/config-health-checker/requests-runner"
	"github.com/friofry/config-health-checker/rpcprovider"
)

func loadChainsToMap(filePath string) (map[int64]chainconfig.ChainConfig, error) {
	chains, err := chainconfig.LoadChains(filePath)
	if err != nil {
		return nil, err
	}

	chainMap := make(map[int64]chainconfig.ChainConfig)
	for _, chain := range chains {
		chainMap[int64(chain.ChainId)] = chain
	}
	return chainMap, nil
}

func loadReferenceChainsToMap(filePath string) (map[int64]chainconfig.ReferenceChainConfig, error) {
	chains, err := chainconfig.LoadReferenceChains(filePath)
	if err != nil {
		return nil, err
	}

	chainMap := make(map[int64]chainconfig.ReferenceChainConfig)
	for _, chain := range chains {
		chainMap[int64(chain.ChainId)] = chain
	}
	return chainMap, nil
}

// EVMMethodCaller defines the interface for calling EVM methods
type EVMMethodCaller interface {
	CallEVMMethod(
		ctx context.Context,
		provider rpcprovider.RpcProvider,
		method string,
		params []interface{},
		timeout time.Duration,
	) requestsrunner.ProviderResult
}

// ChainValidationRunner coordinates validation across multiple chains
type ChainValidationRunner struct {
	chainConfigs        map[int64]chainconfig.ChainConfig
	referenceChainCfgs  map[int64]chainconfig.ReferenceChainConfig
	methodConfigs       []EVMMethodTestConfig
	caller              EVMMethodCaller
	timeout             time.Duration
	outputProvidersPath string
}

// NewChainValidationRunner creates a new validation runner
func NewChainValidationRunner(
	chainCfgs map[int64]chainconfig.ChainConfig,
	referenceCfgs map[int64]chainconfig.ReferenceChainConfig,
	methodConfigs []EVMMethodTestConfig,
	caller EVMMethodCaller,
	timeout time.Duration,
	outputProvidersPath string,
) *ChainValidationRunner {
	return &ChainValidationRunner{
		chainConfigs:        chainCfgs,
		referenceChainCfgs:  referenceCfgs,
		methodConfigs:       methodConfigs,
		caller:              caller,
		timeout:             timeout,
		outputProvidersPath: outputProvidersPath,
	}
}

// Run executes validation across all configured chains
func (r *ChainValidationRunner) Run(ctx context.Context) map[int64]map[string]ProviderValidationResult {
	results := make(map[int64]map[string]ProviderValidationResult)

	for chainId, chainCfg := range r.chainConfigs {
		// Get reference provider for this chain
		refCfg, exists := r.referenceChainCfgs[chainId]
		if !exists {
			continue
		}

		// Run validation for this chain
		chainResults := ValidateMultipleEVMMethods(
			ctx,
			r.methodConfigs,
			r.caller,
			chainCfg.Providers,
			refCfg.Provider,
			r.timeout,
		)

		results[chainId] = chainResults
	}

	return results
}

// NewRunnerFromConfig creates a new ChainValidationRunner from configreader.CheckerConfig
func NewRunnerFromConfig(
	cfg configreader.CheckerConfig,
	caller EVMMethodCaller,
) (*ChainValidationRunner, error) {
	// Load reference chains
	referenceChains, err := loadReferenceChainsToMap(cfg.ReferenceProvidersPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load reference chains: %w", err)
	}

	// Load default chains
	defaultChains, err := loadChainsToMap(cfg.DefaultProvidersPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load default chains: %w", err)
	}

	return NewChainValidationRunner(
		defaultChains,
		referenceChains,
		nil, // MethodConfigs will need to be implemented separately
		caller,
		time.Duration(cfg.IntervalSeconds)*time.Second,
		cfg.OutputProvidersPath,
	), nil
}

// RunForChain executes validation for a specific chain
func (r *ChainValidationRunner) RunForChain(
	ctx context.Context,
	chainId int64,
) (map[string]ProviderValidationResult, error) {
	chainCfg, exists := r.chainConfigs[chainId]
	if !exists {
		return nil, fmt.Errorf("chain config not found for chainId: %d", chainId)
	}

	refCfg, exists := r.referenceChainCfgs[chainId]
	if !exists {
		return nil, fmt.Errorf("reference config not found for chainId: %d", chainId)
	}

	return ValidateMultipleEVMMethods(
		ctx,
		r.methodConfigs,
		r.caller,
		chainCfg.Providers,
		refCfg.Provider,
		r.timeout,
	), nil
}
