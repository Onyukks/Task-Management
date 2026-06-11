// Package middleware contains cross-cutting HTTP middleware.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/tech4mation/tasks-api/internal/auth"
	"github.com/tech4mation/tasks-api/internal/httpx"
)

// CookieName is the auth cookie that carries the JWT.
const CookieName = "tm_token"

type ctxKey string

const principalKey ctxKey = "principal"

// Principal is the authenticated identity attached to a request context.
type Principal struct {
	UserID uuid.UUID
	Role   string
}

// Authenticate verifies the JWT from the auth cookie (or Authorization header
// as a fallback) and attaches the Principal to the request context.
func Authenticate(issuer *auth.Issuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := tokenFromRequest(r)
			if token == "" {
				httpx.Unauthorized(w, "Authentication required.")
				return
			}
			claims, err := issuer.Verify(token)
			if err != nil {
				httpx.Unauthorized(w, "Invalid or expired session.")
				return
			}
			ctx := context.WithValue(r.Context(), principalKey, Principal{
				UserID: claims.UserID,
				Role:   claims.Role,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func tokenFromRequest(r *http.Request) string {
	if c, err := r.Cookie(CookieName); err == nil && c.Value != "" {
		return c.Value
	}
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

// PrincipalFrom returns the authenticated principal from the context.
func PrincipalFrom(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey).(Principal)
	return p, ok
}

// RequireAdmin rejects non-admin principals (used for admin-only routes).
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := PrincipalFrom(r.Context())
		if !ok || p.Role != "admin" {
			httpx.Forbidden(w, "Admin access required.")
			return
		}
		next.ServeHTTP(w, r)
	})
}
