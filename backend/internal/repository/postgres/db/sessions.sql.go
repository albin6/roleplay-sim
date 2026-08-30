package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createSession = `-- name: CreateSession :one
INSERT INTO sessions (room_id, scenario_id, difficulty, state)
VALUES ($1, $2, $3, $4)
RETURNING id, room_id, scenario_id, difficulty, state, spin_seed, started_at, ended_at, created_at, updated_at
`

type CreateSessionParams struct {
	RoomID     string    `json:"room_id"`
	ScenarioID uuid.UUID `json:"scenario_id"`
	Difficulty string    `json:"difficulty"`
	State      string    `json:"state"`
}

func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error) {
	row := q.db.QueryRow(ctx, createSession,
		arg.RoomID,
		arg.ScenarioID,
		arg.Difficulty,
		arg.State,
	)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.RoomID,
		&i.ScenarioID,
		&i.Difficulty,
		&i.State,
		&i.SpinSeed,
		&i.StartedAt,
		&i.EndedAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getSessionByID = `-- name: GetSessionByID :one
SELECT id, room_id, scenario_id, difficulty, state, spin_seed, started_at, ended_at, created_at, updated_at FROM sessions WHERE id = $1
`

func (q *Queries) GetSessionByID(ctx context.Context, id uuid.UUID) (Session, error) {
	row := q.db.QueryRow(ctx, getSessionByID, id)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.RoomID,
		&i.ScenarioID,
		&i.Difficulty,
		&i.State,
		&i.SpinSeed,
		&i.StartedAt,
		&i.EndedAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getSessionByRoomID = `-- name: GetSessionByRoomID :one
SELECT id, room_id, scenario_id, difficulty, state, spin_seed, started_at, ended_at, created_at, updated_at FROM sessions WHERE room_id = $1
`

func (q *Queries) GetSessionByRoomID(ctx context.Context, roomID string) (Session, error) {
	row := q.db.QueryRow(ctx, getSessionByRoomID, roomID)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.RoomID,
		&i.ScenarioID,
		&i.Difficulty,
		&i.State,
		&i.SpinSeed,
		&i.StartedAt,
		&i.EndedAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateSessionState = `-- name: UpdateSessionState :exec
UPDATE sessions SET state = $2, updated_at = NOW() WHERE id = $1
`

type UpdateSessionStateParams struct {
	ID    uuid.UUID `json:"id"`
	State string    `json:"state"`
}

func (q *Queries) UpdateSessionState(ctx context.Context, arg UpdateSessionStateParams) error {
	_, err := q.db.Exec(ctx, updateSessionState, arg.ID, arg.State)
	return err
}

const setSessionStartedAt = `-- name: SetSessionStartedAt :exec
UPDATE sessions SET started_at = $2, updated_at = NOW() WHERE id = $1
`

type SetSessionStartedAtParams struct {
	ID        uuid.UUID  `json:"id"`
	StartedAt *time.Time `json:"started_at"`
}

func (q *Queries) SetSessionStartedAt(ctx context.Context, arg SetSessionStartedAtParams) error {
	_, err := q.db.Exec(ctx, setSessionStartedAt, arg.ID, arg.StartedAt)
	return err
}

const setSessionEndedAt = `-- name: SetSessionEndedAt :exec
UPDATE sessions SET ended_at = $2, updated_at = NOW() WHERE id = $1
`

type SetSessionEndedAtParams struct {
	ID      uuid.UUID  `json:"id"`
	EndedAt *time.Time `json:"ended_at"`
}

func (q *Queries) SetSessionEndedAt(ctx context.Context, arg SetSessionEndedAtParams) error {
	_, err := q.db.Exec(ctx, setSessionEndedAt, arg.ID, arg.EndedAt)
	return err
}

const addSessionParticipant = `-- name: AddSessionParticipant :one
INSERT INTO session_participants (session_id, user_id, role_id, seat)
VALUES ($1, $2, $3, $4)
RETURNING id, session_id, user_id, role_id, seat, joined_at
`

type AddSessionParticipantParams struct {
	SessionID uuid.UUID `json:"session_id"`
	UserID    uuid.UUID `json:"user_id"`
	RoleID    uuid.UUID `json:"role_id"`
	Seat      string    `json:"seat"`
}

func (q *Queries) AddSessionParticipant(ctx context.Context, arg AddSessionParticipantParams) (SessionParticipant, error) {
	row := q.db.QueryRow(ctx, addSessionParticipant,
		arg.SessionID,
		arg.UserID,
		arg.RoleID,
		arg.Seat,
	)
	var i SessionParticipant
	err := row.Scan(
		&i.ID,
		&i.SessionID,
		&i.UserID,
		&i.RoleID,
		&i.Seat,
		&i.JoinedAt,
	)
	return i, err
}

const getSessionParticipants = `-- name: GetSessionParticipants :many
SELECT sp.id, sp.session_id, sp.user_id, sp.role_id, sp.seat, sp.joined_at, r.name as role_name, u.display_name, u.avatar_url
FROM session_participants sp
JOIN roles r ON r.id = sp.role_id
JOIN users u ON u.id = sp.user_id
WHERE sp.session_id = $1
ORDER BY sp.seat
`

type GetSessionParticipantsRow struct {
	ID          uuid.UUID `json:"id"`
	SessionID   uuid.UUID `json:"session_id"`
	UserID      uuid.UUID `json:"user_id"`
	RoleID      uuid.UUID `json:"role_id"`
	Seat        string    `json:"seat"`
	JoinedAt    time.Time `json:"joined_at"`
	RoleName    string    `json:"role_name"`
	DisplayName string    `json:"display_name"`
	AvatarUrl   *string   `json:"avatar_url"`
}

func (q *Queries) GetSessionParticipants(ctx context.Context, sessionID uuid.UUID) ([]GetSessionParticipantsRow, error) {
	rows, err := q.db.Query(ctx, getSessionParticipants, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetSessionParticipantsRow
	for rows.Next() {
		var i GetSessionParticipantsRow
		if err := rows.Scan(
			&i.ID,
			&i.SessionID,
			&i.UserID,
			&i.RoleID,
			&i.Seat,
			&i.JoinedAt,
			&i.RoleName,
			&i.DisplayName,
			&i.AvatarUrl,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getParticipantHistory = `-- name: GetParticipantHistory :many
SELECT
    s.id as session_id,
    sc.title as scenario_title,
    s.difficulty,
    r.name as role_played,
    COALESCE(u2.display_name, '') as opponent_display_name,
    e.overall_score,
    e.objective_achieved,
    e.elo_delta,
    COALESCE(s.ended_at, s.created_at) as played_at
FROM session_participants sp
JOIN sessions s ON s.id = sp.session_id
JOIN scenarios sc ON sc.id = s.scenario_id
JOIN roles r ON r.id = sp.role_id
LEFT JOIN session_participants sp2 ON sp2.session_id = s.id AND sp2.user_id != sp.user_id
LEFT JOIN users u2 ON u2.id = sp2.user_id
LEFT JOIN evaluations e ON e.session_id = s.id AND e.participant_id = sp.id
WHERE sp.user_id = $1
  AND s.state = 'complete'
ORDER BY s.ended_at DESC NULLS LAST
LIMIT $2 OFFSET $3
`

type GetParticipantHistoryParams struct {
	UserID uuid.UUID `json:"user_id"`
	Limit  int32     `json:"limit"`
	Offset int32     `json:"offset"`
}

type GetParticipantHistoryRow struct {
	SessionID           uuid.UUID `json:"session_id"`
	ScenarioTitle       string    `json:"scenario_title"`
	Difficulty          string    `json:"difficulty"`
	RolePlayed          string    `json:"role_played"`
	OpponentDisplayName string    `json:"opponent_display_name"`
	OverallScore        *float64  `json:"overall_score"`
	ObjectiveAchieved   *bool     `json:"objective_achieved"`
	EloDelta            *float64  `json:"elo_delta"`
	PlayedAt            time.Time `json:"played_at"`
}

func (q *Queries) GetParticipantHistory(ctx context.Context, arg GetParticipantHistoryParams) ([]GetParticipantHistoryRow, error) {
	rows, err := q.db.Query(ctx, getParticipantHistory, arg.UserID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetParticipantHistoryRow
	for rows.Next() {
		var i GetParticipantHistoryRow
		if err := rows.Scan(
			&i.SessionID,
			&i.ScenarioTitle,
			&i.Difficulty,
			&i.RolePlayed,
			&i.OpponentDisplayName,
			&i.OverallScore,
			&i.ObjectiveAchieved,
			&i.EloDelta,
			&i.PlayedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const countParticipantHistory = `-- name: CountParticipantHistory :one
SELECT COUNT(*)
FROM session_participants sp
JOIN sessions s ON s.id = sp.session_id
WHERE sp.user_id = $1 AND s.state = 'complete'
`

func (q *Queries) CountParticipantHistory(ctx context.Context, userID uuid.UUID) (int64, error) {
	row := q.db.QueryRow(ctx, countParticipantHistory, userID)
	var count int64
	err := row.Scan(&count)
	return count, err
}