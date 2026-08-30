-- name: GetScenarioByID :one
SELECT * FROM scenarios WHERE id = $1 AND is_active = true;

-- name: GetRandomScenario :one
SELECT * FROM scenarios
WHERE is_active = true
  AND difficulty = $1
  AND ($2::uuid IS NULL OR context_id = $2)
ORDER BY RANDOM()
LIMIT 1;

-- name: ListScenarios :many
SELECT * FROM scenarios
WHERE is_active = true
  AND ($1::varchar IS NULL OR difficulty = $1)
  AND ($2::uuid IS NULL OR context_id = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountScenarios :one
SELECT COUNT(*) FROM scenarios
WHERE is_active = true
  AND ($1::varchar IS NULL OR difficulty = $1)
  AND ($2::uuid IS NULL OR context_id = $2);

-- name: CreateScenario :one
INSERT INTO scenarios (
    context_id, role_a_id, role_b_id, title, difficulty, background_context,
    role_a_objective, role_a_constraints,
    role_b_objective, role_b_constraints,
    prep_duration_seconds, session_duration_seconds
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
RETURNING *;

-- name: UpdateScenario :one
UPDATE scenarios
SET title=$2, difficulty=$3, background_context=$4,
    role_a_objective=$5, role_a_constraints=$6,
    role_b_objective=$7, role_b_constraints=$8,
    prep_duration_seconds=$9, session_duration_seconds=$10,
    updated_at=NOW()
WHERE id=$1 AND is_active=true
RETURNING *;

-- name: SoftDeleteScenario :exec
UPDATE scenarios SET is_active=false, updated_at=NOW() WHERE id=$1;
