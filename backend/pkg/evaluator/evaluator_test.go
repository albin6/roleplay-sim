package evaluator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildUserPrompt(t *testing.T) {
	params := UserPromptParams{
		ScenarioTitle:        "Leave Request Under Deadline Pressure",
		Difficulty:           "medium",
		BackgroundContext:    "Sprint release is 3 weeks away.",
		EvaluatedRole:        "Junior Developer",
		EvaluatedLevel:       1,
		EvaluatedObjective:   "Secure 2 days leave.",
		EvaluatedConstraints: []string{"Do not reveal personal emergency", "Stay professional"},
		PeerRole:             "Team Lead",
		PeerLevel:            3,
		PeerObjective:        "Limit leave to 1 day.",
		Transcript:           "[00:01] [Junior Developer]: Hello\n[00:04] [Team Lead]: Hi",
		SessionSeconds:       360,
	}

	prompt := BuildUserPrompt(params)
	assert.Contains(t, prompt, "Leave Request Under Deadline Pressure")
	assert.Contains(t, prompt, "Junior Developer")
	assert.Contains(t, prompt, "Secure 2 days leave.")
	assert.Contains(t, prompt, "communication_clarity")
}

func TestEvaluatorService_Fallback(t *testing.T) {
	svc := NewEvaluatorService("") // Empty key triggers fallback
	params := UserPromptParams{
		ScenarioTitle:  "Leave Request",
		Difficulty:     "medium",
		EvaluatedRole:  "Junior Developer",
		EvaluatedLevel: 1,
	}

	res, err := svc.Evaluate(context.Background(), params)
	require.NoError(t, err)
	assert.True(t, res.IsFallback)
	assert.Equal(t, 6, len(res.RubricScores))
	assert.GreaterOrEqual(t, res.OverallScore, 70.0)
	assert.LessOrEqual(t, res.OverallScore, 100.0)
}