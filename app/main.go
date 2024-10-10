package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"

	"github.com/pampatzoglou/chain-view/internal/endpoints"
	"github.com/pampatzoglou/chain-view/internal/logging"
	"github.com/pampatzoglou/chain-view/internal/metrics"
)

// Config represents the configuration structure
type Config struct {
	Server struct {
		Port    int `yaml:"port"`
		Logging struct {
			Level string `yaml:"level"`
		} `yaml:"logging"`
	} `yaml:"server"`
	Database struct {
		URL string `yaml:"url"`
	} `yaml:"database"`
	Redis struct {
		URL string `yaml:"url"`
	} `yaml:"redis"`
	Endpoints struct {
		Chains []endpoints.Endpoint `yaml:"chains"`
	} `yaml:"endpoints"`
}

func main() {
	// Load configuration
	config := loadConfig("config/config.yaml")

	// Initialize the logger with the specified log level
	logger := logging.NewLogger(config.Server.Logging.Level)

	// Initialize the endpoint pool
	ep := endpoints.EndpointPool{
		Endpoints: config.Endpoints.Chains,
	}

	// Start a goroutine to simulate updating metrics
	go func() {
		for {
			// Fetch data from endpoints and update metrics
			updateMetricsFromEndpoints(ep, logger)
			time.Sleep(10 * time.Second) // Adjust the interval as needed
		}
	}()

	// Set up HTTP handlers
	http.HandleFunc("/healthz/health", healthCheckHandler)
	http.HandleFunc("/healthz/ready", readinessCheckHandler)
	http.HandleFunc("/healthz/start", startupCheckHandler)
	http.HandleFunc("/healthz/level", func(w http.ResponseWriter, r *http.Request) {
		handleLogLevelUpdate(w, r, logger)
	})

	// Prometheus metrics endpoint
	http.Handle("/healthz/metrics", promhttp.Handler())

	// Start the HTTP server
	server := &http.Server{Addr: fmt.Sprintf(":%d", config.Server.Port)}

	// Channel for graceful shutdown signals
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info(fmt.Sprintf("Starting server on :%d", config.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(err.Error())
		}
	}()

	// Wait for a shutdown signal
	<-shutdownChan

	logger.Info("Shutting down server...")

	// Set a timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shut down the server gracefully
	if err := server.Shutdown(ctx); err != nil {
		// Change this line to use log.Fatalf
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited gracefully")
}

// Function to load configuration from a YAML file
func loadConfig(filename string) Config {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	return config
}

// Function to fetch data from endpoints and update metrics
func updateMetricsFromEndpoints(ep endpoints.EndpointPool, logger *logging.Logger) {
	for _, endpoint := range ep.Endpoints {
		// Simulate making a request to the endpoint
		logger.Info(fmt.Sprintf("Fetching data from %s", endpoint.Name))
		// Here you would add logic to make an actual request to the endpoint

		// Update metrics with dummy data for demonstration purposes
		metrics.FinalizedBlocks.WithLabelValues("Ethereum", endpoint.Name).Set(1234567)
		metrics.CurrentBlockHeight.WithLabelValues("Ethereum", endpoint.Name).Set(1234568)
	}
}

// Health check handler
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Readiness check handler
func readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Here you could check database connections, etc.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

// Startup check handler
func startupCheckHandler(w http.ResponseWriter, r *http.Request) {
	// You can implement logic to determine if the app is still starting up
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Startup complete"))
}

// Handler for updating log level
func handleLogLevelUpdate(w http.ResponseWriter, r *http.Request, logger *logging.Logger) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Level string `json:"level"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Update log level
	logger.SetLevel(req.Level)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Log level updated to: %s", req.Level)))
}
