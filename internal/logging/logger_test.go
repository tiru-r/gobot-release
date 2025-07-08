package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(DebugLevel, TextFormat, &buf)
	
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")
	
	output := buf.String()
	
	if !strings.Contains(output, "DEBUG") {
		t.Error("Expected DEBUG level in output")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("Expected INFO level in output")
	}
	if !strings.Contains(output, "WARN") {
		t.Error("Expected WARN level in output")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("Expected ERROR level in output")
	}
}

func TestLogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(WarnLevel, TextFormat, &buf)
	
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")
	
	output := buf.String()
	
	if strings.Contains(output, "DEBUG") {
		t.Error("DEBUG should be filtered out")
	}
	if strings.Contains(output, "INFO") {
		t.Error("INFO should be filtered out")
	}
	if !strings.Contains(output, "WARN") {
		t.Error("Expected WARN level in output")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("Expected ERROR level in output")
	}
}

func TestComponentLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(InfoLevel, TextFormat, &buf)
	componentLogger := logger.WithComponent("test-component")
	
	componentLogger.Info("test message")
	
	output := buf.String()
	if !strings.Contains(output, "[test-component]") {
		t.Error("Expected component name in output")
	}
}

func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(InfoLevel, JSONFormat, &buf)
	
	logger.Info("test message")
	
	output := buf.String()
	if !strings.Contains(output, `"level":"INFO"`) {
		t.Error("Expected JSON formatted level")
	}
	if !strings.Contains(output, `"message":"test message"`) {
		t.Error("Expected JSON formatted message")
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", DebugLevel},
		{"DEBUG", DebugLevel},
		{"info", InfoLevel},
		{"warn", WarnLevel},
		{"warning", WarnLevel},
		{"error", ErrorLevel},
		{"invalid", InfoLevel}, // default
	}
	
	for _, test := range tests {
		result := ParseLogLevel(test.input)
		if result != test.expected {
			t.Errorf("ParseLogLevel(%s) = %v, expected %v", test.input, result, test.expected)
		}
	}
}