package ws

import (
	"context"
	"encoding/json"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/roleplay-sim/backend/internal/domain/entity"
	"github.com/roleplay-sim/backend/internal/domain/repository"
	"github.com/roleplay-sim/backend/pkg/webrtc"
	"github.com/rs/zerolog/log"
)

// RoomManager coordinates the real-time FSM for a single 2-player room.
type RoomManager struct {
	ID             string
	SessionID      uuid.UUID
	State          entity.RoomState
	UserA          *entity.User
	UserB          *entity.User
	ClientA        *Client
	ClientB        *Client
	Scenario       *entity.Scenario
	RoleA          entity.Role
	RoleB          entity.Role
	Context        entity.RoleContext
	Difficulty     entity.Difficulty
	SpinSeed       int64
	PrepReadyA     bool
	PrepReadyB     bool
	PrepSeconds    int
	SessionSeconds int
	timerCancel    context.CancelFunc
	mu             sync.Mutex

	sessionRepo  repository.SessionRepository
	scenarioRepo repository.ScenarioRepository
	roomState    repository.RoomStateRepository
	webrtcCache  repository.WebRTCCacheRepository
}

func NewRoomManager(
	roomID string,
	userA, userB *entity.User,
	difficulty entity.Difficulty,
	scenario *entity.Scenario,
	sessionRepo repository.SessionRepository,
	scenarioRepo repository.ScenarioRepository,
	roomState repository.RoomStateRepository,
	webrtcCache repository.WebRTCCacheRepository,
) *RoomManager {
	prepSecs := 180
	sessSecs := 360
	if scenario != nil {
		if scenario.PrepDurationSeconds > 0 {
			prepSecs = scenario.PrepDurationSeconds
		}
		if scenario.SessionDurationSeconds > 0 {
			sessSecs = scenario.SessionDurationSeconds
		}
	}

	return &RoomManager{
		ID:             roomID,
		State:          entity.RoomStateWaiting,
		UserA:          userA,
		UserB:          userB,
		Difficulty:     difficulty,
		Scenario:       scenario,
		PrepSeconds:    prepSecs,
		SessionSeconds: sessSecs,
		RoleA: entity.Role{
			ID:             uuid.New(),
			Name:           "Junior Developer",
			HierarchyLevel: 1,
			Description:    ptr("Entry-level engineer raising a workplace request"),
		},
		RoleB: entity.Role{
			ID:             uuid.New(),
			Name:           "Team Lead",
			HierarchyLevel: 3,
			Description:    ptr("Engineering manager balancing team deadlines and morale"),
		},
		Context: entity.RoleContext{
			ID:   uuid.New(),
			Name: "IT Team",
			Slug: "it-team",
		},
		sessionRepo:  sessionRepo,
		scenarioRepo: scenarioRepo,
		roomState:    roomState,
		webrtcCache:  webrtcCache,
	}
}

func ptr[T any](v T) *T { return &v }

// AttachClient binds an incoming WebSocket client to either Seat A or Seat B.
func (r *RoomManager) AttachClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	client.SetRoomID(r.ID)

	if client.UserID == r.UserA.ID.String() {
		r.ClientA = client
		log.Info().Str("room_id", r.ID).Str("user_id", client.UserID).Msg("ws: Client A attached to room")
	} else if client.UserID == r.UserB.ID.String() {
		r.ClientB = client
		log.Info().Str("room_id", r.ID).Str("user_id", client.UserID).Msg("ws: Client B attached to room")
	} else {
		log.Warn().Str("room_id", r.ID).Str("user_id", client.UserID).Msg("ws: unauthorized user attempted to join room")
		client.Send(EventError, ErrorPayload{
			Code:      "FORBIDDEN",
			Message:   "You are not a participant in this room",
			Retryable: false,
		})
		return
	}

	// Check if late-joining peer has cached WebRTC signals
	go r.flushCachedSignals(client)

	// If both connected and room is waiting, advance to Ready
	if r.ClientA != nil && r.ClientB != nil && r.State == entity.RoomStateWaiting {
		r.advanceToReadyLocked()
	}
}

