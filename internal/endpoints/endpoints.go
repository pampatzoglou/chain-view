package endpoints

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pampatzoglou/chain-view/internal/logging"
	"github.com/sirupsen/logrus" // Add this import for logrus
)

// Endpoint represents a single endpoint with its properties.
type Endpoint struct {
	Name    string
	URL     string
	APIKey  string // Optional API key
	Timeout time.Duration
}

// EndpointPool holds the list of endpoints and provides pooling strategies.
type EndpointPool struct {
	Endpoints []Endpoint
	current   int // To keep track of the current endpoint for round robin
}

// RoundRobin returns the next endpoint in a round-robin manner.
func (ep *EndpointPool) RoundRobin() *Endpoint {
	// Return the current endpoint and move to the next one
	endpoint := &ep.Endpoints[ep.current]
	ep.current = (ep.current + 1) % len(ep.Endpoints)
	return endpoint
}

// Fastest sends a request to all endpoints concurrently and returns the first successful response.
func (ep *EndpointPool) Fastest(ctx context.Context, requestBody []byte, logger *logging.Logger) (string, error) {
	type result struct {
		status string
		err    error
	}

	results := make(chan result, len(ep.Endpoints))

	for _, endpoint := range ep.Endpoints {
		go func(ep Endpoint) {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, ep.URL, bytes.NewReader(requestBody))
			if err != nil {
				results <- result{"", err}
				return
			}

			client := &http.Client{Timeout: ep.Timeout}
			resp, err := client.Do(req)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"endpoint": ep.Name,
					"error":    err.Error(),
				}).Warn("Failed to reach endpoint")
				results <- result{"", err}
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				results <- result{resp.Status, nil}
			} else {
				results <- result{"", fmt.Errorf("unexpected status: %s", resp.Status)}
			}
		}(endpoint)
	}

	// Wait for the first successful response or an error
	for i := 0; i < len(ep.Endpoints); i++ {
		res := <-results
		if res.err == nil {
			return res.status, nil
		}
	}

	return "", fmt.Errorf("all endpoints failed")
}

// RetryWithNext attempts to send a request to each endpoint until successful.
func (ep *EndpointPool) RetryWithNext(ctx context.Context, requestBody []byte, logger *logging.Logger) (string, error) {
	for _, endpoint := range ep.Endpoints {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.URL, bytes.NewReader(requestBody))
		if err != nil {
			logger.WithError(err).Error("Failed to create request")
			return "", err
		}
		client := &http.Client{Timeout: endpoint.Timeout}
		resp, err := client.Do(req)
		if err == nil {
			return resp.Status, nil // Successful response
		}
		// Log error with structured logging
		logger.WithFields(logrus.Fields{
			"endpoint": endpoint.Name,
			"error":    err.Error(),
		}).Warn("Failed to reach endpoint")
	}

	return "", fmt.Errorf("all endpoints failed")
}
