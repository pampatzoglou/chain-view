package config

import (
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/pampatzoglou/chain-view/internal/logging"
)

// Config structure to hold application configuration
type Config struct {
	Server    ServerConfig   `yaml:"server"`
	Database  DatabaseConfig `yaml:"database"`
	Redis     RedisConfig    `yaml:"redis"`
	Endpoints EndpointConfig `yaml:"endpoints"`
}

// ServerConfig structure to hold server-specific settings
type ServerConfig struct {
	Port    int           `yaml:"port"`
	Logging LoggingConfig `yaml:"logging"`
}

// LoggingConfig structure to hold logging settings
type LoggingConfig struct {
	Level string `yaml:"level"` // Options: debug, info, warn, error
}

// DatabaseConfig structure to hold database settings
type DatabaseConfig struct {
	URL string `yaml:"url"`
}

// RedisConfig structure to hold Redis settings
type RedisConfig struct {
	URL string `yaml:"url"`
}

// EndpointConfig structure to hold endpoint settings
type EndpointConfig struct {
	Chains []Endpoint `yaml:"chains"`
}

// Endpoint structure to hold individual endpoint details
type Endpoint struct {
	URL  string `yaml:"url"`
	Name string `yaml:"name"`
}

// LoadConfig loads the configuration from a YAML file
func LoadConfig(filename string, logger *logging.Logger) (*Config, error) {
	logger.WithFields(logrus.Fields{
		"filename": filename,
	}).Info("Loading configuration file")

	file, err := os.Open(filename)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"file":  filename,
		}).Error("Failed to open config file")
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
			"file":  filename,
		}).Error("Failed to decode config file")
		return nil, err
	}

	logger.WithFields(logrus.Fields{
		"server_port": config.Server.Port,
		"log_level":   config.Server.Logging.Level,
	}).Info("Configuration loaded successfully")

	return &config, nil
}