func (r *RoomManager) advanceToReadyLocked() {
	r.State = entity.RoomStateReady
	_ = r.roomState.UpdateState(context.Background(), r.ID, entity.RoomStateReady)

	// Send ROOM_READY to Client A (includes Peer B details)
	avatarB := ""
	if r.UserB.AvatarURL != nil {
		avatarB = *r.UserB.AvatarURL
	}
	r.ClientA.Send(EventRoomReady, RoomReadyPayload{
		RoomID:          r.ID,
		PeerDisplayName: r.UserB.DisplayName,
		PeerAvatarURL:   avatarB,
		PeerEloRating:   r.UserB.EloRating,
		Seat:            "A",
	})

	// Send ROOM_READY to Client B (includes Peer A details)
	avatarA := ""
	if r.UserA.AvatarURL != nil {
		avatarA = *r.UserA.AvatarURL
	}
	r.ClientB.Send(EventRoomReady, RoomReadyPayload{
		RoomID:          r.ID,
		PeerDisplayName: r.UserA.DisplayName,
		PeerAvatarURL:   avatarA,
		PeerEloRating:   r.UserA.EloRating,
		Seat:            "B",
	})

	// Schedule Spin Phase after 1s
	go func() {
		time.Sleep(1 * time.Second)
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.State == entity.RoomStateReady {
			r.advanceToSpinningLocked()
		}
	}()
}

func (r *RoomManager) advanceToSpinningLocked() {
	r.State = entity.RoomStateSpinning
	r.SpinSeed = rand.Int63n(8999999) + 1000000
	_ = r.roomState.UpdateState(context.Background(), r.ID, entity.RoomStateSpinning)

	spinPayload := SpinStartPayload{
		RoomID:              r.ID,
		SpinSeed:            r.SpinSeed,
		AnimationDurationMs: 3500,
	}

	if r.ClientA != nil {
		r.ClientA.Send(EventSpinStart, spinPayload)
	}
	if r.ClientB != nil {
		r.ClientB.Send(EventSpinStart, spinPayload)
	}

	// Transition to Scenario after 3500ms spin animation
	go func() {
		time.Sleep(3600 * time.Millisecond)
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.State == entity.RoomStateSpinning {
			r.advanceToScenarioLocked()
		}
	}()
}

func (r *RoomManager) advanceToScenarioLocked() {
	r.State = entity.RoomStateScenario
	_ = r.roomState.UpdateState(context.Background(), r.ID, entity.RoomStateScenario)

	ctxInfo := ContextInfo{
		ID:   r.Context.ID.String(),
		Name: r.Context.Name,
		Slug: r.Context.Slug,
	}

	// SPIN_RESULT for A
	if r.ClientA != nil {
		r.ClientA.Send(EventSpinResult, SpinResultPayload{
			RoomID:     r.ID,
			Context:    ctxInfo,
			YourRole:   RoleInfo{ID: r.RoleA.ID.String(), Name: r.RoleA.Name, HierarchyLevel: r.RoleA.HierarchyLevel},
			PeerRole:   RoleInfo{ID: r.RoleB.ID.String(), Name: r.RoleB.Name, HierarchyLevel: r.RoleB.HierarchyLevel},
			Difficulty: string(r.Difficulty),
		})
	}

	// SPIN_RESULT for B
	if r.ClientB != nil {
		r.ClientB.Send(EventSpinResult, SpinResultPayload{
			RoomID:     r.ID,
			Context:    ctxInfo,
			YourRole:   RoleInfo{ID: r.RoleB.ID.String(), Name: r.RoleB.Name, HierarchyLevel: r.RoleB.HierarchyLevel},
			PeerRole:   RoleInfo{ID: r.RoleA.ID.String(), Name: r.RoleA.Name, HierarchyLevel: r.RoleA.HierarchyLevel},
			Difficulty: string(r.Difficulty),
		})
	}

	// SCENARIO_ASSIGN with private objectives
	if r.Scenario != nil {
		if r.ClientA != nil {
			r.ClientA.Send(EventScenarioAssign, ScenarioAssignPayload{
				RoomID:                 r.ID,
				ScenarioID:             r.Scenario.ID.String(),
				Title:                  r.Scenario.Title,
				Difficulty:             string(r.Scenario.Difficulty),
				BackgroundContext:      r.Scenario.BackgroundContext,
				YourObjective:          r.Scenario.RoleAObjective,
				YourConstraints:        r.Scenario.RoleAConstraints,
				PrepDurationSeconds:    r.PrepSeconds,
				SessionDurationSeconds: r.SessionSeconds,
			})
		}
		if r.ClientB != nil {
			r.ClientB.Send(EventScenarioAssign, ScenarioAssignPayload{
				RoomID:                 r.ID,
				ScenarioID:             r.Scenario.ID.String(),
				Title:                  r.Scenario.Title,
				Difficulty:             string(r.Scenario.Difficulty),
				BackgroundContext:      r.Scenario.BackgroundContext,
				YourObjective:          r.Scenario.RoleBObjective,
				YourConstraints:        r.Scenario.RoleBConstraints,
				PrepDurationSeconds:    r.PrepSeconds,
				SessionDurationSeconds: r.SessionSeconds,
			})
		}
	}

	// Advance to Prep Phase
	r.advanceToPrepLocked()
}

