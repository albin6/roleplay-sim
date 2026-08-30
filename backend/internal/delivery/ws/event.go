package ws

import (
	"encoding/json"
	"time"
)

type EventType string

const (
	// Server -> Client
	EventConnected        EventType = "CONNECTED"
	EventRoomReady        EventType = "ROOM_READY"
	EventSpinStart        EventType = "SPIN_START"
	EventSpinResult       EventType = "SPIN_RESULT"
	EventScenarioAssign   EventType = "SCENARIO_ASSIGN"
	EventPrepTimerTick    EventType = "PREP_TIMER_TICK"
	EventPrepEnd          EventType = "PREP_END"
	EventSessionTimerTick EventType = "SESSION_TIMER_TICK"
	EventSessionComplete  EventType = "SESSION_COMPLETE"
	EventEvaluationReady  EventType = "EVALUATION_READY"
	EventPeerDisconnected EventType = "PEER_DISCONNECTED"
	EventRoomClosed       EventType = "ROOM_CLOSED"
	EventPong             EventType = "PONG"
	EventError            EventType = "ERROR"

	// Client -> Server
	EventJoinQueue   EventType = "JOIN_QUEUE"
	EventLeaveQueue  EventType = "LEAVE_QUEUE"
	EventSpinAck     EventType = "SPIN_ACK"
	EventScenarioAck EventType = "SCENARIO_ACK"
	EventPrepReady   EventType = "PREP_READY"
	EventSessionEnd  EventType = "SESSION_END"
	EventPing        EventType = "PING"

	// Bidirectional
	EventSignal EventType = "SIGNAL"
)

// Envelope is the standard framing for all JSON WebSocket messages.
type Envelope struct {
	Event     EventType       `json:"event"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp string          `json:"timestamp"`
	Seq       int64           `json:"seq"`
}

// NewEnvelope constructs a serialized Envelope ready for sending.
func NewEnvelope(event EventType, payload any, seq int64) ([]byte, error) {
	var raw json.RawMessage
	if payload != nil {
		bytes, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		raw = bytes
	} else {
		raw = json.RawMessage(`{}`)
	}

	env := Envelope{
		Event:     event,
		Payload:   raw,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Seq:       seq,
	}

	return json.Marshal(env)
}

// --- Server Payload Structs ---

type ConnectedPayload struct {
	ConnectionID string `json:"connection_id"`
	ServerTime   string `json:"server_time"`
}

type RoomReadyPayload struct {
	RoomID          string  `json:"room_id"`
	PeerDisplayName string  `json:"peer_display_name"`
	PeerAvatarURL   string  `json:"peer_avatar_url"`
	PeerEloRating   float64 `json:"peer_elo_rating"`
	Seat            string  `json:"seat"` // "A" or "B"
}

type SpinStartPayload struct {
	RoomID              string `json:"room_id"`
	SpinSeed            int64  `json:"spin_seed"`
	AnimationDurationMs int    `json:"animation_duration_ms"`
}

type RoleInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	HierarchyLevel int    `json:"hierarchy_level"`
	Description    string `json:"description"`
}

type ContextInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type SpinResultPayload struct {
	RoomID     string      `json:"room_id"`
	Context    ContextInfo `json:"context"`
	YourRole   RoleInfo    `json:"your_role"`
	PeerRole   RoleInfo    `json:"peer_role"`
	Difficulty string      `json:"difficulty"`
}

type ScenarioAssignPayload struct {
	RoomID                 string   `json:"room_id"`
	ScenarioID             string   `json:"scenario_id"`
	Title                  string   `json:"title"`
	Difficulty             string   `json:"difficulty"`
	BackgroundContext      string   `json:"background_context"`
	YourObjective          string   `json:"your_objective"`
	YourConstraints        []string `json:"your_constraints"`
	PrepDurationSeconds    int      `json:"prep_duration_seconds"`
	SessionDurationSeconds int      `json:"session_duration_seconds"`
}

type PrepTimerTickPayload struct {
	RoomID           string `json:"room_id"`
	SecondsRemaining int    `json:"seconds_remaining"`
	PeerReady        bool   `json:"peer_ready"`
}

type PrepEndPayload struct {
	RoomID        string `json:"room_id"`
	InitiatorSeat string `json:"initiator_seat"` // seat that makes offer
}

type SessionTimerTickPayload struct {
	RoomID           string `json:"room_id"`
	SecondsRemaining int    `json:"seconds_remaining"`
	Phase            string `json:"phase"`
}

type PeerDisconnectedPayload struct {
	RoomID                 string `json:"room_id"`
	ReconnectWindowSeconds int    `json:"reconnect_window_seconds"`
}

type RoomClosedPayload struct {
	RoomID string `json:"room_id"`
	Reason string `json:"reason"`
}

type ErrorPayload struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

type PongPayload struct {
	ServerTime    string `json:"server_time"`
	LatencyHintMs int64  `json:"latency_hint_ms"`
}

type RubricScoreItem struct {
	Dimension     string  `json:"dimension"`
	Score         float64 `json:"score"`
	Weight        float64 `json:"weight"`
	Justification string  `json:"justification"`
}

type EvaluationScorePayload struct {
	OverallScore        float64           `json:"overall_score"`
	ObjectiveAchieved   bool              `json:"objective_achieved"`
	EloDelta            float64           `json:"elo_delta"`
	EloNew              float64           `json:"elo_new"`
	SummaryFeedback     string            `json:"summary_feedback"`
	Strengths           []string          `json:"strengths"`
	AreasForImprovement []string          `json:"areas_for_improvement"`
	RubricScores        []RubricScoreItem `json:"rubric_scores"`
}

type EvaluationPeerScorePayload struct {
	OverallScore      float64 `json:"overall_score"`
	ObjectiveAchieved bool    `json:"objective_achieved"`
	EloDelta          float64 `json:"elo_delta"`
}

type EvaluationReadyPayload struct {
	RoomID    string                     `json:"room_id"`
	SessionID string                     `json:"session_id"`
	YourScore EvaluationScorePayload     `json:"your_score"`
	PeerScore EvaluationPeerScorePayload `json:"peer_score"`
}

// --- Client Payload Structs ---

type RoomActionPayload struct {
	RoomID string `json:"room_id"`
}

type SessionEndPayload struct {
	RoomID string `json:"room_id"`
	Reason string `json:"reason"` // "timer_expired" | "user_ended"
}

type PingPayload struct {
	ClientTime string `json:"client_time"`
}

type SignalEnvelopePayload struct {
	RoomID string          `json:"room_id"`
	Signal json.RawMessage `json:"signal"`
}