-- name: CreateSession :one
INSERT INTO sessions (room_id, scenario_id, difficulty, state)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSessionByID :one
SELECT * FROM sessions WHERE id = $1;

-- name: GetSessionByRoomID :one
SELECT * FROM sessions WHERE room_id = $1;

-- name: UpdateSessionState :exec
UPDATE sessions SET state = $2, updated_at = NOW() WHERE id = $1;

-- name: SetSessionStartedAt :exec
UPDATE sessions SET session_started_at = $2, updated_at = NOW() WHERE id = $1;

-- name: SetSessionEndedAt :exec
UPDATE sessions SET session_ended_at = $2, updated_at = NOW() WHERE id = $1;

-- name: AddSessionParticipant :one
INSERT INTO session_participants (session_id, user_id, role_id, seat, elo_before)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSessionParticipants :many
SELECT sp.*, r.name as role_name, u.display_name, u.avatar_url
FROM session_participants sp
JOIN roles r ON r.id = sp.role_id
JOIN users u ON u.id = sp.user_id
WHERE sp.session_id = $1
ORDER BY sp.seat;

-- name: GetParticipantHistory :many
SELECT
    s.id as session_id,
    sc.title as scenario_title,
    s.difficulty,
    r.name as role_played,
    u2.display_name as opponent_display_name,
    e.overall_score,
    e.objective_achieved,
    e.elo_delta,
    s.session_ended_at as played_at
FROM session_participants sp
JOIN sessions s ON s.id = sp.session_id
JOIN scenarios sc ON sc.id = s.scenario_id
JOIN roles r ON r.id = sp.role_id
LEFT JOIN session_participants sp2 ON sp2.session_id = s.id AND sp2.user_id != sp.user_id
LEFT JOIN users u2 ON u2.id = sp2.user_id
LEFT JOIN evaluations e ON e.session_id = s.id AND e.participant_id = sp.id
WHERE sp.user_id = $1
  AND s.state = 'complete'
ORDER BY s.session_ended_at DESC
LIMIT $2 OFFSET $3;

-- name: CountParticipantHistory :one
SELECT COUNT(*)
FROM session_participants sp
JOIN sessions s ON s.id = sp.session_id
WHERE sp.user_id = $1 AND s.state = 'complete';
