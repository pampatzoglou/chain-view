package database

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pampatzoglou/chain-view/internal/logging"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	dbQueryCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"query"},
	)

	dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"query"},
	)
)

// DB wraps the SQL database object and logger
type DB struct {
	Conn   *sql.DB
	Logger *logging.Logger
}

// Initialize creates a database connection and registers Prometheus metrics
func Initialize(url string, logger *logging.Logger) (*DB, error) {
	logger.WithFields(logrus.Fields{
		"url": url,
	}).Info("Initializing database connection")

	db, err := sql.Open("pgx", url)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"url":   url,
		}).Error("Error connecting to the database")
		return nil, err
	}

	// Register Prometheus metrics
	prometheus.MustRegister(dbQueryCounter)
	prometheus.MustRegister(dbQueryDuration)

	// Ping the database to ensure the connection is valid
	if err := db.Ping(); err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"url":   url,
		}).Error("Database connection failed")
		return nil, err
	}

	logger.Info("Connected to the database successfully")
	return &DB{Conn: db, Logger: logger}, nil
}

// PerformQuery executes a query and updates Prometheus metrics
func (db *DB) PerformQuery(query string) error {
	db.Logger.WithFields(logrus.Fields{
		"query": query,
	}).Debug("Executing database query")

	start := time.Now()

	// Increment query counter
	dbQueryCounter.WithLabelValues(query).Inc()

	// Execute the query
	_, err := db.Conn.Exec(query)

	// Record query duration
	duration := time.Since(start).Seconds()
	dbQueryDuration.WithLabelValues(query).Observe(duration)

	if err != nil {
		db.Logger.WithFields(logrus.Fields{
			"error":    err,
			"query":    query,
			"duration": duration,
		}).Error("Database query failed")
		return err
	}

	db.Logger.WithFields(logrus.Fields{
		"query":    query,
		"duration": duration,
	}).Debug("Database query executed successfully")

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	db.Logger.Info("Closing database connection")
	return db.Conn.Close()
}