func (r *RoomManager) advanceToPrepLocked() {
	r.State = entity.RoomStatePrep
	_ = r.roomState.UpdateState(context.Background(), r.ID, entity.RoomStatePrep)

	ctx, cancel := context.WithCancel(context.Background())
	r.timerCancel = cancel

	go func(ctx context.Context) {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		remaining := r.PrepSeconds
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				remaining -= 5
				r.mu.Lock()
				if remaining <= 0 || (r.PrepReadyA && r.PrepReadyB) {
					r.advanceToSignalingLocked()
					r.mu.Unlock()
					return
				}

				if r.ClientA != nil {
					r.ClientA.Send(EventPrepTimerTick, PrepTimerTickPayload{
						RoomID:           r.ID,
						SecondsRemaining: remaining,
						PeerReady:        r.PrepReadyB,
					})
				}
				if r.ClientB != nil {
					r.ClientB.Send(EventPrepTimerTick, PrepTimerTickPayload{
						RoomID:           r.ID,
						SecondsRemaining: remaining,
						PeerReady:        r.PrepReadyA,
					})
				}
				r.mu.Unlock()
			}
		}
	}(ctx)
}

func (r *RoomManager) advanceToSignalingLocked() {
	if r.timerCancel != nil {
		r.timerCancel()
	}
	r.State = entity.RoomStateSignaling
	_ = r.roomState.UpdateState(context.Background(), r.ID, entity.RoomStateSignaling)

	prepEnd := PrepEndPayload{
		RoomID:        r.ID,
		InitiatorSeat: "A", // Peer A creates WebRTC offer
	}

	if r.ClientA != nil {
		r.ClientA.Send(EventPrepEnd, prepEnd)
	}
	if r.ClientB != nil {
		r.ClientB.Send(EventPrepEnd, prepEnd)
	}

	// Schedule live session timer
	ctx, cancel := context.WithCancel(context.Background())
	r.timerCancel = cancel

	go func(ctx context.Context) {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		remaining := r.SessionSeconds
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				remaining -= 10
				r.mu.Lock()
				if remaining <= 0 {
					r.advanceToEvaluatingLocked("timer_expired")
					r.mu.Unlock()
					return
				}

				tickPayload := SessionTimerTickPayload{
					RoomID:           r.ID,
					SecondsRemaining: remaining,
					Phase:            "live",
				}
				if r.ClientA != nil {
					r.ClientA.Send(EventSessionTimerTick, tickPayload)
				}
				if r.ClientB != nil {
					r.ClientB.Send(EventSessionTimerTick, tickPayload)
				}
				r.mu.Unlock()
			}
		}
	}(ctx)
}

