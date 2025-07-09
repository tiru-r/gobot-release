package config

import (
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Config holds global configuration for Gobot
type Config struct {
	// Logging configuration
	LogLevel    string
	LogFormat   string
	LogOutput   string
	EnableDebug bool

	// Hardware configuration
	GPIOPollInterval   time.Duration
	I2CRetryAttempts   int
	SPIMaxSpeed        int
	SerialTimeout      time.Duration
	ConnectionTimeout  time.Duration

	// API configuration
	APIPort    int
	APIHost    string
	EnableCORS bool
	EnableAuth bool

	// Performance settings
	MaxConcurrentDevices int
	MemoryLimit         int64
}

// Default returns a configuration with sensible defaults
func Default() *Config {
	return &Config{
		// Logging defaults
		LogLevel:    getEnvString("GOBOT_LOG_LEVEL", "info"),
		LogFormat:   getEnvString("GOBOT_LOG_FORMAT", "text"),
		LogOutput:   getEnvString("GOBOT_LOG_OUTPUT", "stdout"),
		EnableDebug: getEnvBool("GOBOT_DEBUG", false),

		// Hardware defaults
		GPIOPollInterval:   getEnvDuration("GOBOT_GPIO_POLL_INTERVAL", 10*time.Millisecond),
		I2CRetryAttempts:   getEnvInt("GOBOT_I2C_RETRY_ATTEMPTS", 3),
		SPIMaxSpeed:        getEnvInt("GOBOT_SPI_MAX_SPEED", 1000000),
		SerialTimeout:      getEnvDuration("GOBOT_SERIAL_TIMEOUT", 5*time.Second),
		ConnectionTimeout:  getEnvDuration("GOBOT_CONNECTION_TIMEOUT", 30*time.Second),

		// API defaults
		APIPort:    getEnvInt("GOBOT_API_PORT", 3000),
		APIHost:    getEnvString("GOBOT_API_HOST", "0.0.0.0"),
		EnableCORS: getEnvBool("GOBOT_API_CORS", true),
		EnableAuth: getEnvBool("GOBOT_API_AUTH", false),

		// Performance defaults
		MaxConcurrentDevices: getEnvInt("GOBOT_MAX_DEVICES", 100),
		MemoryLimit:         getEnvInt64("GOBOT_MEMORY_LIMIT", 512*1024*1024), // 512MB
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.GPIOPollInterval <= 0 {
		return &ConfigError{Field: "GPIOPollInterval", Reason: "must be positive"}
	}
	if c.I2CRetryAttempts < 0 {
		return &ConfigError{Field: "I2CRetryAttempts", Reason: "must be non-negative"}
	}
	if c.SPIMaxSpeed <= 0 {
		return &ConfigError{Field: "SPIMaxSpeed", Reason: "must be positive"}
	}
	if c.SerialTimeout <= 0 {
		return &ConfigError{Field: "SerialTimeout", Reason: "must be positive"}
	}
	if c.ConnectionTimeout <= 0 {
		return &ConfigError{Field: "ConnectionTimeout", Reason: "must be positive"}
	}
	if c.APIPort <= 0 || c.APIPort > 65535 {
		return &ConfigError{Field: "APIPort", Reason: "must be between 1 and 65535"}
	}
	if c.MaxConcurrentDevices <= 0 {
		return &ConfigError{Field: "MaxConcurrentDevices", Reason: "must be positive"}
	}
	if c.MemoryLimit <= 0 {
		return &ConfigError{Field: "MemoryLimit", Reason: "must be positive"}
	}

	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !slices.Contains(validLogLevels, strings.ToLower(c.LogLevel)) {
		return &ConfigError{Field: "LogLevel", Reason: "must be one of: debug, info, warn, error"}
	}

	validLogFormats := []string{"text", "json"}
	if !slices.Contains(validLogFormats, strings.ToLower(c.LogFormat)) {
		return &ConfigError{Field: "LogFormat", Reason: "must be one of: text, json"}
	}

	return nil
}

// ConfigError represents a configuration validation error
type ConfigError struct {
	Field  string
	Reason string
}

func (e *ConfigError) Error() string {
	return "config error in field '" + e.Field + "': " + e.Reason
}

// Helper functions for environment variable parsing
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

