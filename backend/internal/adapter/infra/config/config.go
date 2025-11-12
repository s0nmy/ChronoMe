package config

import (
	"os"
	"strconv"
	"time"
)

// DefaultSessionSecret is used for local development only.
const DefaultSessionSecret = "dev-secret-change-me"

// Config aggregates runtime configuration loaded from environment variables.
type Config struct {
	Address                string
	DBDriver               string
	DBDsn                  string
	SessionTTLValue        time.Duration
	SessionSecret          string
	SessionCookieSecure    bool
	AllowedOrigin          string
	Environment            string
	DefaultProjectColorHex string
}

// Load returns the configuration with sane defaults for local development.
func Load() Config {
	env := getEnv("APP_ENV", "development")
	cfg := Config{
		Address:                getEnv("SERVER_ADDRESS", ":8080"),
		DBDriver:               getEnv("DB_DRIVER", "sqlite"),
		DBDsn:                  getEnv("DB_DSN", "dev.db"),
		AllowedOrigin:          getEnv("ALLOWED_ORIGIN", "http://localhost:3000"),
		SessionTTLValue:        12 * time.Hour,
		SessionSecret:          getEnv("SESSION_SECRET", DefaultSessionSecret),
		Environment:            env,
		DefaultProjectColorHex: getEnv("DEFAULT_PROJECT_COLOR", "#3B82F6"),
	}
	cfg.SessionCookieSecure = getEnvBool("SESSION_COOKIE_SECURE", env == "production")
	if ttlRaw := os.Getenv("SESSION_TTL"); ttlRaw != "" {
		if parsed, err := time.ParseDuration(ttlRaw); err == nil {
			cfg.SessionTTLValue = parsed
		}
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	if parsed, err := strconv.ParseBool(val); err == nil {
		return parsed
	}
	return fallback
}

// DefaultProjectColor returns the configured default color for new projects.
func (c Config) DefaultProjectColor() string {
	if c.DefaultProjectColorHex == "" {
		return "#3B82F6"
	}
	return c.DefaultProjectColorHex
}

// SessionTTL returns the configured session duration.
func (c Config) SessionTTL() time.Duration {
	return c.SessionTTLValue
}
