package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/friofry/config-health-checker/checker"
	"github.com/friofry/config-health-checker/confighttpserver"
	"github.com/friofry/config-health-checker/configreader"
	"github.com/friofry/config-health-checker/periodictask"
	requestsrunner "github.com/friofry/config-health-checker/requests-runner"
)

func main() {
	// Parse command line flags
	checkerConfigPath := flag.String("checker-config", "checker_config.json", "path to checker config")
	defaultProvidersPath := flag.String("default-providers", "", "path to default providers JSON")
	referenceProvidersPath := flag.String("reference-providers", "", "path to reference providers JSON")
	flag.Parse()

	// Read configuration
	config, err := configreader.ReadConfig(*checkerConfigPath)
	if err != nil {
		log.Fatalf("failed to read checker configuration: %v", err)
	}

	// Set provider paths from flags if provided
	if *defaultProvidersPath != "" {
		config.DefaultProvidersPath = *defaultProvidersPath
	}
	if *referenceProvidersPath != "" {
		config.ReferenceProvidersPath = *referenceProvidersPath
	}
	if err != nil {
		log.Fatalf("failed to read configuration: %v", err)
	}

	// Create EVM method caller using RequestsRunner
	caller := requestsrunner.NewRequestsRunner()

	// Create periodic task for running validation
	validationTask := periodictask.New(
		time.Duration(config.IntervalSeconds)*time.Second,
		func() {
			// Create fresh runner for each execution
			// Create a copy of config with updated provider paths
			runnerConfig := *config
			if *defaultProvidersPath != "" {
				runnerConfig.DefaultProvidersPath = *defaultProvidersPath
			}
			if *referenceProvidersPath != "" {
				runnerConfig.ReferenceProvidersPath = *referenceProvidersPath
			}

			runner, err := checker.NewRunnerFromConfig(runnerConfig, caller)
			if err != nil {
				log.Printf("failed to create runner: %v", err)
				return
			}
			runner.Run(context.Background())
		},
	)

	// Start the periodic task
	validationTask.Start()
	defer validationTask.Stop()

	// Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := confighttpserver.New(
		port,
		config.OutputProvidersPath,
	)
	if err := server.Start(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
