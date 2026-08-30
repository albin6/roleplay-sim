package evaluation

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/roleplay-sim/backend/internal/domain/entity"
	"github.com/roleplay-sim/backend/internal/domain/repository"
	"github.com/roleplay-sim/backend/pkg/elo"
	"github.com/roleplay-sim/backend/pkg/evaluator"
	"github.com/roleplay-sim/backend/pkg/stt"
	"github.com/rs/zerolog/log"
)

type SessionEvaluationResult struct {
	SessionID uuid.UUID
	RoomID    string

	// Results for Participant A
	ScoreA          *evaluator.EvaluationResponse
	ParticipantIDA  uuid.UUID
	EloDeltaA       float64
	EloNewA         float64

	// Results for Participant B
	ScoreB          *evaluator.EvaluationResponse
	ParticipantIDB  uuid.UUID
	EloDeltaB       float64
	EloNewB         float64

	Transcript *stt.TranscriptResult
}

type EvaluateSessionUseCase struct {
	audioBuffer  repository.AudioBufferRepository
	sttProvider  stt.STTProvider
	evaluator    *evaluator.EvaluatorService
	userRepo     repository.UserRepository
	sessionRepo  repository.SessionRepository
	evalRepo     repository.EvaluationRepository
}

func NewEvaluateSessionUseCase(
	audioBuffer repository.AudioBufferRepository,
	sttProvider stt.STTProvider,
	evaluator *evaluator.EvaluatorService,
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	evalRepo repository.EvaluationRepository,
) *EvaluateSessionUseCase {
	return &EvaluateSessionUseCase{
		audioBuffer: audioBuffer,
		sttProvider: sttProvider,
		evaluator:   evaluator,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		evalRepo:    evalRepo,
	}
}

