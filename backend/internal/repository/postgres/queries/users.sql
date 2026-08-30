-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, display_name, avatar_url, role)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND is_active = true;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 AND is_active = true;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1 AND is_active = true;

-- name: UpdateUser :one
UPDATE users
SET display_name = $2, avatar_url = $3, updated_at = NOW()
WHERE id = $1 AND is_active = true
RETURNING *;

-- name: UpdateEloRating :exec
UPDATE users
SET elo_rating = $2, updated_at = NOW()
WHERE id = $1;

-- name: IncrementWin :exec
UPDATE users
SET total_sessions = total_sessions + 1, wins = wins + 1, updated_at = NOW()
WHERE id = $1;

-- name: IncrementLoss :exec
UPDATE users
SET total_sessions = total_sessions + 1, losses = losses + 1, updated_at = NOW()
WHERE id = $1;

-- name: IncrementSessions :exec
UPDATE users
SET total_sessions = total_sessions + 1, updated_at = NOW()
WHERE id = $1;

-- name: GetLeaderboard :many
SELECT * FROM users
WHERE is_active = true
ORDER BY elo_rating DESC, total_sessions DESC
LIMIT $1 OFFSET $2;

-- name: GetLeaderboardCount :one
SELECT COUNT(*) FROM users WHERE is_active = true;

-- name: GetUserRank :one
SELECT COUNT(*) + 1 AS rank
FROM users
WHERE elo_rating > (SELECT elo_rating FROM users WHERE id = $1)
  AND is_active = true;
