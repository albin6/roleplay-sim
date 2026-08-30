package entity

import "time"

// RoomState mirrors SessionState but lives in Redis for real-time access.
type RoomState string

const (
	RoomStateWaiting    RoomState = "waiting"
	RoomStateReady      RoomState = "ready"
	RoomStateSpinning   RoomState = "spinning"
	RoomStateScenario   RoomState = "scenario"
	RoomStatePrep       RoomState = "prep"
	RoomStateSignaling  RoomState = "signaling"
	RoomStateLive       RoomState = "live"
	RoomStateEvaluating RoomState = "evaluating"
	RoomStateComplete   RoomState = "complete"
)

// Peer holds per-user data within a room.
type Peer struct {
	UserID      string
	DisplayName string
	AvatarURL   string
	EloRating   float64
	Seat        string // "A" or "B"
	Connected   bool
}

// Room is the in-memory (Redis) state for an active session room.
type Room struct {
	ID          string
	State       RoomState
	PeerA       Peer
	PeerB       Peer
	ScenarioID  string
	SpinSeed    int64
	Difficulty  Difficulty
	CreatedAt   time.Time
	ExpiresAt   time.Time
}
