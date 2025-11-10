package session

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSignedCookieStore_RoundTrip(t *testing.T) {
	store, err := NewSignedCookieStore("super-secret")
	require.NoError(t, err)
	fixed := time.Unix(1_700_000_000, 0).UTC()
	store.now = func() time.Time { return fixed }

	userID := uuid.New()
	token, err := store.Create(userID, time.Hour)
	require.NoError(t, err)

	got, ok := store.Get(token)
	require.True(t, ok)
	require.Equal(t, userID, got)

	store.Delete(token)
	_, ok = store.Get(token)
	require.False(t, ok)
}

func TestSignedCookieStore_DetectsTampering(t *testing.T) {
	store, err := NewSignedCookieStore("super-secret")
	require.NoError(t, err)
	token, err := store.Create(uuid.New(), time.Hour)
	require.NoError(t, err)

	raw, err := base64.RawURLEncoding.DecodeString(token)
	require.NoError(t, err)
	mutated := string(raw)
	mutated = strings.Replace(mutated, "|", "/|", 1)
	broken := base64.RawURLEncoding.EncodeToString([]byte(mutated))

	_, ok := store.Get(broken)
	require.False(t, ok)
}

func TestSignedCookieStore_ExpiresTokens(t *testing.T) {
	store, err := NewSignedCookieStore("super-secret")
	require.NoError(t, err)
	start := time.Unix(1_700_000_000, 0).UTC()
	store.now = func() time.Time { return start }
	token, err := store.Create(uuid.New(), time.Hour)
	require.NoError(t, err)

	store.now = func() time.Time { return start.Add(2 * time.Hour) }
	_, ok := store.Get(token)
	require.False(t, ok)
}

func TestSignedCookieStoreRequiresSecret(t *testing.T) {
	_, err := NewSignedCookieStore("  ")
	require.Error(t, err)
}
