package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/pampatzoglou/chain-view/internal/logging"
)

// Metrics for tracking finalized blocks and current block heights
var (
	FinalizedBlocks = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "finalized_blocks",
			Help: "Number of finalized blocks per chain and provider",
		},
		[]string{"chain", "provider"},
	)

	CurrentBlockHeight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "current_block_height",
			Help: "Current block height per chain and provider",
		},
		[]string{"chain", "provider"},
	)
)

// MetricsManager handles the registration and updating of metrics
type MetricsManager struct {
	logger *logging.Logger
}

// NewMetricsManager creates a new MetricsManager
func NewMetricsManager(logger *logging.Logger) *MetricsManager {
	return &MetricsManager{
		logger: logger,
	}
}

// RegisterMetrics registers all metrics with Prometheus
func (mm *MetricsManager) RegisterMetrics() {
	mm.logger.Info("Registering Prometheus metrics")

	prometheus.MustRegister(FinalizedBlocks)
	prometheus.MustRegister(CurrentBlockHeight)

	mm.logger.Info("Prometheus metrics registered successfully")
}

// UpdateMetrics updates the metrics with new values
func (mm *MetricsManager) UpdateMetrics(chain, provider string, finalizedBlocks, currentHeight float64) {
	mm.logger.WithFields(logrus.Fields{
		"chain":           chain,
		"provider":        provider,
		"finalizedBlocks": finalizedBlocks,
		"currentHeight":   currentHeight,
	}).Debug("Updating metrics")

	FinalizedBlocks.WithLabelValues(chain, provider).Set(finalizedBlocks)
	CurrentBlockHeight.WithLabelValues(chain, provider).Set(currentHeight)
}
