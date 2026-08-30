package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/roleplay-sim/backend/internal/domain/repository"
	"github.com/roleplay-sim/backend/pkg/jwt"
)

type contextKey string

const (
	ContextKeyUserID    contextKey = "user_id"
	ContextKeySessionID contextKey = "session_id"
	ContextKeyRole      contextKey = "role"
)

func GetUserID(ctx context.Context) string {
	v, _ := ctx.Value(ContextKeyUserID).(string)
	return v
}

func GetRole(ctx context.Context) string {
	v, _ := ctx.Value(ContextKeyRole).(string)
	return v
}

type AuthMiddleware struct {
	jwtSvc       *jwt.Service
	sessionStore repository.SessionStoreRepository
}

func NewAuthMiddleware(jwtSvc *jwt.Service, sessionStore repository.SessionStoreRepository) *AuthMiddleware {
	return &AuthMiddleware{jwtSvc: jwtSvc, sessionStore: sessionStore}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearerToken(r)
		if token == "" {
			writeUnauthorized(w, "missing authorization token")
			return
		}

		claims, err := m.jwtSvc.ValidateAccessToken(token)
		if err != nil {
			writeUnauthorized(w, "invalid or expired token")
			return
		}

		// Check Redis session exists (revocation check)
		_, err = m.sessionStore.Get(r.Context(), claims.SessionID)
		if err != nil {
			writeUnauthorized(w, "session has been revoked")
			return
		}

		ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, ContextKeySessionID, claims.SessionID)
		ctx = context.WithValue(ctx, ContextKeyRole, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := GetRole(r.Context())
		if role != "admin" {
			http.Error(w, `{"error":{"code":"FORBIDDEN","message":"admin access required"}}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(auth, "Bearer ")
}

func writeUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":{"code":"UNAUTHORIZED","message":"` + msg + `"}}`))
}
