package middleware

import (
	"context"
	"net/http"

	"github.com/dhruv15803/url-shortener-service/internal/httpresponse"
	"github.com/dhruv15803/url-shortener-service/internal/token"
)

// SessionCookieName is the cookie the session JWT is stored in. It is the
// single source of truth shared by the handler that sets it and the
// middleware that reads it.
const SessionCookieName = "session"

type contextKey string

const userIDContextKey contextKey = "userID"

// Auth rejects requests without a valid session JWT, and stores the
// authenticated user's id in the request context for downstream handlers.
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil || cookie.Value == "" {
				httpresponse.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			claims, err := token.Parse(jwtSecret, cookie.Value)
			if err != nil {
				httpresponse.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int)
	return userID, ok
}
