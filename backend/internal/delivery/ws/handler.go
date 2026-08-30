package ws

import (
	"net/http"
	"strings"
	"time"

	"github.com/roleplay-sim/backend/internal/domain/repository"
	"github.com/roleplay-sim/backend/pkg/jwt"
	"github.com/rs/zerolog/log"
	"nhooyr.io/websocket"
)

// Handler serves the WebSocket upgrade endpoint.
type Handler struct {
	hub          *Hub
	jwtSvc       *jwt.Service
	sessionStore repository.SessionStoreRepository
}

func NewHandler(hub *Hub, jwtSvc *jwt.Service, sessionStore repository.SessionStoreRepository) *Handler {
	return &Handler{
		hub:          hub,
		jwtSvc:       jwtSvc,
		sessionStore: sessionStore,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract token from query param or Authorization header
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if tokenStr == "" {
		http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"missing token"}}`, http.StatusUnauthorized)
		return
	}

	// Validate JWT
	claims, err := h.jwtSvc.ValidateAccessToken(tokenStr)
	if err != nil {
		http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"invalid or expired token"}}`, http.StatusUnauthorized)
		return
	}

	// Verify session revocation in Redis
	_, err = h.sessionStore.Get(r.Context(), claims.SessionID)
	if err != nil {
		http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"session has been revoked"}}`, http.StatusUnauthorized)
		return
	}

	// Upgrade connection
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		log.Error().Err(err).Msg("ws: failed to accept upgrade")
		return
	}
	conn.SetReadLimit(10 * 1024 * 1024) // 10 MB limit for binary audio streaming chunks

	client := NewClient(claims.UserID, claims.Role, conn, h.hub)

	// Send initial CONNECTED envelope
	client.Send(EventConnected, ConnectedPayload{
		ConnectionID: client.ID,
		ServerTime:   time.Now().UTC().Format(time.RFC3339),
	})

	h.hub.Register(client)

	// Start read & write pumps
	go client.WritePump()
	client.ReadPump() // blocks until client disconnects
}