package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/friofry/config-health-checker/confighttpserver"
	"github.com/friofry/config-health-checker/periodictask"
)

type CheckerConfig struct {
	IntervalSeconds int `json:"interval_seconds"`
}

func main() {
	// Read checker_config.json
	configData, err := os.ReadFile("checker_config.json")
	if err != nil {
		log.Fatalf("failed to read checker_config.json: %v", err)
	}

	var config CheckerConfig
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatalf("failed to unmarshal checker_config.json: %v", err)
	}

	if config.IntervalSeconds <= 0 {
		config.IntervalSeconds = 60
	}

	// Initialize providers.json
	if err := confighttpserver.UpdateProviders(); err != nil {
		log.Printf("initial update providers failed: %v", err)
	}

	// Create periodic task for updating providers
	updateTask := periodictask.New(
		time.Duration(config.IntervalSeconds)*time.Second,
		func() {
			if err := confighttpserver.UpdateProviders(); err != nil {
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

	server := confighttpserver.New(port)
	if err := server.Start(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
