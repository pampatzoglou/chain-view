package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"database/sql"

	_ "github.com/lib/pq" // PostgreSQL driver
	_ "github.com/lib/pq" // PostgreSQL driver
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

var logger *logging.Logger

var db *sql.DB

func main() {
	// Load configuration
	config := loadConfig("config/config.yaml")

	// Initialize the logger with the specified log level
	logger = logging.NewLogger(config.Server.Logging.Level)

	// Connect to the database
	var err error
	db, err = sql.Open("postgres", config.Database.URL)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to the database")
	}
	defer db.Close()

	// Ping the database to verify the connection
	err = db.Ping()
	if err != nil {
		logger.WithError(err).Fatal("Failed to ping the database")
	}

	// // Initialize the endpoint pool
	// ep := endpoints.EndpointPool{
	// 	Endpoints: config.Endpoints.Chains,
	// }

	// Start a goroutine to simulate updating metrics
	go func() {
		for {
			// Fetch data from endpoints and update metrics
			// updateMetricsFromEndpoints(ep)
			time.Sleep(10 * time.Second) // Adjust the interval as needed
		}
	}()

	// Set up HTTP handlers
	http.HandleFunc("/healthz/health", healthCheckHandler)
	// http.HandleFunc("/healthz/ready", readinessCheckHandler)
	http.HandleFunc("/healthz/start", startupCheckHandler)
	http.HandleFunc("/healthz/level", handleLogLevelUpdate)

	// Prometheus metrics endpoint
	http.Handle("/healthz/metrics", promhttp.Handler())

	// Start the HTTP server
	server := &http.Server{Addr: fmt.Sprintf(":%d", config.Server.Port)}

	// Channel for graceful shutdown signals
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.WithFields(map[string]interface{}{"port": config.Server.Port}).Info("Starting server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Server failed to start")
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
		logger.WithError(err).Fatal("Server forced to shutdown")
	}

	logger.Info("Server exited gracefully")
}

// Function to load configuration from a YAML file
func loadConfig(filename string) Config {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.WithError(err).Fatal("Error reading config file")
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		logger.WithError(err).Fatal("Error parsing config file")
	}

	return config
}

// Function to fetch data from endpoints and update metrics
func updateMetricsFromEndpoints(ep endpoints.EndpointPool) {
	for _, endpoint := range ep.Endpoints {
		// Simulate making a request to the endpoint
		logger.WithFields(map[string]interface{}{"name": endpoint.Name}).Info("Fetching data from endpoint")
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

// // Readiness check handler
func readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the database connection is established
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Database connection not established"))
		return
	}

	// Perform a simple query to check if the database is ready
	_, err := db.Exec("SELECT 1") // Replace with a suitable query for your database
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(fmt.Sprintf("Database not ready: %v", err)))
		return
	}

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
func handleLogLevelUpdate(w http.ResponseWriter, r *http.Request) {
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
	logger.WithFields(map[string]interface{}{"level": req.Level}).Info("Log level updated")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Log level updated to: %s", req.Level)))
}
