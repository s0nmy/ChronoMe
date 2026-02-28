package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"chronome/internal/adapter/infra/session"
)

type contextKey string

const userIDKey contextKey = "chronome_user_id"

// WithSession はクッキーが有効な場合に認証ユーザーをコンテキストへ付与する。
func WithSession(store session.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err == nil && cookie.Value != "" {
				if userID, ok := store.Get(cookie.Value); ok {
					ctx := context.WithValue(r.Context(), userIDKey, userID)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuth は WithSession がユーザーを付与していない場合にリクエストを止める。
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

// SessionCookieName はハンドラ間で一貫させるため公開している。
const SessionCookieName = "chronome_session"
