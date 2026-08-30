package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/roleplay-sim/backend/internal/delivery/http/middleware"
	leaderboardusecase "github.com/roleplay-sim/backend/internal/usecase/leaderboard"
)

type LeaderboardHandler struct {
	leaderboardUC *leaderboardusecase.FetchLeaderboardUseCase
}

func NewLeaderboardHandler(leaderboardUC *leaderboardusecase.FetchLeaderboardUseCase) *LeaderboardHandler {
	return &LeaderboardHandler{leaderboardUC: leaderboardUC}
}

func (h *LeaderboardHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserID(r.Context())
	uid, _ := uuid.Parse(userIDStr) // it's ok if empty/anonymous here

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	res, err := h.leaderboardUC.Execute(r.Context(), uid, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch leaderboard")
		return
	}

	// Format response
	type userRank struct {
		Rank          int64   `json:"rank"`
		UserID        string  `json:"user_id"`
		DisplayName   string  `json:"display_name"`
		AvatarURL     *string `json:"avatar_url"`
		EloRating     float64 `json:"elo_rating"`
		TotalSessions int     `json:"total_sessions"`
		Wins          int     `json:"wins"`
	}

	data := make([]userRank, len(res.Users))
	for i, u := range res.Users {
		data[i] = userRank{
			Rank:          int64(offset + i + 1),
			UserID:        u.ID.String(),
			DisplayName:   u.DisplayName,
			AvatarURL:     u.AvatarURL,
			EloRating:     u.EloRating,
			TotalSessions: u.TotalSessions,
			Wins:          u.Wins,
		}
	}

	totalPages := (res.Total + int64(limit) - 1) / int64(limit)

	writeJSON(w, http.StatusOK, map[string]any{
		"data":    data,
		"my_rank": res.MyRank,
		"my_elo":  res.MyElo,
		"pagination": map[string]any{
			"page":        page,
			"limit":       limit,
			"total":       res.Total,
			"total_pages": totalPages,
		},
	})
}
