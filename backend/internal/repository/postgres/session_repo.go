package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roleplay-sim/backend/internal/domain/entity"
	domainerrors "github.com/roleplay-sim/backend/internal/domain/errors"
	"github.com/roleplay-sim/backend/internal/domain/repository"
	"github.com/roleplay-sim/backend/internal/repository/postgres/db"
)

type SessionRepo struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewSessionRepo(pool *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{pool: pool, queries: db.New(pool)}
}

func (r *SessionRepo) Create(ctx context.Context, session *entity.Session) error {
	row, err := r.queries.CreateSession(ctx, db.CreateSessionParams{
		RoomID:     session.RoomID,
		ScenarioID: session.ScenarioID,
		Difficulty: string(session.Difficulty),
		State:      string(session.State),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domainerrors.ErrRoomNotFound
		}
		return fmt.Errorf("session_repo: create: %w", err)
	}
	session.ID = row.ID
	session.CreatedAt = row.CreatedAt
	session.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *SessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Session, error) {
	row, err := r.queries.GetSessionByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrSessionNotFound
		}
		return nil, fmt.Errorf("session_repo: get by id: %w", err)
	}
	return dbSessionToEntity(row), nil
}

func (r *SessionRepo) GetByRoomID(ctx context.Context, roomID string) (*entity.Session, error) {
	row, err := r.queries.GetSessionByRoomID(ctx, roomID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrSessionNotFound
		}
		return nil, fmt.Errorf("session_repo: get by room id: %w", err)
	}
	return dbSessionToEntity(row), nil
}

func (r *SessionRepo) UpdateState(ctx context.Context, id uuid.UUID, state entity.SessionState) error {
	err := r.queries.UpdateSessionState(ctx, db.UpdateSessionStateParams{
		ID:    id,
		State: string(state),
	})
	if err != nil {
		return fmt.Errorf("session_repo: update state: %w", err)
	}
	return nil
}

func (r *SessionRepo) SetStartedAt(ctx context.Context, id uuid.UUID, t time.Time) error {
	err := r.queries.SetSessionStartedAt(ctx, db.SetSessionStartedAtParams{
		ID:        id,
		StartedAt: &t,
	})
	if err != nil {
		return fmt.Errorf("session_repo: set started at: %w", err)
	}
	return nil
}

func (r *SessionRepo) SetEndedAt(ctx context.Context, id uuid.UUID, t time.Time) error {
	err := r.queries.SetSessionEndedAt(ctx, db.SetSessionEndedAtParams{
		ID:      id,
		EndedAt: &t,
	})
	if err != nil {
		return fmt.Errorf("session_repo: set ended at: %w", err)
	}
	return nil
}

func (r *SessionRepo) AddParticipant(ctx context.Context, participant *entity.SessionParticipant) error {
	row, err := r.queries.AddSessionParticipant(ctx, db.AddSessionParticipantParams{
		SessionID: participant.SessionID,
		UserID:    participant.UserID,
		RoleID:    participant.RoleID,
		Seat:      participant.Seat,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domainerrors.ErrInvalidRoomState
		}
		return fmt.Errorf("session_repo: add participant: %w", err)
	}
	participant.ID = row.ID
	participant.JoinedAt = row.JoinedAt
	return nil
}

func (r *SessionRepo) GetParticipants(ctx context.Context, sessionID uuid.UUID) ([]*entity.SessionParticipant, error) {
	rows, err := r.queries.GetSessionParticipants(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session_repo: get participants: %w", err)
	}
	participants := make([]*entity.SessionParticipant, len(rows))
	for i, row := range rows {
		participants[i] = &entity.SessionParticipant{
			ID:        row.ID,
			SessionID: row.SessionID,
			UserID:    row.UserID,
			RoleID:    row.RoleID,
			Seat:      row.Seat,
			JoinedAt:  row.JoinedAt,
		}
	}
	return participants, nil
}

func (r *SessionRepo) GetParticipantHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]repository.SessionHistoryRow, error) {
	rows, err := r.queries.GetParticipantHistory(ctx, db.GetParticipantHistoryParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("session_repo: get history: %w", err)
	}

	history := make([]repository.SessionHistoryRow, len(rows))
	for i, row := range rows {
		history[i] = repository.SessionHistoryRow{
			SessionID:           row.SessionID,
			ScenarioTitle:       row.ScenarioTitle,
			Difficulty:          entity.Difficulty(row.Difficulty),
			RolePlayed:          row.RolePlayed,
			OpponentDisplayName: row.OpponentDisplayName,
			OverallScore:        row.OverallScore,
			ObjectiveAchieved:   row.ObjectiveAchieved,
			EloDelta:            row.EloDelta,
			PlayedAt:            row.PlayedAt,
		}
	}
	return history, nil
}

func dbSessionToEntity(row db.Session) *entity.Session {
	return &entity.Session{
		ID:         row.ID,
		RoomID:     row.RoomID,
		ScenarioID: row.ScenarioID,
		Difficulty: entity.Difficulty(row.Difficulty),
		State:      entity.SessionState(row.State),
		SpinSeed:   row.SpinSeed,
		StartedAt:  row.StartedAt,
		EndedAt:    row.EndedAt,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}