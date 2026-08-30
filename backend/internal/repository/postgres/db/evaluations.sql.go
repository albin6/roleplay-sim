package db

import (
	"context"

	"github.com/google/uuid"
)

const createEvaluation = `-- name: CreateEvaluation :one
INSERT INTO evaluations (
    session_id, participant_id, overall_score, objective_achieved,
    objective_achievement_reasoning, summary_feedback,
    strengths, areas_for_improvement,
    elo_delta, elo_rating_before, elo_rating_after,
    llm_model_used, stt_provider, prompt_version, raw_llm_response, is_fallback
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
RETURNING id, session_id, participant_id, overall_score, objective_achieved, objective_achievement_reasoning, summary_feedback, strengths, areas_for_improvement, elo_delta, elo_rating_before, elo_rating_after, llm_model_used, stt_provider, prompt_version, raw_llm_response, is_fallback, created_at
`

type CreateEvaluationParams struct {
	SessionID                     uuid.UUID `json:"session_id"`
	ParticipantID                 uuid.UUID `json:"participant_id"`
	OverallScore                  *float64  `json:"overall_score"`
	ObjectiveAchieved             *bool     `json:"objective_achieved"`
	ObjectiveAchievementReasoning *string   `json:"objective_achievement_reasoning"`
	SummaryFeedback               *string   `json:"summary_feedback"`
	Strengths                     []byte    `json:"strengths"`
	AreasForImprovement           []byte    `json:"areas_for_improvement"`
	EloDelta                      *float64  `json:"elo_delta"`
	EloRatingBefore               *float64  `json:"elo_rating_before"`
	EloRatingAfter                *float64  `json:"elo_rating_after"`
	LlmModelUsed                  *string   `json:"llm_model_used"`
	SttProvider                   *string   `json:"stt_provider"`
	PromptVersion                 *string   `json:"prompt_version"`
	RawLlmResponse                []byte    `json:"raw_llm_response"`
	IsFallback                    bool      `json:"is_fallback"`
}

func (q *Queries) CreateEvaluation(ctx context.Context, arg CreateEvaluationParams) (Evaluation, error) {
	row := q.db.QueryRow(ctx, createEvaluation,
		arg.SessionID,
		arg.ParticipantID,
		arg.OverallScore,
		arg.ObjectiveAchieved,
		arg.ObjectiveAchievementReasoning,
		arg.SummaryFeedback,
		arg.Strengths,
		arg.AreasForImprovement,
		arg.EloDelta,
		arg.EloRatingBefore,
		arg.EloRatingAfter,
		arg.LlmModelUsed,
		arg.SttProvider,
		arg.PromptVersion,
		arg.RawLlmResponse,
		arg.IsFallback,
	)
	var i Evaluation
	err := row.Scan(
		&i.ID,
		&i.SessionID,
		&i.ParticipantID,
		&i.OverallScore,
		&i.ObjectiveAchieved,
		&i.ObjectiveAchievementReasoning,
		&i.SummaryFeedback,
		&i.Strengths,
		&i.AreasForImprovement,
		&i.EloDelta,
		&i.EloRatingBefore,
		&i.EloRatingAfter,
		&i.LlmModelUsed,
		&i.SttProvider,
		&i.PromptVersion,
		&i.RawLlmResponse,
		&i.IsFallback,
		&i.CreatedAt,
	)
	return i, err
}

const createRubricScore = `-- name: CreateRubricScore :one
INSERT INTO rubric_scores (
    evaluation_id, dimension, score, weight, justification, evidence_quotes
) VALUES ($1,$2,$3,$4,$5,$6)
RETURNING id, evaluation_id, dimension, score, weight, justification, evidence_quotes
`

type CreateRubricScoreParams struct {
	EvaluationID   uuid.UUID `json:"evaluation_id"`
	Dimension      string    `json:"dimension"`
	Score          float64   `json:"score"`
	Weight         float64   `json:"weight"`
	Justification  string    `json:"justification"`
	EvidenceQuotes []byte    `json:"evidence_quotes"`
}

func (q *Queries) CreateRubricScore(ctx context.Context, arg CreateRubricScoreParams) (RubricScore, error) {
	row := q.db.QueryRow(ctx, createRubricScore,
		arg.EvaluationID,
		arg.Dimension,
		arg.Score,
		arg.Weight,
		arg.Justification,
		arg.EvidenceQuotes,
	)
	var i RubricScore
	err := row.Scan(
		&i.ID,
		&i.EvaluationID,
		&i.Dimension,
		&i.Score,
		&i.Weight,
		&i.Justification,
		&i.EvidenceQuotes,
	)
	return i, err
}

const getEvaluationBySessionAndParticipant = `-- name: GetEvaluationBySessionAndParticipant :one
SELECT id, session_id, participant_id, overall_score, objective_achieved, objective_achievement_reasoning, summary_feedback, strengths, areas_for_improvement, elo_delta, elo_rating_before, elo_rating_after, llm_model_used, stt_provider, prompt_version, raw_llm_response, is_fallback, created_at FROM evaluations
WHERE session_id = $1 AND participant_id = $2
`

type GetEvaluationBySessionAndParticipantParams struct {
	SessionID     uuid.UUID `json:"session_id"`
	ParticipantID uuid.UUID `json:"participant_id"`
}

func (q *Queries) GetEvaluationBySessionAndParticipant(ctx context.Context, arg GetEvaluationBySessionAndParticipantParams) (Evaluation, error) {
	row := q.db.QueryRow(ctx, getEvaluationBySessionAndParticipant, arg.SessionID, arg.ParticipantID)
	var i Evaluation
	err := row.Scan(
		&i.ID,
		&i.SessionID,
		&i.ParticipantID,
		&i.OverallScore,
		&i.ObjectiveAchieved,
		&i.ObjectiveAchievementReasoning,
		&i.SummaryFeedback,
		&i.Strengths,
		&i.AreasForImprovement,
		&i.EloDelta,
		&i.EloRatingBefore,
		&i.EloRatingAfter,
		&i.LlmModelUsed,
		&i.SttProvider,
		&i.PromptVersion,
		&i.RawLlmResponse,
		&i.IsFallback,
		&i.CreatedAt,
	)
	return i, err
}

const getRubricScoresByEvaluation = `-- name: GetRubricScoresByEvaluation :many
SELECT id, evaluation_id, dimension, score, weight, justification, evidence_quotes FROM rubric_scores
WHERE evaluation_id = $1
ORDER BY dimension
`

func (q *Queries) GetRubricScoresByEvaluation(ctx context.Context, evaluationID uuid.UUID) ([]RubricScore, error) {
	rows, err := q.db.Query(ctx, getRubricScoresByEvaluation, evaluationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []RubricScore
	for rows.Next() {
		var i RubricScore
		if err := rows.Scan(
			&i.ID,
			&i.EvaluationID,
			&i.Dimension,
			&i.Score,
			&i.Weight,
			&i.Justification,
			&i.EvidenceQuotes,
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

const evaluationExists = `-- name: EvaluationExists :one
SELECT EXISTS(
    SELECT 1 FROM evaluations WHERE session_id=$1 AND participant_id=$2
) AS exists
`

type EvaluationExistsParams struct {
	SessionID     uuid.UUID `json:"session_id"`
	ParticipantID uuid.UUID `json:"participant_id"`
}

func (q *Queries) EvaluationExists(ctx context.Context, arg EvaluationExistsParams) (bool, error) {
	row := q.db.QueryRow(ctx, evaluationExists, arg.SessionID, arg.ParticipantID)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}