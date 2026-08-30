package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/roleplay-sim/backend/internal/domain/entity"
)

// --- PostgreSQL Repositories ---

// UserRepository defines persistence operations for users.
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	UpdateEloRating(ctx context.Context, userID uuid.UUID, newRating float64) error
	IncrementSessionCount(ctx context.Context, userID uuid.UUID, won bool) error
	GetLeaderboard(ctx context.Context, limit, offset int) ([]*entity.User, int64, error)
	GetRank(ctx context.Context, userID uuid.UUID) (int64, error)
}

// SessionRepository defines persistence for sessions.
type SessionRepository interface {
	Create(ctx context.Context, session *entity.Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Session, error)
	GetByRoomID(ctx context.Context, roomID string) (*entity.Session, error)
	UpdateState(ctx context.Context, id uuid.UUID, state entity.SessionState) error
	SetStartedAt(ctx context.Context, id uuid.UUID, t time.Time) error
	SetEndedAt(ctx context.Context, id uuid.UUID, t time.Time) error
	AddParticipant(ctx context.Context, p *entity.SessionParticipant) error
	GetParticipants(ctx context.Context, sessionID uuid.UUID) ([]*entity.SessionParticipant, error)
	GetParticipantHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]SessionHistoryRow, error)
}

// SessionHistoryRow is a flattened view for a user's session history.
type SessionHistoryRow struct {
	SessionID            uuid.UUID
	ScenarioTitle        string
	Difficulty           entity.Difficulty
	RolePlayed           string
	OpponentDisplayName  string
	OverallScore         *float64
	ObjectiveAchieved    *bool
	EloDelta             *float64
	PlayedAt             time.Time
}

// ScenarioRepository defines operations for scenario retrieval.
type ScenarioRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Scenario, error)
	GetRandom(ctx context.Context, difficulty entity.Difficulty, contextID *uuid.UUID) (*entity.Scenario, error)
	ListAll(ctx context.Context, difficulty *entity.Difficulty, contextID *uuid.UUID, limit, offset int) ([]*entity.Scenario, int64, error)
	Create(ctx context.Context, s *entity.Scenario) error
	Update(ctx context.Context, s *entity.Scenario) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

// EvaluationRepository defines persistence for AI evaluations.
type EvaluationRepository interface {
	Create(ctx context.Context, eval *entity.Evaluation, scores []*entity.RubricScore) error
	GetBySessionAndParticipant(ctx context.Context, sessionID, participantID uuid.UUID) (*entity.Evaluation, error)
	GetScoresByEvaluation(ctx context.Context, evalID uuid.UUID) ([]*entity.RubricScore, error)
	Exists(ctx context.Context, sessionID, participantID uuid.UUID) (bool, error)
}

// --- Redis Repositories ---

// MatchmakingQueueRepository manages the matchmaking sorted set in Redis.
type MatchmakingQueueRepository interface {
	Enqueue(ctx context.Context, userID string, eloRating float64) error
	Dequeue(ctx context.Context) (userIDA, userIDB string, err error)
	Remove(ctx context.Context, userID string) error
	IsQueued(ctx context.Context, userID string) (bool, error)
	Size(ctx context.Context) (int64, error)
}

// RoomStateRepository manages active room state in Redis.
type RoomStateRepository interface {
	Create(ctx context.Context, room *entity.Room) error
	Get(ctx context.Context, roomID string) (*entity.Room, error)
	UpdateState(ctx context.Context, roomID string, state entity.RoomState) error
	SetPeerConnected(ctx context.Context, roomID, userID string, connected bool) error
	Delete(ctx context.Context, roomID string) error
	Exists(ctx context.Context, roomID string) (bool, error)
}

// SessionStoreRepository manages JWT session tokens in Redis.
type SessionStoreRepository interface {
	Create(ctx context.Context, sessionID, userID string, ttl time.Duration) error
	Get(ctx context.Context, sessionID string) (string, error) // returns userID
	Delete(ctx context.Context, sessionID string) error
	DeleteAllForUser(ctx context.Context, userID string) error
	AddUserSession(ctx context.Context, userID, sessionID string, ttl time.Duration) error
}

// WebRTCCacheRepository caches SDP offers/answers and ICE candidates.
type WebRTCCacheRepository interface {
	StoreSignal(ctx context.Context, roomID, fromUserID string, signal []byte, ttl time.Duration) error
	GetSignals(ctx context.Context, roomID, forUserID string) ([][]byte, error)
	ClearSignals(ctx context.Context, roomID, forUserID string) error
}
