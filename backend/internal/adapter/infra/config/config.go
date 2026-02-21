package config

import (
	"os"
	"strconv"
	"time"
)

// DefaultSessionSecret はローカル開発専用。
const DefaultSessionSecret = "dev-secret-change-me"

// Config は環境変数から読み込む実行時設定をまとめる。
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

// Load はローカル開発向けの妥当なデフォルトを含む設定を返す。
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

// DefaultProjectColor は新規プロジェクト用のデフォルト色を返す。
func (c Config) DefaultProjectColor() string {
	if c.DefaultProjectColorHex == "" {
		return "#3B82F6"
	}
	return c.DefaultProjectColorHex
}

// SessionTTL は設定されたセッション期限を返す。
func (c Config) SessionTTL() time.Duration {
	return c.SessionTTLValue
}
