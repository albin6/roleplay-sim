package elo

import (
	"math"
)

const (
	KBase         = 32.0
	MinEloRating  = 800.0
	StartingElo   = 1200.0
)

// Input holds all data needed to calculate Elo deltas for one match.
type Input struct {
	RatingA       float64
	RatingB       float64
	SessionsA     int
	SessionsB     int
	HierarchyA    int // 1-4
	HierarchyB    int // 1-4
	ScoreA        float64 // 0-100
	ScoreB        float64 // 0-100
	AchievedA     bool
	AchievedB     bool
	Difficulty    string // "easy", "medium", "hard"
}

// Result holds the Elo deltas.
type Result struct {
	NewRatingA float64
	NewRatingB float64
	DeltaA     float64
	DeltaB     float64
}

// Calculate computes Elo rating changes per the project spec.
func Calculate(in Input) Result {
	// Expected scores
	ea := 1.0 / (1.0 + math.Pow(10, (in.RatingB-in.RatingA)/400.0))
	eb := 1.0 - ea

	// Raw outcomes
	sa, sb := rawOutcomes(in.AchievedA, in.AchievedB)

	// Score-weighted modifier: (score/100 - 0.5) * 0.2, clamped ±0.10
	sa = clamp(sa+scoreBonus(in.ScoreA), 0.0, 1.0)
	sb = clamp(sb+scoreBonus(in.ScoreB), 0.0, 1.0)

	// K-factors
	kDiff := kDifficulty(in.Difficulty)
	hierarchyDiff := abs(in.HierarchyA - in.HierarchyB)
	kHier := kHierarchy(hierarchyDiff)

	// Apply hierarchy bonus only to the lower-ranked participant
	var ka, kb float64
	if in.HierarchyA <= in.HierarchyB {
		ka = KBase * kDiff * kHier * kExperience(in.SessionsA)
		kb = KBase * kDiff * 1.0 * kExperience(in.SessionsB)
	} else {
		ka = KBase * kDiff * 1.0 * kExperience(in.SessionsA)
		kb = KBase * kDiff * kHier * kExperience(in.SessionsB)
	}

	deltaA := ka * (sa - ea)
	deltaB := kb * (sb - eb)

	newA := math.Max(MinEloRating, in.RatingA+deltaA)
	newB := math.Max(MinEloRating, in.RatingB+deltaB)

	return Result{
		NewRatingA: newA,
		NewRatingB: newB,
		DeltaA:     newA - in.RatingA,
		DeltaB:     newB - in.RatingB,
	}
}

func rawOutcomes(achievedA, achievedB bool) (float64, float64) {
	switch {
	case achievedA && !achievedB:
		return 1.0, 0.0
	case !achievedA && achievedB:
		return 0.0, 1.0
	default: // both or neither
		return 0.5, 0.5
	}
}

func scoreBonus(score float64) float64 {
	return clamp((score/100.0-0.5)*0.2, -0.10, 0.10)
}

func kDifficulty(d string) float64 {
	switch d {
	case "easy":
		return 0.75
	case "hard":
		return 1.25
	default: // medium
		return 1.0
	}
}

func kHierarchy(diff int) float64 {
	base := 1.0 + float64(diff)*0.05
	if diff > 3 {
		base = 1.15
	}
	return base
}

func kExperience(sessions int) float64 {
	switch {
	case sessions < 10:
		return 1.50
	case sessions <= 50:
		return 1.00
	default:
		return 0.80
	}
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
