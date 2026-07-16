package supabase

import (
	"errors"

	"chronome/internal/adapter/infra/config"
)

// Client は Supabase 連携に必要な設定を保持する。
type Client struct {
	URL            string
	AnonKey        string
	ServiceRoleKey string
	JWTSecret      string
}

// NewClient は環境変数由来の Supabase 設定を検証して返す。
func NewClient(cfg config.Config) (*Client, error) {
	if cfg.SupabaseURL == "" {
		return nil, errors.New("SUPABASE_URL is required")
	}
	if cfg.SupabaseJWTSecret == "" {
		return nil, errors.New("SUPABASE_JWT_SECRET is required")
	}
	return &Client{
		URL:            cfg.SupabaseURL,
		AnonKey:        cfg.SupabaseAnonKey,
		ServiceRoleKey: cfg.SupabaseServiceKey,
		JWTSecret:      cfg.SupabaseJWTSecret,
	}, nil
}
