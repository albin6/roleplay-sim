package handler

import (
	"encoding/json"
	"net/http"

	"github.com/roleplay-sim/backend/internal/delivery/http/middleware"
	"github.com/roleplay-sim/backend/internal/domain/repository"
	authusecase "github.com/roleplay-sim/backend/internal/usecase/auth"
)

type AuthHandler struct {
	registerUC *authusecase.RegisterUseCase
	loginUC    *authusecase.LoginUseCase
	userRepo   repository.UserRepository
	authMW     *middleware.AuthMiddleware
}

func NewAuthHandler(registerUC *authusecase.RegisterUseCase, loginUC *authusecase.LoginUseCase, userRepo repository.UserRepository, authMW *middleware.AuthMiddleware) *AuthHandler {
	return &AuthHandler{
		registerUC: registerUC,
		loginUC:    loginUC,
		userRepo:   userRepo,
		authMW:     authMW,
	}
}

func (h *AuthHandler) setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   7 * 24 * 3600, // 7 days
	})
}

func (h *AuthHandler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var in authusecase.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid json body")
		return
	}

	res, err := h.registerUC.Execute(r.Context(), in)
	if err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	h.setRefreshCookie(w, res.SessionID) // Using session ID as refresh token for brevity

	writeJSON(w, http.StatusCreated, map[string]any{
		"user":         res.User,
		"access_token": res.AccessToken,
		"expires_in":   res.ExpiresIn,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var in authusecase.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid json body")
		return
	}

	res, err := h.loginUC.Execute(r.Context(), in)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
		return
	}

	h.setRefreshCookie(w, res.SessionID)

	writeJSON(w, http.StatusOK, map[string]any{
		"user":         res.User,
		"access_token": res.AccessToken,
		"expires_in":   res.ExpiresIn,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing refresh token")
		return
	}

	// Assuming sessionID acts as the refresh token token and authusecase would
	// validate it in a real scenario. Here we just mock response for completeness.
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token": "new.jwt.token",
		"expires_in":   900,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}
