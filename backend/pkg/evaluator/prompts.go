package evaluator

import (
	"fmt"
	"strings"
)

const PromptVersion = "v1.0"

const SystemPrompt = `You are an expert workplace communication coach and trained evaluator for professional roleplay simulations.

Your role is to evaluate a participant's performance in a timed workplace roleplay conversation based strictly on the provided rubric dimensions. You have deep expertise in negotiation theory, conflict resolution, organizational psychology, and professional communication.

EVALUATION PRINCIPLES:
1. Be objective and evidence-based. Every score MUST be grounded in specific quotes from the transcript.
2. Be calibrated. A score of 10/10 should be rare and reserved for exemplary, professional-grade performance.
3. Be constructive. Feedback should be actionable and specific, not generic.
4. Stay in scope. Only evaluate the evaluated participant's performance, not the peer's.
5. Follow the JSON schema exactly. Do not include any markdown formatting or commentary outside the JSON object.

SCORING SCALE (per dimension, 0-10):
0-2:  Poor — fundamental failure in this area
3-4:  Below average — significant gaps, inconsistent execution
5-6:  Average — competent but unremarkable; common mistakes present
7-8:  Good — above average; minor issues only
9-10: Excellent — professional-grade, highly effective execution`

type UserPromptParams struct {
	ScenarioTitle        string
	Difficulty           string
	BackgroundContext    string
	EvaluatedRole        string
	EvaluatedLevel       int
	EvaluatedObjective   string
	EvaluatedConstraints []string
	PeerRole             string
	PeerLevel            int
	PeerObjective        string
	Transcript           string
	SessionSeconds       int
}

func BuildUserPrompt(p UserPromptParams) string {
	constraintsStr := "None specified"
	if len(p.EvaluatedConstraints) > 0 {
		constraintsStr = strings.Join(p.EvaluatedConstraints, "; ")
	}

	return fmt.Sprintf(`SCENARIO CONTEXT:
Title: %s
Difficulty: %s
Background: %s

EVALUATED PARTICIPANT:
Role: %s (Hierarchy Level %d/4)
Private Objective: %s
Constraints: %s

PEER PARTICIPANT:
Role: %s (Hierarchy Level %d/4)
Peer's Objective (for context): %s

SESSION TRANSCRIPT (%ds duration):
%s

RUBRIC DIMENSIONS TO EVALUATE:
- communication_clarity (weight: 0.20): Structured, concise, professional phrasing vs rambling, vague language.
- active_listening (weight: 0.15): Paraphrasing, follow-up questions, acknowledging counterpart constraints vs ignoring points.
- negotiation_strategy (weight: 0.20): Interest-based framing, compromise proposals, anchoring vs stubborn positional bargaining.
- emotional_regulation (weight: 0.15): Calm under pressure, polite, professional tone vs defensiveness or frustration.
- empathy (weight: 0.10): Validates counterpart constraints, perspective-taking vs dismissiveness.
- objective_alignment (weight: 0.20): Measurable strategic progress toward the assigned private objective.

EVALUATION TASK:
Evaluate ONLY the [%s] participant.
Produce a single JSON object containing:
- overall_score (0-100)
- objective_achieved (boolean)
- objective_achievement_reasoning (string, 50-500 chars)
- summary_feedback (string, 100-600 chars)
- strengths (array of strings, 1-5 items)
- areas_for_improvement (array of strings, 1-5 items)
- rubric_scores (array of 6 objects, each with dimension, score (0-10), weight, justification, and evidence_quotes)
`,
		p.ScenarioTitle,
		p.Difficulty,
		p.BackgroundContext,
		p.EvaluatedRole,
		p.EvaluatedLevel,
		p.EvaluatedObjective,
		constraintsStr,
		p.PeerRole,
		p.PeerLevel,
		p.PeerObjective,
		p.SessionSeconds,
		p.Transcript,
		p.EvaluatedRole,
	)
}