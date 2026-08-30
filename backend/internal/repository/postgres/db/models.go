package db

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	UserRolePlayer UserRole = "player"
	UserRoleAdmin  UserRole = "admin"
)

type User struct {
	ID            uuid.UUID  `json:"id"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	PasswordHash  string     `json:"password_hash"`
	DisplayName   string     `json:"display_name"`
	AvatarUrl     *string    `json:"avatar_url"`
	EloRating     float64    `json:"elo_rating"`
	TotalSessions int32      `json:"total_sessions"`
	Wins          int32      `json:"wins"`
	Losses        int32      `json:"losses"`
	Role          string     `json:"role"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type RoleContext struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}

type Role struct {
	ID             uuid.UUID `json:"id"`
	ContextID      uuid.UUID `json:"context_id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	HierarchyLevel int32     `json:"hierarchy_level"`
	Description    *string   `json:"description"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

type Scenario struct {
	ID                     uuid.UUID `json:"id"`
	ContextID              uuid.UUID `json:"context_id"`
	Title                  string    `json:"title"`
	Difficulty             string    `json:"difficulty"`
	BackgroundContext      string    `json:"background_context"`
	RoleAObjective         string    `json:"role_a_objective"`
	RoleAConstraints       []byte    `json:"role_a_constraints"`
	RoleBObjective         string    `json:"role_b_objective"`
	RoleBConstraints       []byte    `json:"role_b_constraints"`
	PrepDurationSeconds    int32     `json:"prep_duration_seconds"`
	SessionDurationSeconds int32     `json:"session_duration_seconds"`
	IsActive               bool      `json:"is_active"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type Session struct {
	ID         uuid.UUID  `json:"id"`
	RoomID     string     `json:"room_id"`
	ScenarioID uuid.UUID  `json:"scenario_id"`
	Difficulty string     `json:"difficulty"`
	State      string     `json:"state"`
	SpinSeed   *int64     `json:"spin_seed"`
	StartedAt  *time.Time `json:"started_at"`
	EndedAt    *time.Time `json:"ended_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type SessionParticipant struct {
	ID        uuid.UUID `json:"id"`
	SessionID uuid.UUID `json:"session_id"`
	UserID    uuid.UUID `json:"user_id"`
	RoleID    uuid.UUID `json:"role_id"`
	Seat      string    `json:"seat"`
	JoinedAt  time.Time `json:"joined_at"`
}

type Evaluation struct {
	ID                            uuid.UUID  `json:"id"`
	SessionID                     uuid.UUID  `json:"session_id"`
	ParticipantID                 uuid.UUID  `json:"participant_id"`
	OverallScore                  *float64   `json:"overall_score"`
	ObjectiveAchieved             *bool      `json:"objective_achieved"`
	ObjectiveAchievementReasoning *string    `json:"objective_achievement_reasoning"`
	SummaryFeedback               *string    `json:"summary_feedback"`
	Strengths                     []byte     `json:"strengths"`
	AreasForImprovement           []byte     `json:"areas_for_improvement"`
	EloDelta                      *float64   `json:"elo_delta"`
	EloRatingBefore               *float64   `json:"elo_rating_before"`
	EloRatingAfter                *float64   `json:"elo_rating_after"`
	LlmModelUsed                  *string    `json:"llm_model_used"`
	SttProvider                   *string    `json:"stt_provider"`
	PromptVersion                 *string    `json:"prompt_version"`
	RawLlmResponse                []byte     `json:"raw_llm_response"`
	IsFallback                    bool       `json:"is_fallback"`
	CreatedAt                     time.Time  `json:"created_at"`
}

type RubricScore struct {
	ID             uuid.UUID `json:"id"`
	EvaluationID   uuid.UUID `json:"evaluation_id"`
	Dimension      string    `json:"dimension"`
	Score          float64   `json:"score"`
	Weight         float64   `json:"weight"`
	Justification  string    `json:"justification"`
	EvidenceQuotes []byte    `json:"evidence_quotes"`
}

type LeaderboardHistory struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	SessionID  uuid.UUID `json:"session_id"`
	EloBefore  float64   `json:"elo_before"`
	EloAfter   float64   `json:"elo_after"`
	EloDelta   float64   `json:"elo_delta"`
	RankBefore *int32    `json:"rank_before"`
	RankAfter  *int32    `json:"rank_after"`
	RecordedAt time.Time `json:"recorded_at"`
}