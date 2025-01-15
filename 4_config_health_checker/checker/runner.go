package checker

import (
	"context"
	"fmt"
	"time"

	"github.com/friofry/config-health-checker/chainconfig"
	requestsrunner "github.com/friofry/config-health-checker/requests-runner"
	"github.com/friofry/config-health-checker/rpcprovider"
)

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
	chainConfigs       map[int64]chainconfig.ChainConfig
	referenceChainCfgs map[int64]chainconfig.ReferenceChainConfig
	methodConfigs      []EVMMethodTestConfig
	caller             EVMMethodCaller
	timeout            time.Duration
}

// NewChainValidationRunner creates a new validation runner
func NewChainValidationRunner(
	chainCfgs map[int64]chainconfig.ChainConfig,
	referenceCfgs map[int64]chainconfig.ReferenceChainConfig,
	methodConfigs []EVMMethodTestConfig,
	caller EVMMethodCaller,
	timeout time.Duration,
) *ChainValidationRunner {
	return &ChainValidationRunner{
		chainConfigs:       chainCfgs,
		referenceChainCfgs: referenceCfgs,
		methodConfigs:      methodConfigs,
		caller:             caller,
		timeout:            timeout,
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
