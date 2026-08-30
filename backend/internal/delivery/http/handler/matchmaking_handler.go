package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/roleplay-sim/backend/internal/delivery/http/middleware"
	"github.com/roleplay-sim/backend/internal/domain/repository"
	matchmakingusecase "github.com/roleplay-sim/backend/internal/usecase/matchmaking"
)

type MatchmakingHandler struct {
	enqueueUC *matchmakingusecase.EnqueueUseCase
	dequeueUC *matchmakingusecase.DequeueUseCase
	userRepo  repository.UserRepository
}

func NewMatchmakingHandler(enqueueUC *matchmakingusecase.EnqueueUseCase, dequeueUC *matchmakingusecase.DequeueUseCase, userRepo repository.UserRepository) *MatchmakingHandler {
	return &MatchmakingHandler{
		enqueueUC: enqueueUC,
		dequeueUC: dequeueUC,
		userRepo:  userRepo,
	}
}

type enqueueRequest struct {
	PreferredDifficulty string `json:"preferred_difficulty"`
	PreferredContext    string `json:"preferred_context"`
}

func (h *MatchmakingHandler) Enqueue(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user ID")
		return
	}

	var in enqueueRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid json body")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), uid)
	if err != nil {
		writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
		return
	}

	res, err := h.enqueueUC.Execute(r.Context(), matchmakingusecase.EnqueueInput{
		UserID:     uid,
		EloRating:  user.EloRating,
		Difficulty: in.PreferredDifficulty,
		Context:    in.PreferredContext,
	})
	if err != nil {
		if err.Error() == "already queued" { // Simplified check based on domain error string
			writeError(w, http.StatusBadRequest, "ALREADY_QUEUED", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":                 "queued",
		"position":               res.Position,
		"estimated_wait_seconds": res.Position * 5, // mock calc
	})
}

func (h *MatchmakingHandler) Dequeue(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user ID")
		return
	}

	if err := h.dequeueUC.Execute(r.Context(), uid); err != nil {
		// Ignore if not queued
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MatchmakingHandler) Status(w http.ResponseWriter, r *http.Request) {
	// Simple stub response
	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "queued",
		"position": 1,
		"room_id":  nil,
	})
}
