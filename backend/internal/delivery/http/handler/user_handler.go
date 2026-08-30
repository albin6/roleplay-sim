package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/roleplay-sim/backend/internal/delivery/http/middleware"
	"github.com/roleplay-sim/backend/internal/domain/repository"
)

type UserHandler struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
}

func NewUserHandler(userRepo repository.UserRepository, sessionRepo repository.SessionRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo, sessionRepo: sessionRepo}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user ID")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), uid)
	if err != nil {
		writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

type updateMeRequest struct {
	DisplayName *string `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user ID")
		return
	}

	var in updateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid json body")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), uid)
	if err != nil {
		writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
		return
	}

	if in.DisplayName != nil {
		user.DisplayName = *in.DisplayName
	}
	if in.AvatarURL != nil {
		user.AvatarURL = in.AvatarURL
	}

	if err := h.userRepo.Update(r.Context(), user); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update user")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *UserHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user ID")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	offset := (page - 1) * limit
	history, err := h.sessionRepo.GetParticipantHistory(r.Context(), uid, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch history")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": history,
		"pagination": map[string]any{
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *UserHandler) GetHistoryDetail(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionID")
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "missing sessionID")
		return
	}

	// In real app, fetch session details from repo
	writeJSON(w, http.StatusOK, map[string]any{
		"session_id": sessionID,
		"scenario": map[string]any{
			"title":              "Stubbed Scenario",
			"difficulty":         "medium",
			"background_context": "...",
		},
		"evaluation": map[string]any{
			"overall_score":      0.0,
			"objective_achieved": false,
		},
	})
}
