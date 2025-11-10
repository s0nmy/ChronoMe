package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

const (
	// CSRFHeaderName is the HTTP header clients must send with mutating requests.
	CSRFHeaderName = "X-CSRF-Token"
	// CSRFCookieName stores the double-submit token accessible to frontend JS.
	CSRFCookieName = "chronome_csrf"
)

// RequireCSRF enforces double-submit CSRF tokens plus optional origin checks.
func RequireCSRF(allowedOrigin string) func(http.Handler) http.Handler {
	allowed := normalizeOrigin(allowedOrigin)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !needsCSRFProtection(r.Method) {
				next.ServeHTTP(w, r)
				return
			}
			if !originAllowed(r.Header.Get("Origin"), allowed) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			cookie, err := r.Cookie(CSRFCookieName)
			if err != nil || cookie.Value == "" {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			header := r.Header.Get(CSRFHeaderName)
			if header == "" {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(header)) != 1 {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func needsCSRFProtection(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}

func originAllowed(originHeader, allowed string) bool {
	if allowed == "" || originHeader == "" {
		return true
	}
	return normalizeOrigin(originHeader) == allowed
}

func normalizeOrigin(origin string) string {
	trimmed := strings.TrimSpace(strings.ToLower(origin))
	return strings.TrimSuffix(trimmed, "/")
}
