// Package config loads and validates all environment variables
// that the server needs at start-up time.
// Every other package should depend on this package, not on os.Getenv directly.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration values.
type Config struct {
	// Port is the TCP port the HTTP server will listen on.
	Port string

	// MongoURI is the full MongoDB connection string.
	// Local:  mongodb://localhost:27017
	// Atlas:  mongodb+srv://user:pass@cluster.mongodb.net/
	MongoURI string

	// DBName is the name of the MongoDB database to use.
	DBName string

	// AdminSecret is the API key that protects admin-only endpoints.
	// Every request to /api/admin/* must supply this value.
	AdminSecret string

	// AllowedOrigins is the list of frontend origins permitted by CORS.
	// Parsed from a comma-separated ALLOWED_ORIGINS env variable.
	AllowedOrigins []string
}

// Load reads the .env file (if present) then reads each variable from the
// process environment.  A missing required variable causes an immediate error.
func Load() (*Config, error) {
	// Silently ignore the error when .env is absent – it is optional in production
	// (env vars should be injected by the hosting platform instead).
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		MongoURI:    getEnv("MONGODB_URI", ""),
		DBName:      getEnv("DB_NAME", "ai_camp"),
		AdminSecret: getEnv("ADMIN_SECRET", ""),
	}

	// Parse comma-separated origins and trim whitespace.
	rawOrigins := getEnv("ALLOWED_ORIGINS", "http://localhost:5173")
	for _, o := range strings.Split(rawOrigins, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			cfg.AllowedOrigins = append(cfg.AllowedOrigins, o)
		}
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks that all required fields are populated.
func (c *Config) validate() error {
	if c.MongoURI == "" {
		return fmt.Errorf("MONGODB_URI environment variable is required")
	}
	if c.AdminSecret == "" {
		return fmt.Errorf("ADMIN_SECRET environment variable is required")
	}
	if c.AdminSecret == "change_me_to_a_strong_secret_key" {
		return fmt.Errorf("ADMIN_SECRET must be changed from the default placeholder value")
	}
	return nil
}

// getEnv returns the value of the environment variable named by key,
// or fallback when the variable is absent or empty.
func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
