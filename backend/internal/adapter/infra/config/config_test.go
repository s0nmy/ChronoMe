package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigValidateRequiresJWTSecret(t *testing.T) {
	cfg := Config{Environment: "development"}

	require.EqualError(t, cfg.Validate(), "SUPABASE_JWT_SECRET must be provided")
}

func TestConfigValidateRejectsExampleJWTSecretOutsideTests(t *testing.T) {
	cfg := Config{
		Environment:       "development",
		SupabaseJWTSecret: "your-jwt-signing-secret",
	}

	require.EqualError(t, cfg.Validate(), "SUPABASE_JWT_SECRET must be replaced with the Supabase JWT signing secret")
}

func TestConfigValidateAllowsExampleJWTSecretInTests(t *testing.T) {
	cfg := Config{
		Environment:       "test",
		SupabaseJWTSecret: "your-jwt-signing-secret",
	}

	require.NoError(t, cfg.Validate())
}
