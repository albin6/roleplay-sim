package elo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculate_BasicWin(t *testing.T) {
	in := Input{
		RatingA:    1200.0,
		RatingB:    1200.0,
		SessionsA:  5,
		SessionsB:  5,
		HierarchyA: 1,
		HierarchyB: 3,
		ScoreA:     85.0,
		ScoreB:     40.0,
		AchievedA:  true,
		AchievedB:  false,
		Difficulty: "medium",
	}

	res := Calculate(in)

	assert.True(t, res.DeltaA > 0, "Winner should gain Elo")
	assert.True(t, res.DeltaB < 0, "Loser should lose Elo")
	assert.Equal(t, res.NewRatingA, in.RatingA+res.DeltaA)
	assert.Equal(t, res.NewRatingB, in.RatingB+res.DeltaB)
}

func TestCalculate_MinimumFloor(t *testing.T) {
	in := Input{
		RatingA:    810.0,
		RatingB:    1600.0,
		SessionsA:  2,
		SessionsB:  50,
		HierarchyA: 1,
		HierarchyB: 4,
		ScoreA:     10.0,
		ScoreB:     95.0,
		AchievedA:  false,
		AchievedB:  true,
		Difficulty: "hard",
	}

	res := Calculate(in)

	assert.GreaterOrEqual(t, res.NewRatingA, MinEloRating, "Rating must not drop below minimum floor")
}

func TestCalculate_DifficultyScaling(t *testing.T) {
	baseInput := Input{
		RatingA:    1200.0,
		RatingB:    1200.0,
		SessionsA:  20,
		SessionsB:  20,
		HierarchyA: 2,
		HierarchyB: 2,
		ScoreA:     80.0,
		ScoreB:     30.0,
		AchievedA:  true,
		AchievedB:  false,
	}

	easyInput := baseInput
	easyInput.Difficulty = "easy"
	resEasy := Calculate(easyInput)

	hardInput := baseInput
	hardInput.Difficulty = "hard"
	resHard := Calculate(hardInput)

	assert.True(t, resHard.DeltaA > resEasy.DeltaA, "Hard difficulty should have higher K-factor and larger delta")
}

func TestCalculate_Draw(t *testing.T) {
	in := Input{
		RatingA:    1200.0,
		RatingB:    1200.0,
		SessionsA:  15,
		SessionsB:  15,
		HierarchyA: 2,
		HierarchyB: 2,
		ScoreA:     50.0,
		ScoreB:     50.0,
		AchievedA:  true,
		AchievedB:  true,
		Difficulty: "medium",
	}

	res := Calculate(in)

	assert.InDelta(t, 0.0, res.DeltaA, 0.01, "Equal draw should yield near-zero delta")
	assert.InDelta(t, 0.0, res.DeltaB, 0.01, "Equal draw should yield near-zero delta")
}