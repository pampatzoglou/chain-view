package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
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

func init() {
	// Register the metrics with Prometheus
	prometheus.MustRegister(FinalizedBlocks)
	prometheus.MustRegister(CurrentBlockHeight)
}
