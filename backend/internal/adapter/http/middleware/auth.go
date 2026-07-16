package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"chronome/internal/domain/entity"
	"chronome/internal/domain/repository"
)

type contextKey string

const userIDKey contextKey = "chronome_user_id"

type supabaseClaims struct {
	Subject      string         `json:"sub"`
	Email        string         `json:"email"`
	Expiry       int64          `json:"exp"`
	Issuer       string         `json:"iss"`
	UserMetadata map[string]any `json:"user_metadata"`
}

// WithSupabaseAuth は Bearer JWT を検証し、Supabase user ID を内部 user ID に変換する。
func WithSupabaseAuth(users repository.UserRepository, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := bearerToken(r.Header.Get("Authorization"))
			if tokenString == "" {
				next.ServeHTTP(w, r)
				return
			}
			claims, err := verifySupabaseJWT(tokenString, jwtSecret)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			supabaseID, err := uuid.Parse(claims.Subject)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			user, err := resolveSupabaseUser(r.Context(), users, supabaseID, claims)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, user.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func resolveSupabaseUser(ctx context.Context, users repository.UserRepository, supabaseID uuid.UUID, claims supabaseClaims) (*entity.User, error) {
	user, err := users.GetBySupabaseID(ctx, supabaseID)
	if err == nil {
		return user, nil
	}
	email := strings.TrimSpace(strings.ToLower(claims.Email))
	if email == "" {
		return nil, errors.New("jwt email is missing")
	}
	user, err = users.GetByEmail(ctx, email)
	if err == nil {
		if user.SupabaseUserID == nil {
			if updateErr := users.UpdateSupabaseID(ctx, user.ID, supabaseID); updateErr != nil {
				return nil, updateErr
			}
			user.SupabaseUserID = &supabaseID
			user.IsMigrated = true
		}
		return user, nil
	}
	displayName := ""
	if rawName, ok := claims.UserMetadata["full_name"].(string); ok {
		displayName = rawName
	} else if rawName, ok := claims.UserMetadata["name"].(string); ok {
		displayName = rawName
	}
	user = &entity.User{
		ID:             uuid.New(),
		Email:          email,
		PasswordHash:   "",
		SupabaseUserID: &supabaseID,
		IsMigrated:     true,
		DisplayName:    displayName,
		TimeZone:       "UTC",
	}
	user.Normalize()
	if err := user.Validate(); err != nil {
		return nil, err
	}
	if err := users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// RequireAuth は認証ユーザーがない場合にリクエストを止める。
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := UserIDFromContext(r.Context()); !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// UserIDFromContext は認証済みユーザー ID を取り出す。
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	if val, ok := ctx.Value(userIDKey).(uuid.UUID); ok {
		return val, true
	}
	return uuid.Nil, false
}

func bearerToken(header string) string {
	prefix := "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

func verifySupabaseJWT(tokenString string, jwtSecret string) (supabaseClaims, error) {
	if jwtSecret == "" {
		return supabaseClaims{}, errors.New("supabase jwt secret is not configured")
	}
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return supabaseClaims{}, errors.New("invalid jwt format")
	}
	header, err := decodeJWTPart(parts[0])
	if err != nil {
		return supabaseClaims{}, err
	}
	var headerPayload struct {
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(header, &headerPayload); err != nil {
		return supabaseClaims{}, err
	}
	if headerPayload.Alg != "HS256" {
		return supabaseClaims{}, errors.New("unexpected jwt signing method")
	}
	mac := hmac.New(sha256.New, []byte(jwtSecret))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	expected := mac.Sum(nil)
	actual, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return supabaseClaims{}, err
	}
	if !hmac.Equal(actual, expected) {
		return supabaseClaims{}, errors.New("invalid jwt signature")
	}
	payload, err := decodeJWTPart(parts[1])
	if err != nil {
		return supabaseClaims{}, err
	}
	var claims supabaseClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return supabaseClaims{}, err
	}
	if claims.Subject == "" {
		return supabaseClaims{}, errors.New("jwt subject is missing")
	}
	if claims.Expiry > 0 && time.Now().Unix() >= claims.Expiry {
		return supabaseClaims{}, errors.New("jwt is expired")
	}
	return claims, nil
}

func decodeJWTPart(part string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(part)
}
