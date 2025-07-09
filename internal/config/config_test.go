package config

import (
	"os"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	t.Parallel()
	config := Default()
	
	if config.LogLevel != "info" {
		t.Errorf("Expected default log level 'info', got '%s'", config.LogLevel)
	}
	
	if config.GPIOPollInterval != 10*time.Millisecond {
		t.Errorf("Expected default GPIO poll interval 10ms, got %v", config.GPIOPollInterval)
	}
	
	if config.APIPort != 3000 {
		t.Errorf("Expected default API port 3000, got %d", config.APIPort)
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()
	config := Default()
	
	// Valid config should pass
	if err := config.Validate(); err != nil {
		t.Errorf("Valid config failed validation: %v", err)
	}
	
	// Invalid GPIO poll interval
	config.GPIOPollInterval = -1 * time.Millisecond
	if err := config.Validate(); err == nil {
		t.Error("Expected validation error for negative GPIO poll interval")
	}
	
	// Reset and test invalid log level
	config = Default()
	config.LogLevel = "invalid"
	if err := config.Validate(); err == nil {
		t.Error("Expected validation error for invalid log level")
	}
}

func TestEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("GOBOT_LOG_LEVEL", "debug")
	os.Setenv("GOBOT_API_PORT", "8080")
	os.Setenv("GOBOT_DEBUG", "true")
	defer func() {
		os.Unsetenv("GOBOT_LOG_LEVEL")
		os.Unsetenv("GOBOT_API_PORT")
		os.Unsetenv("GOBOT_DEBUG")
	}()
	
	config := Default()
	
	if config.LogLevel != "debug" {
		t.Errorf("Expected log level from env 'debug', got '%s'", config.LogLevel)
	}
	
	if config.APIPort != 8080 {
		t.Errorf("Expected API port from env 8080, got %d", config.APIPort)
	}
	
	if !config.EnableDebug {
		t.Error("Expected debug enabled from env variable")
	}
}