func (uc *EvaluateSessionUseCase) Execute(
	ctx context.Context,
	sessionID uuid.UUID,
	roomID string,
	userA, userB *entity.User,
	scenario *entity.Scenario,
	roleA, roleB entity.Role,
) (*SessionEvaluationResult, error) {
	log.Info().Str("room_id", roomID).Str("session_id", sessionID.String()).Msg("eval_pipeline: starting evaluation worker")

	// 1. Fetch participants to map Participant IDs
	participants, err := uc.sessionRepo.GetParticipants(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("eval_pipeline: get participants: %w", err)
	}

	var partIDA, partIDB uuid.UUID
	for _, p := range participants {
		if p.Seat == "A" {
			partIDA = p.ID
		} else if p.Seat == "B" {
			partIDB = p.ID
		}
	}

	// 2. Fetch audio chunks from Redis Streams
	var audioA, audioB []byte
	if uc.audioBuffer != nil {
		chunksA, errA := uc.audioBuffer.GetChunks(ctx, roomID, userA.ID.String())
		if errA == nil && len(chunksA) > 0 {
			audioA = bytes.Join(chunksA, nil)
		}
		chunksB, errB := uc.audioBuffer.GetChunks(ctx, roomID, userB.ID.String())
		if errB == nil && len(chunksB) > 0 {
			audioB = bytes.Join(chunksB, nil)
		}
	}

	// 3. Transcribe audio via STT Provider
	transcript, err := uc.sttProvider.Transcribe(ctx, audioA, audioB, roleA.Name, roleB.Name)
	if err != nil {
		log.Warn().Err(err).Msg("eval_pipeline: STT failed, falling back to mock dialogue")
		transcript, _ = stt.NewMockSTT().Transcribe(ctx, nil, nil, roleA.Name, roleB.Name)
	}

	// 4. Parallel LLM Evaluation for User A and User B
	var (
		evalRespA *evaluator.EvaluationResponse
		evalRespB *evaluator.EvaluationResponse
		errA, errB error
		wg         sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		evalRespA, errA = uc.evaluator.Evaluate(ctx, evaluator.UserPromptParams{
			ScenarioTitle:        scenario.Title,
			Difficulty:           string(scenario.Difficulty),
			BackgroundContext:    scenario.BackgroundContext,
			EvaluatedRole:        roleA.Name,
			EvaluatedLevel:       roleA.HierarchyLevel,
			EvaluatedObjective:   scenario.RoleAObjective,
			EvaluatedConstraints: scenario.RoleAConstraints,
			PeerRole:             roleB.Name,
			PeerLevel:            roleB.HierarchyLevel,
			PeerObjective:        scenario.RoleBObjective,
			Transcript:           transcript.FullInterleaved,
			SessionSeconds:       scenario.SessionDurationSeconds,
		})
	}()

	go func() {
		defer wg.Done()
		evalRespB, errB = uc.evaluator.Evaluate(ctx, evaluator.UserPromptParams{
			ScenarioTitle:        scenario.Title,
			Difficulty:           string(scenario.Difficulty),
			BackgroundContext:    scenario.BackgroundContext,
			EvaluatedRole:        roleB.Name,
			EvaluatedLevel:       roleB.HierarchyLevel,
			EvaluatedObjective:   scenario.RoleBObjective,
			EvaluatedConstraints: scenario.RoleBConstraints,
			PeerRole:             roleA.Name,
			PeerLevel:            roleA.HierarchyLevel,
			PeerObjective:        scenario.RoleAObjective,
			Transcript:           transcript.FullInterleaved,
			SessionSeconds:       scenario.SessionDurationSeconds,
		})
	}()

	wg.Wait()

	if errA != nil {
		return nil, fmt.Errorf("eval_pipeline: evaluate participant A: %w", errA)
	}
	if errB != nil {
		return nil, fmt.Errorf("eval_pipeline: evaluate participant B: %w", errB)
	}

	// 5. Compute Elo deltas
	achievedA := evalRespA.ObjectiveAchieved != nil && *evalRespA.ObjectiveAchieved
	achievedB := evalRespB.ObjectiveAchieved != nil && *evalRespB.ObjectiveAchieved

	eloResult := elo.Calculate(elo.Input{
		RatingA:    userA.EloRating,
		RatingB:    userB.EloRating,
		SessionsA:  userA.TotalSessions,
		SessionsB:  userB.TotalSessions,
		HierarchyA: roleA.HierarchyLevel,
		HierarchyB: roleB.HierarchyLevel,
		ScoreA:     evalRespA.OverallScore,
		ScoreB:     evalRespB.OverallScore,
		AchievedA:  achievedA,
		AchievedB:  achievedB,
		Difficulty: string(scenario.Difficulty),
	})

	// If fallback, zero out Elo delta
	if evalRespA.IsFallback || evalRespB.IsFallback {
		eloResult.DeltaA = 0
		eloResult.DeltaB = 0
		eloResult.NewRatingA = userA.EloRating
		eloResult.NewRatingB = userB.EloRating
	}

	// 6. Persist Evaluation for Participant A
	rubricScoresA := make([]*entity.RubricScore, len(evalRespA.RubricScores))
	for i, r := range evalRespA.RubricScores {
		rubricScoresA[i] = &entity.RubricScore{
			Dimension:      entity.RubricDimension(r.Dimension),
			Score:          r.Score,
			Weight:         r.Weight,
			Justification:  r.Justification,
			EvidenceQuotes: r.EvidenceQuotes,
		}
	}

	evalEntityA := &entity.Evaluation{
		SessionID:                  sessionID,
		ParticipantID:              partIDA,
		OverallScore:               evalRespA.OverallScore,
		ObjectiveAchieved:          evalRespA.ObjectiveAchieved,
		ObjectiveAchievementReason: evalRespA.ObjectiveAchievementReason,
		SummaryFeedback:            evalRespA.SummaryFeedback,
		Strengths:                  evalRespA.Strengths,
		AreasForImprovement:        evalRespA.AreasForImprovement,
		EloDelta:                   eloResult.DeltaA,
		EloRatingBefore:            userA.EloRating,
		EloRatingAfter:             eloResult.NewRatingA,
		LLMModelUsed:               evalRespA.ModelUsed,
		STTProvider:                transcript.Provider,
		PromptVersion:              evaluator.PromptVersion,
		IsFallback:                 evalRespA.IsFallback,
		CreatedAt:                  time.Now().UTC(),
	}
	_ = uc.evalRepo.Create(ctx, evalEntityA, rubricScoresA)

	// 7. Persist Evaluation for Participant B
	rubricScoresB := make([]*entity.RubricScore, len(evalRespB.RubricScores))
	for i, r := range evalRespB.RubricScores {
		rubricScoresB[i] = &entity.RubricScore{
			Dimension:      entity.RubricDimension(r.Dimension),
			Score:          r.Score,
			Weight:         r.Weight,
			Justification:  r.Justification,
			EvidenceQuotes: r.EvidenceQuotes,
		}
	}

	evalEntityB := &entity.Evaluation{
		SessionID:                  sessionID,
		ParticipantID:              partIDB,
		OverallScore:               evalRespB.OverallScore,
		ObjectiveAchieved:          evalRespB.ObjectiveAchieved,
		ObjectiveAchievementReason: evalRespB.ObjectiveAchievementReason,
		SummaryFeedback:            evalRespB.SummaryFeedback,
		Strengths:                  evalRespB.Strengths,
		AreasForImprovement:        evalRespB.AreasForImprovement,
		EloDelta:                   eloResult.DeltaB,
		EloRatingBefore:            userB.EloRating,
		EloRatingAfter:             eloResult.NewRatingB,
		LLMModelUsed:               evalRespB.ModelUsed,
		STTProvider:                transcript.Provider,
		PromptVersion:              evaluator.PromptVersion,
		IsFallback:                 evalRespB.IsFallback,
		CreatedAt:                  time.Now().UTC(),
	}
	_ = uc.evalRepo.Create(ctx, evalEntityB, rubricScoresB)

	// 8. Update User Elo ratings and metrics in PostgreSQL
	_ = uc.userRepo.UpdateEloRating(ctx, userA.ID, eloResult.NewRatingA)
	_ = uc.userRepo.IncrementSessionCount(ctx, userA.ID, achievedA)

	_ = uc.userRepo.UpdateEloRating(ctx, userB.ID, eloResult.NewRatingB)
	_ = uc.userRepo.IncrementSessionCount(ctx, userB.ID, achievedB)

	// 9. Update Session state to complete
	_ = uc.sessionRepo.UpdateState(ctx, sessionID, entity.SessionStateComplete)

	// 10. Clean up Redis Streams
	if uc.audioBuffer != nil {
		_ = uc.audioBuffer.ClearBuffer(ctx, roomID, userA.ID.String())
		_ = uc.audioBuffer.ClearBuffer(ctx, roomID, userB.ID.String())
	}

	log.Info().Str("room_id", roomID).
		Float64("scoreA", evalRespA.OverallScore).Float64("deltaA", eloResult.DeltaA).
		Float64("scoreB", evalRespB.OverallScore).Float64("deltaB", eloResult.DeltaB).
		Msg("eval_pipeline: evaluation complete and persisted")

	return &SessionEvaluationResult{
		SessionID:      sessionID,
		RoomID:         roomID,
		ScoreA:         evalRespA,
		ParticipantIDA: partIDA,
		EloDeltaA:      eloResult.DeltaA,
		EloNewA:        eloResult.NewRatingA,
		ScoreB:         evalRespB,
		ParticipantIDB: partIDB,
		EloDeltaB:      eloResult.DeltaB,
		EloNewB:        eloResult.NewRatingB,
		Transcript:     transcript,
	}, nil
}