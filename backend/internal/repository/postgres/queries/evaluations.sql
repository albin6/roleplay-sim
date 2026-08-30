-- name: CreateEvaluation :one
INSERT INTO evaluations (
    session_id, participant_id, overall_score, objective_achieved,
    summary_feedback,
    strengths, areas_for_improvement,
    llm_model_used, prompt_version, raw_transcript, raw_llm_response,
    stt_duration_ms, llm_duration_ms
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
RETURNING *;

-- name: CreateRubricScore :one
INSERT INTO rubric_scores (
    evaluation_id, dimension, score, weight, justification, evidence_quotes
) VALUES ($1,$2,$3,$4,$5,$6)
RETURNING *;

-- name: GetEvaluationBySessionAndParticipant :one
SELECT * FROM evaluations
WHERE session_id = $1 AND participant_id = $2;

-- name: GetRubricScoresByEvaluation :many
SELECT * FROM rubric_scores
WHERE evaluation_id = $1
ORDER BY dimension;

-- name: EvaluationExists :one
SELECT EXISTS(
    SELECT 1 FROM evaluations WHERE session_id=$1 AND participant_id=$2
) AS exists;
