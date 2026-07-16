package config

import (
	"os"
)

// Config は環境変数から読み込む実行時設定をまとめる。
type Config struct {
	Address                string
	DBDriver               string
	DBDsn                  string
	AllowedOrigin          string
	Environment            string
	DefaultProjectColorHex string
	SupabaseURL            string
	SupabaseAnonKey        string
	SupabaseServiceKey     string
	SupabaseJWTSecret      string
}

// Load はローカル開発向けの妥当なデフォルトを含む設定を返す。
func Load() Config {
	env := getEnv("APP_ENV", "development")
	cfg := Config{
		Address:                getEnv("SERVER_ADDRESS", ":8080"),
		DBDriver:               getEnv("DB_DRIVER", "sqlite"),
		DBDsn:                  getEnv("DB_DSN", "dev.db"),
		AllowedOrigin:          getEnv("ALLOWED_ORIGIN", "http://localhost:3000"),
		Environment:            env,
		DefaultProjectColorHex: getEnv("DEFAULT_PROJECT_COLOR", "#3B82F6"),
		SupabaseURL:            getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey:        getEnv("SUPABASE_ANON_KEY", ""),
		SupabaseServiceKey:     getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		SupabaseJWTSecret:      getEnv("SUPABASE_JWT_SECRET", ""),
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
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
