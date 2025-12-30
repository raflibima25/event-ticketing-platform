package config

import (
	"os"
)

// Config holds all application configuration
type Config struct {
	Server ServerConfig
	Resend ResendConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	GRPCPort string
}

// ResendConfig holds Resend email service configuration
type ResendConfig struct {
	APIKey    string
	FromName  string
	FromEmail string
	TestMode  bool
	TestEmail string
}

// Load loads configuration from environment variables
func Load() *Config {
	testMode := getEnv("RESEND_TEST_MODE", "false") == "true"

	return &Config{
		Server: ServerConfig{
			GRPCPort: getEnv("NOTIFICATION_GRPC_PORT", "50055"),
		},
		Resend: ResendConfig{
			APIKey:    getEnv("RESEND_API_KEY", ""),
			FromName:  getEnv("RESEND_FROM_NAME", "Event Ticketing Platform"),
			FromEmail: getEnv("RESEND_FROM_EMAIL", "onboarding@resend.dev"),
			TestMode:  testMode,
			TestEmail: getEnv("RESEND_TEST_EMAIL", ""),
		},
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
