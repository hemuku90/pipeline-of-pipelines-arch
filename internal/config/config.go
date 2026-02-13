package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	Host          string `yaml:"host" env:"APP_HOST"`
	Port          int    `yaml:"port" env:"APP_PORT"`
	Environment   string `yaml:"environment" env:"ENVIRONMENT"`
	LogLevel      string `yaml:"log_level" env:"LOG_LEVEL"`
	MetricsPort   int    `yaml:"metrics_port" env:"METRICS_PORT"`
	DatabaseURL   string `yaml:"database_url" env:"DATABASE_URL"`
	RedisURL      string `yaml:"redis_url" env:"REDIS_URL"`
	JWTSecret     string `yaml:"jwt_secret" env:"JWT_SECRET"`
	MaxHeaderSize int    `yaml:"max_header_size" env:"MAX_HEADER_SIZE"`
	ReadTimeout   int    `yaml:"read_timeout" env:"READ_TIMEOUT"`
	WriteTimeout  int    `yaml:"write_timeout" env:"WRITE_TIMEOUT"`
}

// Load reads configuration from environment variables and config file
func Load() (*Config, error) {
	config := &Config{
		Host:          getEnv("APP_HOST", "0.0.0.0"),
		Port:          getEnvAsInt("APP_PORT", 8080),
		Environment:   getEnv("ENVIRONMENT", "development"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		MetricsPort:   getEnvAsInt("METRICS_PORT", 9090),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		RedisURL:      os.Getenv("REDIS_URL"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		MaxHeaderSize: getEnvAsInt("MAX_HEADER_SIZE", 1048576),
		ReadTimeout:   getEnvAsInt("READ_TIMEOUT", 30),
		WriteTimeout:  getEnvAsInt("WRITE_TIMEOUT", 30),
	}

	return config, nil
}

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables
	config.Host = getEnv("APP_HOST", config.Host)
	config.Port = getEnvAsInt("APP_PORT", config.Port)
	config.Environment = getEnv("ENVIRONMENT", config.Environment)
	config.LogLevel = getEnv("LOG_LEVEL", config.LogLevel)
	config.MetricsPort = getEnvAsInt("METRICS_PORT", config.MetricsPort)
	config.DatabaseURL = os.Getenv("DATABASE_URL")
	config.RedisURL = os.Getenv("REDIS_URL")
	config.JWTSecret = os.Getenv("JWT_SECRET")

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}