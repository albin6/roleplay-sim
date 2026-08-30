package entity

import (
	"time"

	"github.com/google/uuid"
)

// SessionState represents the lifecycle state of a session.
type SessionState string

const (
	SessionStateIdle        SessionState = "idle"
	SessionStateWaiting     SessionState = "waiting"
	SessionStateReady       SessionState = "ready"
	SessionStateSpinning    SessionState = "spinning"
	SessionStateScenario    SessionState = "scenario"
	SessionStatePrep        SessionState = "prep"
	SessionStateSignaling   SessionState = "signaling"
	SessionStateLive        SessionState = "live"
	SessionStateEvaluating  SessionState = "evaluating"
	SessionStateComplete    SessionState = "complete"
	SessionStateClosed      SessionState = "closed"
)

// Session represents a completed or in-progress roleplay session stored in PostgreSQL.
type Session struct {
	ID            uuid.UUID
	RoomID        string
	ScenarioID    uuid.UUID
	Difficulty    Difficulty
	State         SessionState
	SpinSeed      *int64
	StartedAt     *time.Time
	EndedAt       *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// SessionParticipant links a user to a session with their assigned role.
type SessionParticipant struct {
	ID        uuid.UUID
	SessionID uuid.UUID
	UserID    uuid.UUID
	RoleID    uuid.UUID
	Seat      string // "A" or "B"
	JoinedAt  time.Time
}
