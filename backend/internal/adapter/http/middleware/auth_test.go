package middleware

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"chronome/internal/domain/entity"
	"chronome/test/fakes"
)

func TestResolveSupabaseUserCreatesUnlinkedUser(t *testing.T) {
	supabaseID := uuid.New()
	var created *entity.User
	users := &fakes.FakeUserRepository{
		GetBySupabaseIDFn: func(context.Context, uuid.UUID) (*entity.User, error) { return nil, errors.New("not found") },
		GetByEmailFn:      func(context.Context, string) (*entity.User, error) { return nil, errors.New("not found") },
		CreateFn: func(_ context.Context, user *entity.User) error {
			created = user
			return nil
		},
	}

	user, err := resolveSupabaseUser(context.Background(), users, supabaseID, supabaseClaims{Email: " New@Example.com "})

	require.NoError(t, err)
	require.Equal(t, "new@example.com", user.Email)
	require.Equal(t, supabaseID, *user.SupabaseUserID)
	require.Same(t, user, created)
}

func TestResolveSupabaseUserRejectsAutomaticMigration(t *testing.T) {
	users := &fakes.FakeUserRepository{
		GetBySupabaseIDFn: func(context.Context, uuid.UUID) (*entity.User, error) { return nil, errors.New("not found") },
		GetByEmailFn: func(context.Context, string) (*entity.User, error) {
			return &entity.User{ID: uuid.New(), Email: "user@example.com", PasswordHash: "legacy"}, nil
		},
	}

	_, err := resolveSupabaseUser(context.Background(), users, uuid.New(), supabaseClaims{Email: "user@example.com"})

	require.EqualError(t, err, "existing account requires explicit Supabase linking")
}

func TestResolveSupabaseUserRejectsDifferentLinkedID(t *testing.T) {
	linkedID := uuid.New()
	users := &fakes.FakeUserRepository{
		GetBySupabaseIDFn: func(context.Context, uuid.UUID) (*entity.User, error) { return nil, errors.New("not found") },
		GetByEmailFn: func(context.Context, string) (*entity.User, error) {
			return &entity.User{ID: uuid.New(), Email: "user@example.com", SupabaseUserID: &linkedID}, nil
		},
	}

	_, err := resolveSupabaseUser(context.Background(), users, uuid.New(), supabaseClaims{Email: "user@example.com"})

	require.EqualError(t, err, "supabase user id does not match the linked account")
}

func TestResolveSupabaseUserTruncatesOAuthDisplayName(t *testing.T) {
	supabaseID := uuid.New()
	users := &fakes.FakeUserRepository{
		GetBySupabaseIDFn: func(context.Context, uuid.UUID) (*entity.User, error) { return nil, errors.New("not found") },
		GetByEmailFn:      func(context.Context, string) (*entity.User, error) { return nil, errors.New("not found") },
	}

	user, err := resolveSupabaseUser(context.Background(), users, supabaseID, supabaseClaims{
		Email:        "user@example.com",
		UserMetadata: map[string]any{"full_name": strings.Repeat("あ", 51)},
	})

	require.NoError(t, err)
	require.LessOrEqual(t, len(user.DisplayName), 50)
	require.Len(t, []rune(user.DisplayName), 16)
}

func TestResolveSupabaseUserStoresMetadataTimeZone(t *testing.T) {
	supabaseID := uuid.New()
	users := &fakes.FakeUserRepository{
		GetBySupabaseIDFn: func(context.Context, uuid.UUID) (*entity.User, error) { return nil, errors.New("not found") },
		GetByEmailFn:      func(context.Context, string) (*entity.User, error) { return nil, errors.New("not found") },
	}

	user, err := resolveSupabaseUser(context.Background(), users, supabaseID, supabaseClaims{
		Email:        "user@example.com",
		UserMetadata: map[string]any{"time_zone": "Asia/Tokyo"},
	})

	require.NoError(t, err)
	require.Equal(t, "Asia/Tokyo", user.TimeZone)
}

func TestResolveSupabaseUserFallsBackToUTCForInvalidMetadataTimeZone(t *testing.T) {
	supabaseID := uuid.New()
	users := &fakes.FakeUserRepository{
		GetBySupabaseIDFn: func(context.Context, uuid.UUID) (*entity.User, error) { return nil, errors.New("not found") },
		GetByEmailFn:      func(context.Context, string) (*entity.User, error) { return nil, errors.New("not found") },
	}

	user, err := resolveSupabaseUser(context.Background(), users, supabaseID, supabaseClaims{
		Email:        "user@example.com",
		UserMetadata: map[string]any{"time_zone": "not/a-time-zone"},
	})

	require.NoError(t, err)
	require.Equal(t, "UTC", user.TimeZone)
}
