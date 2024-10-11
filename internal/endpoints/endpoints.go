package endpoints

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pampatzoglou/chain-view/config"
	"github.com/pampatzoglou/chain-view/internal/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// Endpoint represents a single endpoint with its properties.
type Endpoint struct {
	Name    string
	URL     string
	Timeout time.Duration
}

// EndpointPool holds the list of endpoints and provides pooling strategies.
type EndpointPool struct {
	Endpoints      []Endpoint
	current        int
	RetryCount     int
	RetryBackoff   time.Duration
	JobQueue       chan Job
	circuitBreaker *CircuitBreaker
	rateLimiter    *rate.Limiter

	// Metrics
	jobSuccesses      *prometheus.CounterVec
	jobFailures       *prometheus.CounterVec
	httpResponseCodes *prometheus.CounterVec
	responseDuration  *prometheus.HistogramVec
}

// Job represents a task to be executed by the worker.
type Job struct {
	Endpoint   Endpoint
	Retries    int
	MaxRetries int
	ctx        context.Context
}

// CircuitBreaker represents a simple circuit breaker.
type CircuitBreaker struct {
	state         string
	failures      int
	successes     int
	failureLimit  int
	retryDuration time.Duration
	mu            sync.RWMutex
}

// CircuitBreakerMetrics stores metrics for the circuit breaker.
type CircuitBreakerMetrics struct {
	Failures     int
	Successes    int
	CircuitState string
}

// NewCircuitBreaker creates a new CircuitBreaker.
func NewCircuitBreaker(failureLimit int, retryDuration time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:         "closed",
		failureLimit:  failureLimit,
		retryDuration: retryDuration,
	}
}

// Allow checks if the circuit breaker allows a request.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state != "open"
}

// RecordFailure records a failure and updates the state.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	if cb.failures >= cb.failureLimit {
		cb.state = "open"
		go func() {
			time.Sleep(cb.retryDuration)
			cb.mu.Lock()
			cb.state = "half-open"
			cb.mu.Unlock()
		}()
	}
}

// RecordSuccess records a successful request and resets failures.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successes++
	cb.failures = 0
	cb.state = "closed"
}

// GetMetrics returns the current metrics of the circuit breaker.
func (cb *CircuitBreaker) GetMetrics() CircuitBreakerMetrics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return CircuitBreakerMetrics{
		Failures:     cb.failures,
		Successes:    cb.successes,
		CircuitState: cb.state,
	}
}

// CreatePools creates endpoint pools for each chain specified in the configuration.
func CreatePools(chains []config.ChainConfig, logger *logging.Logger) ([]*EndpointPool, []error) {
	var pools []*EndpointPool
	var errors []error

	for _, chain := range chains {
		pool, err := NewEndpointPool(chain, logger)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to create endpoint pool for chain %s (ID: %d): %w", chain.Network, chain.ChainID, err))
			continue
		}
		pools = append(pools, pool)

		go pool.LogChainConfig(logger)
	}

	return pools, errors
}

// NewEndpointPool initializes a new EndpointPool with the given chain configuration and logger.
func NewEndpointPool(chain config.ChainConfig, logger *logging.Logger) (*EndpointPool, error) {
	if err := validateChainConfig(chain); err != nil {
		return nil, err
	}

	endpoints := make([]Endpoint, len(chain.Endpoints))
	for i, ep := range chain.Endpoints {
		endpoints[i] = Endpoint{
			Name:    ep.Name,
			URL:     ep.URL,
			Timeout: ep.Timeout.Duration,
		}
	}

	logger.WithFields(logrus.Fields{
		"network":   chain.Network,
		"chain_id":  chain.ChainID,
		"endpoints": chain.Endpoints,
	}).Info("Initialized endpoint pool")

	// Initialize metrics
	jobSuccesses := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "chainview_job_successes_total",
		Help: "Total number of successful jobs",
	}, []string{"chain", "endpoint"})
	jobFailures := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "chainview_job_failures_total",
		Help: "Total number of failed jobs",
	}, []string{"chain", "endpoint"})
	httpResponseCodes := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "chainview_http_response_codes_total",
		Help: "Total number of HTTP response codes by endpoint",
	}, []string{"chain", "endpoint", "code"})
	responseDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "chainview_response_duration_seconds",
		Help:    "Histogram of response durations by endpoint",
		Buckets: prometheus.DefBuckets,
	}, []string{"chain", "endpoint"})

	// Register metrics with Prometheus
	prometheus.MustRegister(jobSuccesses, jobFailures, httpResponseCodes, responseDuration)

	return &EndpointPool{
		Endpoints:         endpoints,
		current:           0,
		RetryCount:        chain.RetryCount,
		RetryBackoff:      chain.RetryBackoff.Duration,
		circuitBreaker:    NewCircuitBreaker(3, 10*time.Second),
		rateLimiter:       rate.NewLimiter(rate.Every(time.Second), 10), // 10 requests per second
		jobSuccesses:      jobSuccesses,
		jobFailures:       jobFailures,
		httpResponseCodes: httpResponseCodes,
		responseDuration:  responseDuration,
	}, nil
}

