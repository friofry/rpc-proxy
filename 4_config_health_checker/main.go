package main

import (
	"context"
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
	// Read configuration
	config, err := configreader.ReadConfig("checker_config.json")
	if err != nil {
		log.Fatalf("failed to read configuration: %v", err)
	}

	// Create EVM method caller using RequestsRunner
	caller := requestsrunner.NewRequestsRunner()

	// Create runner
	runner, err := checker.NewRunnerFromConfig(*config, caller)
	if err != nil {
		log.Fatalf("failed to create runner: %v", err)
	}

	// Create periodic task for running validation
	validationTask := periodictask.New(
		time.Duration(config.IntervalSeconds)*time.Second,
		func() {
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