func (r *RoomManager) advanceToEvaluatingLocked(reason string) {
	if r.timerCancel != nil {
		r.timerCancel()
	}
	r.State = entity.RoomStateEvaluating
	_ = r.roomState.UpdateState(context.Background(), r.ID, entity.RoomStateEvaluating)

	now := time.Now().UTC()
	_ = r.sessionRepo.SetEndedAt(context.Background(), r.SessionID, now)
	_ = r.sessionRepo.UpdateState(context.Background(), r.SessionID, entity.SessionStateEvaluating)

	endPayload := SessionEndPayload{
		RoomID: r.ID,
		Reason: reason,
	}

	if r.ClientA != nil {
		r.ClientA.Send(EventSessionComplete, endPayload)
	}
	if r.ClientB != nil {
		r.ClientB.Send(EventSessionComplete, endPayload)
	}
}

// HandlePrepReady marks a client as ready to begin the roleplay call.
func (r *RoomManager) HandlePrepReady(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State != entity.RoomStatePrep {
		return
	}

	if userID == r.UserA.ID.String() {
		r.PrepReadyA = true
	} else if userID == r.UserB.ID.String() {
		r.PrepReadyB = true
	}

	// If both ready, advance immediately
	if r.PrepReadyA && r.PrepReadyB {
		r.advanceToSignalingLocked()
	}
}

// HandleSignal relays WebRTC offer, answer, or ICE candidates to the other peer.
func (r *RoomManager) HandleSignal(fromUserID string, rawSignal json.RawMessage) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Parse & validate WebRTC signal structure
	_, err := webrtc.ParseSignal(rawSignal)
	if err != nil {
		log.Warn().Err(err).Str("user_id", fromUserID).Msg("ws: invalid webrtc signal rejected")
		return
	}

	var targetClient *Client
	var targetUserID string

	if fromUserID == r.UserA.ID.String() {
		targetClient = r.ClientB
		targetUserID = r.UserB.ID.String()
	} else if fromUserID == r.UserB.ID.String() {
		targetClient = r.ClientA
		targetUserID = r.UserA.ID.String()
	} else {
		return
	}

	signalPayload := SignalEnvelopePayload{
		RoomID: r.ID,
		Signal: rawSignal,
	}

	// If target peer is connected, forward immediately
	if targetClient != nil {
		targetClient.Send(EventSignal, signalPayload)
	} else {
		// Peer is temporarily disconnected, cache signal in Redis (TTL: 60s)
		_ = r.webrtcCache.StoreSignal(context.Background(), r.ID, targetUserID, rawSignal, 60*time.Second)
	}
}

func (r *RoomManager) flushCachedSignals(client *Client) {
	signals, err := r.webrtcCache.GetSignals(context.Background(), r.ID, client.UserID)
	if err != nil || len(signals) == 0 {
		return
	}

	for _, sig := range signals {
		client.Send(EventSignal, SignalEnvelopePayload{
			RoomID: r.ID,
			Signal: json.RawMessage(sig),
		})
	}
	_ = r.webrtcCache.ClearSignals(context.Background(), r.ID, client.UserID)
}

// HandleSessionEnd allows early session termination by either user.
func (r *RoomManager) HandleSessionEnd(userID string, reason string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.State == entity.RoomStateLive || r.State == entity.RoomStateSignaling {
		r.advanceToEvaluatingLocked(reason)
	}
}

// HandleDisconnect marks a peer as disconnected and notifies the counterpart.
func (r *RoomManager) HandleDisconnect(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var counterpart *Client
	if userID == r.UserA.ID.String() {
		r.ClientA = nil
		counterpart = r.ClientB
	} else if userID == r.UserB.ID.String() {
		r.ClientB = nil
		counterpart = r.ClientA
	}

	if counterpart != nil {
		counterpart.Send(EventPeerDisconnected, PeerDisconnectedPayload{
			RoomID:                 r.ID,
			ReconnectWindowSeconds: 30,
		})
	}
}