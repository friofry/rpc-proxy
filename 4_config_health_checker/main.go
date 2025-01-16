package main

import (
	"log"
	"os"
	"time"

	"github.com/friofry/config-health-checker/confighttpserver"
	"github.com/friofry/config-health-checker/configreader"
	"github.com/friofry/config-health-checker/periodictask"
)

func main() {
	// Read configuration
	config, err := configreader.ReadConfig("checker_config.json")
	if err != nil {
		log.Fatalf("failed to read configuration: %v", err)
	}

	// Initialize providers
	if err := confighttpserver.UpdateProviders(
		config.DefaultProvidersPath,
		config.ReferenceProvidersPath,
		config.OutputProvidersPath,
	); err != nil {
		log.Printf("initial update providers failed: %v", err)
	}

	// Create periodic task for updating providers
	updateTask := periodictask.New(
		time.Duration(config.IntervalSeconds)*time.Second,
		func() {
			if err := confighttpserver.UpdateProviders(
				config.DefaultProvidersPath,
				config.ReferenceProvidersPath,
				config.OutputProvidersPath,
			); err != nil {
				log.Printf("error updating providers: %v", err)
			}
		},
	)

	// Start the periodic task
	updateTask.Start()
	defer updateTask.Stop()

	// Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := confighttpserver.New(
		port,
		config.DefaultProvidersPath,
		config.ReferenceProvidersPath,
		config.OutputProvidersPath,
	)
	if err := server.Start(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
