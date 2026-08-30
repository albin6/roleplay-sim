package db

import (
	"context"

	"github.com/google/uuid"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, display_name, avatar_url, role)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, username, email, password_hash, display_name, avatar_url, elo_rating, total_sessions, wins, losses, role, is_active, created_at, updated_at
`

type CreateUserParams struct {
	Username     string   `json:"username"`
	Email        string   `json:"email"`
	PasswordHash string   `json:"password_hash"`
	DisplayName  string   `json:"display_name"`
	AvatarUrl    *string  `json:"avatar_url"`
	Role         string   `json:"role"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.Username,
		arg.Email,
		arg.PasswordHash,
		arg.DisplayName,
		arg.AvatarUrl,
		arg.Role,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.PasswordHash,
		&i.DisplayName,
		&i.AvatarUrl,
		&i.EloRating,
		&i.TotalSessions,
		&i.Wins,
		&i.Losses,
		&i.Role,
		&i.IsActive,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, username, email, password_hash, display_name, avatar_url, elo_rating, total_sessions, wins, losses, role, is_active, created_at, updated_at FROM users WHERE id = $1 AND is_active = true
`

func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	row := q.db.QueryRow(ctx, getUserByID, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.PasswordHash,
		&i.DisplayName,
		&i.AvatarUrl,
		&i.EloRating,
		&i.TotalSessions,
		&i.Wins,
		&i.Losses,
		&i.Role,
		&i.IsActive,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, username, email, password_hash, display_name, avatar_url, elo_rating, total_sessions, wins, losses, role, is_active, created_at, updated_at FROM users WHERE email = $1 AND is_active = true
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.PasswordHash,
		&i.DisplayName,
		&i.AvatarUrl,
		&i.EloRating,
		&i.TotalSessions,
		&i.Wins,
		&i.Losses,
		&i.Role,
		&i.IsActive,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByUsername = `-- name: GetUserByUsername :one
SELECT id, username, email, password_hash, display_name, avatar_url, elo_rating, total_sessions, wins, losses, role, is_active, created_at, updated_at FROM users WHERE username = $1 AND is_active = true
`

func (q *Queries) GetUserByUsername(ctx context.Context, username string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByUsername, username)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.PasswordHash,
		&i.DisplayName,
		&i.AvatarUrl,
		&i.EloRating,
		&i.TotalSessions,
		&i.Wins,
		&i.Losses,
		&i.Role,
		&i.IsActive,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateUser = `-- name: UpdateUser :one
UPDATE users
SET display_name = $2, avatar_url = $3, updated_at = NOW()
WHERE id = $1 AND is_active = true
RETURNING id, username, email, password_hash, display_name, avatar_url, elo_rating, total_sessions, wins, losses, role, is_active, created_at, updated_at
`

type UpdateUserParams struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
	AvatarUrl   *string   `json:"avatar_url"`
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUser, arg.ID, arg.DisplayName, arg.AvatarUrl)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.PasswordHash,
		&i.DisplayName,
		&i.AvatarUrl,
		&i.EloRating,
		&i.TotalSessions,
		&i.Wins,
		&i.Losses,
		&i.Role,
		&i.IsActive,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateEloRating = `-- name: UpdateEloRating :exec
UPDATE users
SET elo_rating = $2, updated_at = NOW()
WHERE id = $1
`

type UpdateEloRatingParams struct {
	ID        uuid.UUID `json:"id"`
	EloRating float64   `json:"elo_rating"`
}

func (q *Queries) UpdateEloRating(ctx context.Context, arg UpdateEloRatingParams) error {
	_, err := q.db.Exec(ctx, updateEloRating, arg.ID, arg.EloRating)
	return err
}

const incrementWin = `-- name: IncrementWin :exec
UPDATE users
SET total_sessions = total_sessions + 1, wins = wins + 1, updated_at = NOW()
WHERE id = $1
`

func (q *Queries) IncrementWin(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, incrementWin, id)
	return err
}

const incrementLoss = `-- name: IncrementLoss :exec
UPDATE users
SET total_sessions = total_sessions + 1, losses = losses + 1, updated_at = NOW()
WHERE id = $1
`

func (q *Queries) IncrementLoss(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, incrementLoss, id)
	return err
}

const incrementSessions = `-- name: IncrementSessions :exec
UPDATE users
SET total_sessions = total_sessions + 1, updated_at = NOW()
WHERE id = $1
`

func (q *Queries) IncrementSessions(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, incrementSessions, id)
	return err
}

const getLeaderboard = `-- name: GetLeaderboard :many
SELECT id, username, email, password_hash, display_name, avatar_url, elo_rating, total_sessions, wins, losses, role, is_active, created_at, updated_at FROM users
WHERE is_active = true
ORDER BY elo_rating DESC, total_sessions DESC
LIMIT $1 OFFSET $2
`

type GetLeaderboardParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) GetLeaderboard(ctx context.Context, arg GetLeaderboardParams) ([]User, error) {
	rows, err := q.db.Query(ctx, getLeaderboard, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []User
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.Email,
			&i.PasswordHash,
			&i.DisplayName,
			&i.AvatarUrl,
			&i.EloRating,
			&i.TotalSessions,
			&i.Wins,
			&i.Losses,
			&i.Role,
			&i.IsActive,
			&i.CreatedAt,
			&i.UpdatedAt,
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

const getLeaderboardCount = `-- name: GetLeaderboardCount :one
SELECT COUNT(*) FROM users WHERE is_active = true
`

func (q *Queries) GetLeaderboardCount(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, getLeaderboardCount)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const getUserRank = `-- name: GetUserRank :one
SELECT COUNT(*) + 1 AS rank
FROM users
WHERE elo_rating > (SELECT elo_rating FROM users WHERE id = $1)
  AND is_active = true
`

func (q *Queries) GetUserRank(ctx context.Context, id uuid.UUID) (int64, error) {
	row := q.db.QueryRow(ctx, getUserRank, id)
	var rank int64
	err := row.Scan(&rank)
	return rank, err
}