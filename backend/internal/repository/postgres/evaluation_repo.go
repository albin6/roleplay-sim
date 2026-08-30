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

type EvaluationRepo struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewEvaluationRepo(pool *pgxpool.Pool) *EvaluationRepo {
	return &EvaluationRepo{pool: pool, queries: db.New(pool)}
}

func (r *EvaluationRepo) Create(ctx context.Context, eval *entity.Evaluation, scores []*entity.RubricScore) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("evaluation_repo: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	strengthsBytes, _ := json.Marshal(eval.Strengths)
	areasBytes, _ := json.Marshal(eval.AreasForImprovement)

	row, err := qtx.CreateEvaluation(ctx, db.CreateEvaluationParams{
		SessionID:                     eval.SessionID,
		ParticipantID:                 eval.ParticipantID,
		OverallScore:                  &eval.OverallScore,
		ObjectiveAchieved:             eval.ObjectiveAchieved,
		ObjectiveAchievementReasoning: &eval.ObjectiveAchievementReason,
		SummaryFeedback:               &eval.SummaryFeedback,
		Strengths:                     strengthsBytes,
		AreasForImprovement:           areasBytes,
		EloDelta:                      &eval.EloDelta,
		EloRatingBefore:               &eval.EloRatingBefore,
		EloRatingAfter:                &eval.EloRatingAfter,
		LlmModelUsed:                  &eval.LLMModelUsed,
		SttProvider:                   &eval.STTProvider,
		PromptVersion:                 &eval.PromptVersion,
		IsFallback:                    eval.IsFallback,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domainerrors.ErrEvaluationExists
		}
		return fmt.Errorf("evaluation_repo: create evaluation: %w", err)
	}

	eval.ID = row.ID
	eval.CreatedAt = row.CreatedAt

	for _, score := range scores {
		quotesBytes, _ := json.Marshal(score.EvidenceQuotes)
		scoreRow, err := qtx.CreateRubricScore(ctx, db.CreateRubricScoreParams{
			EvaluationID:   eval.ID,
			Dimension:      string(score.Dimension),
			Score:          score.Score,
			Weight:         score.Weight,
			Justification:  score.Justification,
			EvidenceQuotes: quotesBytes,
		})
		if err != nil {
			return fmt.Errorf("evaluation_repo: create rubric score: %w", err)
		}
		score.ID = scoreRow.ID
		score.EvaluationID = eval.ID
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("evaluation_repo: commit tx: %w", err)
	}

	return nil
}

func (r *EvaluationRepo) GetBySessionAndParticipant(ctx context.Context, sessionID, participantID uuid.UUID) (*entity.Evaluation, error) {
	row, err := r.queries.GetEvaluationBySessionAndParticipant(ctx, db.GetEvaluationBySessionAndParticipantParams{
		SessionID:     sessionID,
		ParticipantID: participantID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrEvaluationNotFound
		}
		return nil, fmt.Errorf("evaluation_repo: get by session and participant: %w", err)
	}

	return dbEvaluationToEntity(row), nil
}

func (r *EvaluationRepo) GetScoresByEvaluation(ctx context.Context, evaluationID uuid.UUID) ([]*entity.RubricScore, error) {
	rows, err := r.queries.GetRubricScoresByEvaluation(ctx, evaluationID)
	if err != nil {
		return nil, fmt.Errorf("evaluation_repo: get scores: %w", err)
	}

	scores := make([]*entity.RubricScore, len(rows))
	for i, row := range rows {
		var quotes []string
		_ = json.Unmarshal(row.EvidenceQuotes, &quotes)
		scores[i] = &entity.RubricScore{
			ID:             row.ID,
			EvaluationID:   row.EvaluationID,
			Dimension:      entity.RubricDimension(row.Dimension),
			Score:          row.Score,
			Weight:         row.Weight,
			Justification:  row.Justification,
			EvidenceQuotes: quotes,
		}
	}
	return scores, nil
}

func (r *EvaluationRepo) Exists(ctx context.Context, sessionID, participantID uuid.UUID) (bool, error) {
	exists, err := r.queries.EvaluationExists(ctx, db.EvaluationExistsParams{
		SessionID:     sessionID,
		ParticipantID: participantID,
	})
	if err != nil {
		return false, fmt.Errorf("evaluation_repo: exists: %w", err)
	}
	return exists, nil
}

func dbEvaluationToEntity(row db.Evaluation) *entity.Evaluation {
	var strengths, areas []string
	_ = json.Unmarshal(row.Strengths, &strengths)
	_ = json.Unmarshal(row.AreasForImprovement, &areas)

	var score, eloDelta, eloBefore, eloAfter float64
	if row.OverallScore != nil {
		score = *row.OverallScore
	}
	if row.EloDelta != nil {
		eloDelta = *row.EloDelta
	}
	if row.EloRatingBefore != nil {
		eloBefore = *row.EloRatingBefore
	}
	if row.EloRatingAfter != nil {
		eloAfter = *row.EloRatingAfter
	}

	var reason, feedback, model, stt, version string
	if row.ObjectiveAchievementReasoning != nil {
		reason = *row.ObjectiveAchievementReasoning
	}
	if row.SummaryFeedback != nil {
		feedback = *row.SummaryFeedback
	}
	if row.LlmModelUsed != nil {
		model = *row.LlmModelUsed
	}
	if row.SttProvider != nil {
		stt = *row.SttProvider
	}
	if row.PromptVersion != nil {
		version = *row.PromptVersion
	}

	return &entity.Evaluation{
		ID:                         row.ID,
		SessionID:                  row.SessionID,
		ParticipantID:              row.ParticipantID,
		OverallScore:               score,
		ObjectiveAchieved:          row.ObjectiveAchieved,
		ObjectiveAchievementReason: reason,
		SummaryFeedback:            feedback,
		Strengths:                  strengths,
		AreasForImprovement:        areas,
		EloDelta:                   eloDelta,
		EloRatingBefore:            eloBefore,
		EloRatingAfter:             eloAfter,
		LLMModelUsed:               model,
		STTProvider:                stt,
		PromptVersion:              version,
		IsFallback:                 row.IsFallback,
		CreatedAt:                  row.CreatedAt,
	}
}