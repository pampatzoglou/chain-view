package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
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
	URL string `yaml:"url"`
}

// LoadConfig loads the configuration from a YAML file
func LoadConfig(filename string) Config {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("Error decoding config file: %v", err)
	}

	return config
}
