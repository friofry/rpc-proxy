# RPC Proxy Configuration Health Checker

This application validates and monitors RPC provider configurations across multiple chains. Below is a package-by-package explanation of how it works:

## Main Packages

### chainconfig
- Handles loading and managing chain configurations
- Defines ChainConfig and ReferenceChainConfig structs
- Provides methods to load chains from JSON files
- Handles writing validated chain configurations

### checker
- Contains core validation logic
- Implements ChainValidationRunner for coordinating validation
- Validates EVM method responses against reference providers
- Filters and saves valid provider configurations

### confighttpserver
- Manages HTTP server configuration
- Handles API endpoints for valid providers

### configreader
- Reads and parses app configuration files
- Defines CheckerConfig struct for main configuration

### e2e
- Contains end-to-end tests
- Implements test utilities and mocks
- Provides test data and setup helpers
- Validates full application workflow

### periodictask
- Manages periodic execution of tasks
- Handles scheduling and timing

### requests-runner
- Handles parallel RPC requests
- Implements EVMMethodCaller interface
- Manages request timeouts

### rpcprovider
- Defines RPC provider configurations
- Implements provider validation logic

### rpctestsconfig
- Manages configurations for RPC Provider validation
- Defines EVMMethodTestConfig struct
- Handles loading test configurations from files

## Workflow

1. Configuration files are loaded by configreader
2. Chain configurations are loaded by chainconfig
3. PeriodicTask schedules regular health checks
4. Checker coordinates validation across chains
5. Requests-runner executes RPC calls in parallel
6. Results are validated against reference providers
7. Valid configurations are saved by chainconfig
8. Status is exposed via confighttpserver

## Running the Application

```bash
go run main.go --config checker_config.json
```

The application will:
1. Load configurations
2. Start HTTP server
3. Begin periodic health checks
4. Validate RPC providers
5. Save valid configurations
6. Expose status and metrics
