package entity

import (
	"time"

	"github.com/google/uuid"
)

// RubricDimension identifies a scoring rubric axis.
type RubricDimension string

const (
	DimCommunicationClarity  RubricDimension = "communication_clarity"
	DimActiveListening       RubricDimension = "active_listening"
	DimNegotiationStrategy   RubricDimension = "negotiation_strategy"
	DimEmotionalRegulation   RubricDimension = "emotional_regulation"
	DimEmpathy               RubricDimension = "empathy"
	DimObjectiveAlignment    RubricDimension = "objective_alignment"
)

// RubricWeight maps each dimension to its scoring weight.
var RubricWeights = map[RubricDimension]float64{
	DimCommunicationClarity: 0.20,
	DimActiveListening:      0.15,
	DimNegotiationStrategy:  0.20,
	DimEmotionalRegulation:  0.15,
	DimEmpathy:              0.10,
	DimObjectiveAlignment:   0.20,
}

// Evaluation holds the LLM-generated rubric result for one participant.
type Evaluation struct {
	ID                          uuid.UUID
	SessionID                   uuid.UUID
	ParticipantID               uuid.UUID
	OverallScore                float64
	ObjectiveAchieved           *bool
	ObjectiveAchievementReason  string
	SummaryFeedback             string
	Strengths                   []string
	AreasForImprovement         []string
	EloDelta                    float64
	EloRatingBefore             float64
	EloRatingAfter              float64
	LLMModelUsed                string
	STTProvider                 string
	PromptVersion               string
	IsFallback                  bool
	CreatedAt                   time.Time
}

// RubricScore is a per-dimension score within an evaluation.
type RubricScore struct {
	ID              uuid.UUID
	EvaluationID    uuid.UUID
	Dimension       RubricDimension
	Score           float64
	Weight          float64
	Justification   string
	EvidenceQuotes  []string
}