func validateChainConfig(chain config.ChainConfig) error {
	if chain.Network == "" {
		return fmt.Errorf("network name is required")
	}
	if chain.ChainID == 0 {
		return fmt.Errorf("chain ID is required")
	}
	if len(chain.Endpoints) == 0 {
		return fmt.Errorf("at least one endpoint is required")
	}
	for _, ep := range chain.Endpoints {
		if ep.Name == "" || ep.URL == "" {
			return fmt.Errorf("endpoint name and URL are required")
		}
	}
	return nil
}

// LogChainConfig logs the configuration for the chain.
func (ep *EndpointPool) LogChainConfig(logger *logging.Logger) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		logger.WithFields(logrus.Fields{
			"endpoints": ep.Endpoints,
		}).Debug("Logging chain configuration")
	}
}

// Worker processes jobs from the job channel.
func (ep *EndpointPool) worker(id int, jobs <-chan Job, wg *sync.WaitGroup, logger *logging.Logger) {
	defer wg.Done()
	for job := range jobs {
		if err := ep.rateLimiter.Wait(job.ctx); err != nil {
			logger.WithError(err).Warn("Rate limit exceeded, skipping job")
			continue
		}

		if !ep.circuitBreaker.Allow() {
			logger.Warn("Circuit breaker is open, skipping job")
			continue
		}

		start := time.Now()
		err := ep.fetchData(job.ctx, job.Endpoint, logger)
		duration := time.Since(start).Seconds()

		ep.responseDuration.WithLabelValues(job.Endpoint.Name, job.Endpoint.URL).Observe(duration)

		if err != nil {
			ep.jobFailures.WithLabelValues(job.Endpoint.Name, job.Endpoint.URL).Inc()
			ep.circuitBreaker.RecordFailure()
			logger.WithError(err).Errorf("Worker %d failed to fetch data from %s", id, job.Endpoint.URL)
			if job.Retries < job.MaxRetries {
				job.Retries++
				backoff := time.Duration(job.Retries*job.Retries) * time.Second // Exponential backoff
				logger.Infof("Retrying job for %s (%d/%d) after %s", job.Endpoint.URL, job.Retries, job.MaxRetries, backoff)
				time.Sleep(backoff)
				ep.JobQueue <- job
			}
		} else {
			ep.jobSuccesses.WithLabelValues(job.Endpoint.Name, job.Endpoint.URL).Inc()
			ep.circuitBreaker.RecordSuccess()
			logger.Infof("Worker %d successfully processed job for %s", id, job.Endpoint.URL)
		}
	}
}

// StartWorkers initializes the worker pool.
func (ep *EndpointPool) StartWorkers(ctx context.Context, numWorkers int, logger *logging.Logger) {
	var wg sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go ep.worker(w, ep.JobQueue, &wg, logger)
	}

	<-ctx.Done()
	close(ep.JobQueue)
	wg.Wait()
}

// FetchData fetches data from an endpoint and handles timeouts.
func (ep *EndpointPool) fetchData(ctx context.Context, endpoint Endpoint, logger *logging.Logger) error {
	client := &http.Client{
		Timeout: endpoint.Timeout,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	ep.httpResponseCodes.WithLabelValues(endpoint.Name, endpoint.URL, fmt.Sprintf("%d", resp.StatusCode)).Inc()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}
	return nil
}

// GetNextEndpoint returns the next endpoint using a round-robin strategy.
func (ep *EndpointPool) GetNextEndpoint() Endpoint {
	ep.current = (ep.current + 1) % len(ep.Endpoints)
	return ep.Endpoints[ep.current]
}

// ProcessEndpoints starts the processing of endpoints concurrently.
func (ep *EndpointPool) ProcessEndpoints(ctx context.Context, numWorkers int, logger *logging.Logger) {
	ep.JobQueue = make(chan Job, len(ep.Endpoints))

	go ep.StartWorkers(ctx, numWorkers, logger)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping endpoint processing due to context cancellation")
			return
		case <-ticker.C:
			endpoint := ep.GetNextEndpoint()
			select {
			case ep.JobQueue <- Job{Endpoint: endpoint, Retries: 0, MaxRetries: ep.RetryCount, ctx: ctx}:
			default:
				logger.Warn("Job queue is full, skipping job")
			}
		}
	}
}

// LogCircuitBreakerMetrics logs the circuit breaker metrics periodically.
func (ep *EndpointPool) LogCircuitBreakerMetrics(ctx context.Context, logger *logging.Logger) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics := ep.circuitBreaker.GetMetrics()
			logger.WithFields(logrus.Fields{
				"failures":      metrics.Failures,
				"successes":     metrics.Successes,
				"circuit_state": metrics.CircuitState,
			}).Info("Circuit Breaker Metrics")
		}
	}
}

// ReloadConfig allows for dynamic configuration updates.
func (ep *EndpointPool) ReloadConfig(newConfig config.ChainConfig) error {
	if err := validateChainConfig(newConfig); err != nil {
		return err
	}

	ep.RetryCount = newConfig.RetryCount
	ep.RetryBackoff = newConfig.RetryBackoff.Duration

	newEndpoints := make([]Endpoint, len(newConfig.Endpoints))
	for i, epConfig := range newConfig.Endpoints {
		newEndpoints[i] = Endpoint{
			Name:    epConfig.Name,
			URL:     epConfig.URL,
			Timeout: epConfig.Timeout.Duration,
		}
	}

	ep.Endpoints = newEndpoints
	ep.current = 0 // Reset the current endpoint index

	return nil
}
