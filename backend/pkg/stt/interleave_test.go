package stt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterleaveDialogue(t *testing.T) {
	segsA := []DialogueSegment{
		{
			Speaker:   "Junior Developer",
			Text:      "I understand the release pressure, and I want to be transparent about my leave request.",
			StartSecs: 0.0,
			EndSecs:   3.2,
		},
		{
			Speaker:   "Junior Developer",
			Text:      "I need two days off next week, Wednesday and Thursday.",
			StartSecs: 7.1,
			EndSecs:   10.5,
		},
	}

	segsB := []DialogueSegment{
		{
			Speaker:   "Team Lead",
			Text:      "I appreciate you coming to me early. What exactly do you need?",
			StartSecs: 3.5,
			EndSecs:   6.8,
		},
		{
			Speaker:   "Team Lead",
			Text:      "Two days is a big ask given our sprint commitments.",
			StartSecs: 11.0,
			EndSecs:   14.2,
		},
	}

	res := InterleaveDialogue(segsA, segsB)

	assert.Equal(t, 4, len(res.Segments))
	assert.Equal(t, "Junior Developer", res.Segments[0].Speaker)
	assert.Equal(t, "Team Lead", res.Segments[1].Speaker)
	assert.Equal(t, "Junior Developer", res.Segments[2].Speaker)
	assert.Equal(t, "Team Lead", res.Segments[3].Speaker)
	assert.Equal(t, 14.2, res.DurationSeconds)
	assert.Greater(t, res.TotalWords, 20)

	expectedStart := "[00:00] [Junior Developer]: I understand the release pressure"
	assert.Contains(t, res.FullInterleaved, expectedStart)
}