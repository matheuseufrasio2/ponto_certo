package config

import (
	"os"
	"strings"
	"time"
)

// Config holds all configuration variables for the application.
type Config struct {
	Port         string
	Env          string
	DatabaseURL  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Load loads configuration from environment variables and an optional .env file.
func Load() (*Config, error) {
	// Attempt to load local .env file if present
	loadDotEnv(".env")

	cfg := &Config{
		Port:         getEnv("PORT", "8080"),
		Env:          getEnv("ENV", "development"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/ponto_certo?sslmode=disable"),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  time.Minute,
	}

	return cfg, nil
}

// getEnv retrieves the value of an environment variable or returns a fallback default.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

// loadDotEnv parses a simple .env file without external dependencies.
func loadDotEnv(filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return // .env is optional (e.g. in containerized production environments)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, val, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		val = strings.Trim(val, `"'`)

		// Only set if not already defined in the environment
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, val)
		}
	}
}
