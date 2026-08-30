package entity

import (
	"time"

	"github.com/google/uuid"
)

// Difficulty represents the scenario difficulty tier.
type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

// Scenario represents a roleplay scenario template.
type Scenario struct {
	ID                     uuid.UUID
	ContextID              uuid.UUID
	Title                  string
	Difficulty             Difficulty
	BackgroundContext      string
	RoleAObjective         string
	RoleAConstraints       []string
	RoleBObjective         string
	RoleBConstraints       []string
	PrepDurationSeconds    int
	SessionDurationSeconds int
	IsActive               bool
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
