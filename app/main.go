package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"database/sql"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/pampatzoglou/chain-view/config"
	"github.com/pampatzoglou/chain-view/internal/endpoints"
	"github.com/pampatzoglou/chain-view/internal/logging"
)

var logger *logging.Logger
var db *sql.DB
var redisClient *redis.Client

func main() {
	// Load the configuration from the YAML file
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		fmt.Printf("Could not load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger after loading configuration
	logger = logging.NewLogger(cfg.Server.Logging.Level)

	// Database connection setup
	// db, err = sql.Open("postgres", cfg.Database.URL)
	// if err != nil {
	// 	logger.WithError(err).Fatal("Failed to connect to the database")
	// }
	// defer db.Close()

	// err = db.Ping()
	// if err != nil {
	// 	logger.WithError(err).Fatal("Failed to ping the database")
	// }

	// Redis connection setup
	// redisClient = redis.NewClient(&redis.Options{
	// 	Addr: cfg.Redis.URL,
	// })
	// defer redisClient.Close()

	// _, err = redisClient.Ping(context.Background()).Result()
	// if err != nil {
	// 	logger.WithError(err).Fatal("Failed to connect to Redis")
	// }

	// Create endpoint pools based on the loaded configuration
	pools, poolErrors := endpoints.CreatePools(cfg.Chains, logger)
	if len(poolErrors) > 0 {
		for _, err := range poolErrors {
			logger.WithError(err).Error("Error creating endpoint pool")
		}
		if len(pools) == 0 {
			logger.Fatal("Failed to create any endpoint pools")
		}
	}

	// Create a cancelable context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start processing endpoints for each pool
	var wg sync.WaitGroup
	for _, pool := range pools {
		wg.Add(1)
		go func(p *endpoints.EndpointPool) {
			defer wg.Done()
			p.ProcessEndpoints(ctx, cfg.GlobalSettings.MaxWorkers, logger)
		}(pool)

		// Start logging circuit breaker metrics
		go pool.LogCircuitBreakerMetrics(ctx, logger)
	}

	// Set up HTTP handlers
	http.HandleFunc("/healthz/health", healthCheckHandler)
	http.HandleFunc("/healthz/start", startupCheckHandler)
	http.HandleFunc("/healthz/level", handleLogLevelUpdate)
	http.Handle("/healthz/metrics", promhttp.Handler())

	// Start the HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		ReadTimeout:  cfg.GlobalSettings.RequestTimeout.Duration,
		WriteTimeout: cfg.GlobalSettings.RequestTimeout.Duration,
	}

	// Channel for graceful shutdown signals
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.WithFields(logrus.Fields{"port": cfg.Server.Port}).Info("Starting server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Server failed to start")
		}
	}()

	// Wait for a shutdown signal
	<-shutdownChan
	logger.Info("Shutting down server...")

	// Cancel the context to stop all ongoing operations
	cancel()

	// Set a timeout for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	// Shut down the server gracefully
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Fatal("Server forced to shutdown")
	}

	// Wait for all endpoint processing to complete
	wg.Wait()

	logger.Info("Server exited gracefully")
}

// Health check handler
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Startup check handler
func startupCheckHandler(w http.ResponseWriter, r *http.Request) {
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

	// Validate the log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLevels[req.Level] {
		http.Error(w, fmt.Sprintf("Invalid log level: %s", req.Level), http.StatusBadRequest)
		return
	}

	// Update log level
	logger.SetLevel(req.Level)
	logger.WithFields(logrus.Fields{"level": req.Level}).Info("Log level updated")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Log level updated to: %s", req.Level)))
}
