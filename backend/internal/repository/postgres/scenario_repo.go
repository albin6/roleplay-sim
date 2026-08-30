package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roleplay-sim/backend/internal/domain/entity"
	domainerrors "github.com/roleplay-sim/backend/internal/domain/errors"
	"github.com/roleplay-sim/backend/internal/repository/postgres/db"
)

type ScenarioRepo struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewScenarioRepo(pool *pgxpool.Pool) *ScenarioRepo {
	return &ScenarioRepo{pool: pool, queries: db.New(pool)}
}

func (r *ScenarioRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Scenario, error) {
	row, err := r.queries.GetScenarioByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrScenarioNotFound
		}
		return nil, fmt.Errorf("scenario_repo: get by id: %w", err)
	}
	return dbScenarioToEntity(row), nil
}

func (r *ScenarioRepo) GetRandom(ctx context.Context, difficulty entity.Difficulty, contextID *uuid.UUID) (*entity.Scenario, error) {
	row, err := r.queries.GetRandomScenario(ctx, db.GetRandomScenarioParams{
		Difficulty: string(difficulty),
		ContextID:  contextID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrScenarioNotFound
		}
		return nil, fmt.Errorf("scenario_repo: get random: %w", err)
	}
	return dbScenarioToEntity(row), nil
}

func (r *ScenarioRepo) ListAll(ctx context.Context, difficulty *entity.Difficulty, contextID *uuid.UUID, limit, offset int) ([]*entity.Scenario, int64, error) {
	var diffStr *string
	if difficulty != nil {
		s := string(*difficulty)
		diffStr = &s
	}

	rows, err := r.queries.ListScenarios(ctx, db.ListScenariosParams{
		Difficulty: diffStr,
		ContextID:  contextID,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("scenario_repo: list all: %w", err)
	}

	count, err := r.queries.CountScenarios(ctx, db.CountScenariosParams{
		Difficulty: diffStr,
		ContextID:  contextID,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("scenario_repo: count all: %w", err)
	}

	scenarios := make([]*entity.Scenario, len(rows))
	for i, row := range rows {
		scenarios[i] = dbScenarioToEntity(row)
	}
	return scenarios, count, nil
}

func (r *ScenarioRepo) Create(ctx context.Context, scenario *entity.Scenario) error {
	constraintsA, _ := json.Marshal(scenario.RoleAConstraints)
	constraintsB, _ := json.Marshal(scenario.RoleBConstraints)

	row, err := r.queries.CreateScenario(ctx, db.CreateScenarioParams{
		ContextID:              scenario.ContextID,
		Title:                  scenario.Title,
		Difficulty:             string(scenario.Difficulty),
		BackgroundContext:      scenario.BackgroundContext,
		RoleAObjective:         scenario.RoleAObjective,
		RoleAConstraints:       constraintsA,
		RoleBObjective:         scenario.RoleBObjective,
		RoleBConstraints:       constraintsB,
		PrepDurationSeconds:    int32(scenario.PrepDurationSeconds),
		SessionDurationSeconds: int32(scenario.SessionDurationSeconds),
	})
	if err != nil {
		return fmt.Errorf("scenario_repo: create: %w", err)
	}
	scenario.ID = row.ID
	scenario.CreatedAt = row.CreatedAt
	scenario.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *ScenarioRepo) Update(ctx context.Context, scenario *entity.Scenario) error {
	constraintsA, _ := json.Marshal(scenario.RoleAConstraints)
	constraintsB, _ := json.Marshal(scenario.RoleBConstraints)

	row, err := r.queries.UpdateScenario(ctx, db.UpdateScenarioParams{
		ID:                     scenario.ID,
		Title:                  scenario.Title,
		Difficulty:             string(scenario.Difficulty),
		BackgroundContext:      scenario.BackgroundContext,
		RoleAObjective:         scenario.RoleAObjective,
		RoleAConstraints:       constraintsA,
		RoleBObjective:         scenario.RoleBObjective,
		RoleBConstraints:       constraintsB,
		PrepDurationSeconds:    int32(scenario.PrepDurationSeconds),
		SessionDurationSeconds: int32(scenario.SessionDurationSeconds),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainerrors.ErrScenarioNotFound
		}
		return fmt.Errorf("scenario_repo: update: %w", err)
	}
	scenario.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *ScenarioRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	err := r.queries.SoftDeleteScenario(ctx, id)
	if err != nil {
		return fmt.Errorf("scenario_repo: delete: %w", err)
	}
	return nil
}

func dbScenarioToEntity(row db.Scenario) *entity.Scenario {
	var constraintsA, constraintsB []string
	_ = json.Unmarshal(row.RoleAConstraints, &constraintsA)
	_ = json.Unmarshal(row.RoleBConstraints, &constraintsB)

	return &entity.Scenario{
		ID:                     row.ID,
		ContextID:              row.ContextID,
		Title:                  row.Title,
		Difficulty:             entity.Difficulty(row.Difficulty),
		BackgroundContext:      row.BackgroundContext,
		RoleAObjective:         row.RoleAObjective,
		RoleAConstraints:       constraintsA,
		RoleBObjective:         row.RoleBObjective,
		RoleBConstraints:       constraintsB,
		PrepDurationSeconds:    int(row.PrepDurationSeconds),
		SessionDurationSeconds: int(row.SessionDurationSeconds),
		IsActive:               row.IsActive,
		CreatedAt:              row.CreatedAt,
		UpdatedAt:              row.UpdatedAt,
	}
}