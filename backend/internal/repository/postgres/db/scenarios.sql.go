package db

import (
	"context"

	"github.com/google/uuid"
)

const getScenarioByID = `-- name: GetScenarioByID :one
SELECT id, context_id, title, difficulty, background_context, role_a_objective, role_a_constraints, role_b_objective, role_b_constraints, prep_duration_seconds, session_duration_seconds, is_active, created_at, updated_at FROM scenarios WHERE id = $1 AND is_active = true
`

func (q *Queries) GetScenarioByID(ctx context.Context, id uuid.UUID) (Scenario, error) {
	row := q.db.QueryRow(ctx, getScenarioByID, id)
	var i Scenario
	err := row.Scan(
		&i.ID,
		&i.ContextID,
		&i.Title,
		&i.Difficulty,
		&i.BackgroundContext,
		&i.RoleAObjective,
		&i.RoleAConstraints,
		&i.RoleBObjective,
		&i.RoleBConstraints,
		&i.PrepDurationSeconds,
		&i.SessionDurationSeconds,
		&i.IsActive,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getRandomScenario = `-- name: GetRandomScenario :one
SELECT id, context_id, title, difficulty, background_context, role_a_objective, role_a_constraints, role_b_objective, role_b_constraints, prep_duration_seconds, session_duration_seconds, is_active, created_at, updated_at FROM scenarios
WHERE is_active = true
  AND difficulty = $1
  AND ($2::uuid IS NULL OR context_id = $2)
ORDER BY RANDOM()
LIMIT 1
`

type GetRandomScenarioParams struct {
	Difficulty string     `json:"difficulty"`
	ContextID  *uuid.UUID `json:"context_id"`
}

func (q *Queries) GetRandomScenario(ctx context.Context, arg GetRandomScenarioParams) (Scenario, error) {
	row := q.db.QueryRow(ctx, getRandomScenario, arg.Difficulty, arg.ContextID)
	var i Scenario
	err := row.Scan(
		&i.ID,
		&i.ContextID,
		&i.Title,
		&i.Difficulty,
		&i.BackgroundContext,
		&i.RoleAObjective,
		&i.RoleAConstraints,
		&i.RoleBObjective,
		&i.RoleBConstraints,
		&i.PrepDurationSeconds,
		&i.SessionDurationSeconds,
		&i.IsActive,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listScenarios = `-- name: ListScenarios :many
SELECT id, context_id, title, difficulty, background_context, role_a_objective, role_a_constraints, role_b_objective, role_b_constraints, prep_duration_seconds, session_duration_seconds, is_active, created_at, updated_at FROM scenarios
WHERE is_active = true
  AND ($1::varchar IS NULL OR difficulty = $1)
  AND ($2::uuid IS NULL OR context_id = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4
`

type ListScenariosParams struct {
	Difficulty *string    `json:"difficulty"`
	ContextID  *uuid.UUID `json:"context_id"`
	Limit      int32      `json:"limit"`
	Offset     int32      `json:"offset"`
}

func (q *Queries) ListScenarios(ctx context.Context, arg ListScenariosParams) ([]Scenario, error) {
	rows, err := q.db.Query(ctx, listScenarios,
		arg.Difficulty,
		arg.ContextID,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Scenario
	for rows.Next() {
		var i Scenario
		if err := rows.Scan(
			&i.ID,
			&i.ContextID,
			&i.Title,
			&i.Difficulty,
			&i.BackgroundContext,
			&i.RoleAObjective,
			&i.RoleAConstraints,
			&i.RoleBObjective,
			&i.RoleBConstraints,
			&i.PrepDurationSeconds,
			&i.SessionDurationSeconds,
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

const countScenarios = `-- name: CountScenarios :one
SELECT COUNT(*) FROM scenarios
WHERE is_active = true
  AND ($1::varchar IS NULL OR difficulty = $1)
  AND ($2::uuid IS NULL OR context_id = $2)
`

type CountScenariosParams struct {
	Difficulty *string    `json:"difficulty"`
	ContextID  *uuid.UUID `json:"context_id"`
}

func (q *Queries) CountScenarios(ctx context.Context, arg CountScenariosParams) (int64, error) {
	row := q.db.QueryRow(ctx, countScenarios, arg.Difficulty, arg.ContextID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createScenario = `-- name: CreateScenario :one
INSERT INTO scenarios (
    context_id, title, difficulty, background_context,
    role_a_objective, role_a_constraints,
    role_b_objective, role_b_constraints,
    prep_duration_seconds, session_duration_seconds
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
RETURNING id, context_id, title, difficulty, background_context, role_a_objective, role_a_constraints, role_b_objective, role_b_constraints, prep_duration_seconds, session_duration_seconds, is_active, created_at, updated_at
`

type CreateScenarioParams struct {
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
}

func (q *Queries) CreateScenario(ctx context.Context, arg CreateScenarioParams) (Scenario, error) {
	row := q.db.QueryRow(ctx, createScenario,
		arg.ContextID,
		arg.Title,
		arg.Difficulty,
		arg.BackgroundContext,
		arg.RoleAObjective,
		arg.RoleAConstraints,
		arg.RoleBObjective,
		arg.RoleBConstraints,
		arg.PrepDurationSeconds,
		arg.SessionDurationSeconds,
	)
	var i Scenario
	err := row.Scan(
		&i.ID,
		&i.ContextID,
		&i.Title,
		&i.Difficulty,
		&i.BackgroundContext,
		&i.RoleAObjective,
		&i.RoleAConstraints,
		&i.RoleBObjective,
		&i.RoleBConstraints,
		&i.PrepDurationSeconds,
		&i.SessionDurationSeconds,
		&i.IsActive,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateScenario = `-- name: UpdateScenario :one
UPDATE scenarios
SET title=$2, difficulty=$3, background_context=$4,
    role_a_objective=$5, role_a_constraints=$6,
    role_b_objective=$7, role_b_constraints=$8,
    prep_duration_seconds=$9, session_duration_seconds=$10,
    updated_at=NOW()
WHERE id=$1 AND is_active=true
RETURNING id, context_id, title, difficulty, background_context, role_a_objective, role_a_constraints, role_b_objective, role_b_constraints, prep_duration_seconds, session_duration_seconds, is_active, created_at, updated_at
`

type UpdateScenarioParams struct {
	ID                     uuid.UUID `json:"id"`
	Title                  string    `json:"title"`
	Difficulty             string    `json:"difficulty"`
	BackgroundContext      string    `json:"background_context"`
	RoleAObjective         string    `json:"role_a_objective"`
	RoleAConstraints       []byte    `json:"role_a_constraints"`
	RoleBObjective         string    `json:"role_b_objective"`
	RoleBConstraints       []byte    `json:"role_b_constraints"`
	PrepDurationSeconds    int32     `json:"prep_duration_seconds"`
	SessionDurationSeconds int32     `json:"session_duration_seconds"`
}

func (q *Queries) UpdateScenario(ctx context.Context, arg UpdateScenarioParams) (Scenario, error) {
	row := q.db.QueryRow(ctx, updateScenario,
		arg.ID,
		arg.Title,
		arg.Difficulty,
		arg.BackgroundContext,
		arg.RoleAObjective,
		arg.RoleAConstraints,
		arg.RoleBObjective,
		arg.RoleBConstraints,
		arg.PrepDurationSeconds,
		arg.SessionDurationSeconds,
	)
	var i Scenario
	err := row.Scan(
		&i.ID,
		&i.ContextID,
		&i.Title,
		&i.Difficulty,
		&i.BackgroundContext,
		&i.RoleAObjective,
		&i.RoleAConstraints,
		&i.RoleBObjective,
		&i.RoleBConstraints,
		&i.PrepDurationSeconds,
		&i.SessionDurationSeconds,
		&i.IsActive,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const softDeleteScenario = `-- name: SoftDeleteScenario :exec
UPDATE scenarios SET is_active=false, updated_at=NOW() WHERE id=$1
`

func (q *Queries) SoftDeleteScenario(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, softDeleteScenario, id)
	return err
}