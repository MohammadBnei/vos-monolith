package logger

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestLoggerOutputFormat(t *testing.T) {
	// Save original ENV value and restore it after the test
	originalEnv := os.Getenv("ENV")
	defer os.Setenv("ENV", originalEnv)

	tests := []struct {
		name        string
		environment string
		wantJSON    bool
	}{
		{
			name:        "Development environment should use pretty format",
			environment: EnvDevelopment,
			wantJSON:    false,
		},
		{
			name:        "Production environment should use JSON format",
			environment: EnvProduction,
			wantJSON:    true,
		},
		{
			name:        "Test environment should use pretty format",
			environment: EnvTest,
			wantJSON:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture log output
			var buf bytes.Buffer

			// Create logger with the test environment
			config := Config{
				Environment: tt.environment,
				Level:       LevelDebug,
				Output:      &buf,
			}
			log := NewWithConfig(config)

			// Write a test log message
			log.Info().Str("test", "value").Msg("Test message")

			// Get the output
			output := buf.String()

			// Check if the output is JSON
			isJSON := isJSONFormat(output)

			if tt.wantJSON != isJSON {
				t.Errorf("Logger output format = %v, want JSON = %v, got output: %s", 
					isJSON, tt.wantJSON, output)
			}
		})
	}
}

// TestEnvironmentFromEnvVar tests that the ENV environment variable is respected
func TestEnvironmentFromEnvVar(t *testing.T) {
	// Save original ENV value and restore it after the test
	originalEnv := os.Getenv("ENV")
	defer os.Setenv("ENV", originalEnv)

	tests := []struct {
		name     string
		envValue string
		wantJSON bool
	}{
		{
			name:     "ENV=production should use JSON format",
			envValue: "production",
			wantJSON: true,
		},
		{
			name:     "ENV=development should use pretty format",
			envValue: "development",
			wantJSON: false,
		},
		{
			name:     "Empty ENV should default to pretty format",
			envValue: "",
			wantJSON: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the environment variable
			os.Setenv("ENV", tt.envValue)

			// Create a buffer to capture log output
			var buf bytes.Buffer

			// Create logger config
			config := DefaultConfig()
			config.Output = &buf

			// Set environment based on ENV var
			if os.Getenv("ENV") == "production" {
				config.Environment = EnvProduction
			}

			// Create logger
			log := NewWithConfig(config)

			// Write a test log message
			log.Info().Str("test", "value").Msg("Test message")

			// Get the output
			output := buf.String()

			// Check if the output is JSON
			isJSON := isJSONFormat(output)

			if tt.wantJSON != isJSON {
				t.Errorf("Logger output format = %v, want JSON = %v, got output: %s", 
					isJSON, tt.wantJSON, output)
			}
		})
	}
}

// isJSONFormat checks if a string is in JSON format
func isJSONFormat(s string) bool {
	// If it contains color codes or formatted output, it's not JSON
	if strings.Contains(s, "\u001b[") || strings.Contains(s, "INF") {
		return false
	}

	// Try to parse as JSON
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
