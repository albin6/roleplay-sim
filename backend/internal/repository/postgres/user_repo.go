package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roleplay-sim/backend/internal/domain/entity"
	domainerrors "github.com/roleplay-sim/backend/internal/domain/errors"
	"github.com/roleplay-sim/backend/internal/repository/postgres/db"
)

type UserRepo struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool, queries: db.New(pool)}
}

func (r *UserRepo) Create(ctx context.Context, user *entity.User) error {
	row, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		DisplayName:  user.DisplayName,
		AvatarUrl:    user.AvatarURL,
		Role:         string(user.Role),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domainerrors.ErrUserAlreadyExists
		}
		return fmt.Errorf("user_repo: create: %w", err)
	}
	user.ID = row.ID
	user.EloRating = row.EloRating
	user.CreatedAt = row.CreatedAt
	user.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("user_repo: get by id: %w", err)
	}
	return dbUserToEntity(row), nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("user_repo: get by email: %w", err)
	}
	return dbUserToEntity(row), nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	row, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("user_repo: get by username: %w", err)
	}
	return dbUserToEntity(row), nil
}

func (r *UserRepo) Update(ctx context.Context, user *entity.User) error {
	row, err := r.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		AvatarUrl:   user.AvatarURL,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainerrors.ErrUserNotFound
		}
		return fmt.Errorf("user_repo: update: %w", err)
	}
	user.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *UserRepo) UpdateEloRating(ctx context.Context, userID uuid.UUID, newRating float64) error {
	return r.queries.UpdateEloRating(ctx, db.UpdateEloRatingParams{
		ID:        userID,
		EloRating: newRating,
	})
}

func (r *UserRepo) IncrementSessionCount(ctx context.Context, userID uuid.UUID, won bool) error {
	if won {
		return r.queries.IncrementWin(ctx, userID)
	}
	return r.queries.IncrementLoss(ctx, userID)
}

func (r *UserRepo) GetLeaderboard(ctx context.Context, limit, offset int) ([]*entity.User, int64, error) {
	rows, err := r.queries.GetLeaderboard(ctx, db.GetLeaderboardParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("user_repo: get leaderboard: %w", err)
	}
	count, err := r.queries.GetLeaderboardCount(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("user_repo: get leaderboard count: %w", err)
	}
	users := make([]*entity.User, len(rows))
	for i, row := range rows {
		users[i] = dbUserToEntity(row)
	}
	return users, count, nil
}

func (r *UserRepo) GetRank(ctx context.Context, userID uuid.UUID) (int64, error) {
	rank, err := r.queries.GetUserRank(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("user_repo: get rank: %w", err)
	}
	return rank, nil
}

// dbUserToEntity maps a DB row to a domain entity.
func dbUserToEntity(row db.User) *entity.User {
	return &entity.User{
		ID:            row.ID,
		Username:      row.Username,
		Email:         row.Email,
		PasswordHash:  row.PasswordHash,
		DisplayName:   row.DisplayName,
		AvatarURL:     row.AvatarUrl,
		EloRating:     row.EloRating,
		TotalSessions: int(row.TotalSessions),
		Wins:          int(row.Wins),
		Losses:        int(row.Losses),
		Role:          entity.UserRole(row.Role),
		IsActive:      row.IsActive,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

// isUniqueViolation checks if the error is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, &pgxUniqueError{}) || containsCode(err, "23505")
}

type pgxUniqueError struct{}
func (e *pgxUniqueError) Error() string { return "unique violation" }
func (e *pgxUniqueError) Is(target error) bool {
	_, ok := target.(*pgxUniqueError)
	return ok
}

func containsCode(err error, code string) bool {
	type pgError interface{ SQLState() string }
	var pge pgError
	if errors.As(err, &pge) {
		return pge.SQLState() == code
	}
	return false
}
