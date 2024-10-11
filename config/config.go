package config

import (
	"fmt"
	"time"

	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config represents the top-level configuration structure
type Config struct {
	Server         ServerConfig   `yaml:"server"`
	Database       DatabaseConfig `yaml:"database"` // Database configuration at the top level
	Redis          RedisConfig    `yaml:"redis"`
	Chains         []ChainConfig  `yaml:"chains"`
	GlobalSettings GlobalSettings `yaml:"global_settings"`
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Port    int           `yaml:"port"`
	Logging LoggingConfig `yaml:"logging"`
}

// LoggingConfig represents the logging configuration
type LoggingConfig struct {
	Level string `yaml:"level"`
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	URL string `yaml:"url"`
}

// RedisConfig represents the Redis configuration
type RedisConfig struct {
	URL string `yaml:"url"`
}

// ChainConfig represents the configuration for a single chain
type ChainConfig struct {
	ChainID         int              `yaml:"chain_id"`
	Network         string           `yaml:"network"`
	Endpoints       []EndpointConfig `yaml:"endpoints"`
	PoolingStrategy string           `yaml:"pooling_strategy"`
	RetryCount      int              `yaml:"retry_count"`
	RetryBackoff    Duration         `yaml:"retry_backoff"`
}

// EndpointConfig represents a single endpoint configuration
type EndpointConfig struct {
	Name    string   `yaml:"name"`
	URL     string   `yaml:"url"`
	Timeout Duration `yaml:"timeout"`
}

// Duration is a wrapper around time.Duration to handle YAML duration parsing
type Duration struct {
	time.Duration
}

// UnmarshalYAML customizes the unmarshal function for the Duration type
func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = duration
	return nil
}

// GlobalSettings represents the global settings configuration
type GlobalSettings struct {
	RequestTimeout Duration `yaml:"request_timeout"`
	MaxRetries     int      `yaml:"max_retries"`
	MaxWorkers     int      `yaml:"max_workers"`
	RetryBackoff   Duration `yaml:"retry_backoff"`
}

// LoadConfig loads the configuration from the specified YAML file
func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &config, nil
}
