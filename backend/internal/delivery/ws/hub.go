package ws

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/roleplay-sim/backend/internal/domain/entity"
	"github.com/roleplay-sim/backend/internal/domain/repository"
	"github.com/roleplay-sim/backend/internal/usecase/evaluation"
	"github.com/rs/zerolog/log"
)

// Hub is the central registry and router for real-time WebSocket traffic.
type Hub struct {
	clients map[string]*Client      // userID -> *Client
	rooms   map[string]*RoomManager // roomID -> *RoomManager
	mu      sync.RWMutex

	userRepo     repository.UserRepository
	sessionRepo  repository.SessionRepository
	scenarioRepo repository.ScenarioRepository
	queueRepo    repository.MatchmakingQueueRepository
	roomState    repository.RoomStateRepository
	webrtcCache  repository.WebRTCCacheRepository
	audioBuffer  repository.AudioBufferRepository
	evalUseCase  *evaluation.EvaluateSessionUseCase
}

func NewHub(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	scenarioRepo repository.ScenarioRepository,
	queueRepo repository.MatchmakingQueueRepository,
	roomState repository.RoomStateRepository,
	webrtcCache repository.WebRTCCacheRepository,
	audioBuffer repository.AudioBufferRepository,
	evalUseCase *evaluation.EvaluateSessionUseCase,
) *Hub {
	return &Hub{
		clients:      make(map[string]*Client),
		rooms:        make(map[string]*RoomManager),
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		scenarioRepo: scenarioRepo,
		queueRepo:    queueRepo,
		roomState:    roomState,
		webrtcCache:  webrtcCache,
		audioBuffer:  audioBuffer,
		evalUseCase:  evalUseCase,
	}
}

// Run starts background workers such as the matchmaking queue consumer.
func (h *Hub) Run(ctx context.Context) {
	log.Info().Msg("ws: hub started")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("ws: hub stopping")
			return
		case <-ticker.C:
			h.pollMatchmakingQueue(ctx)
		}
	}
}

func (h *Hub) pollMatchmakingQueue(ctx context.Context) {
	difficulties := []string{"easy", "medium", "hard"}
	for _, diff := range difficulties {
		userIDA, userIDB, err := h.queueRepo.Dequeue(ctx, diff)
		if err != nil {
			log.Error().Err(err).Str("difficulty", diff).Msg("ws: failed to dequeue pair")
			continue
		}
		if userIDA == "" || userIDB == "" {
			continue
		}

		go func(d string, a, b string) {
			if err := h.createMatch(context.Background(), a, b, d); err != nil {
				log.Error().Err(err).Str("userA", a).Str("userB", b).Str("difficulty", d).Msg("ws: failed to create match")
			}
		}(diff, userIDA, userIDB)
	}
}

func (h *Hub) createMatch(ctx context.Context, userIDA, userIDB, difficulty string) error {
	idA, err := uuid.Parse(userIDA)
	if err != nil {
		return err
	}
	idB, err := uuid.Parse(userIDB)
	if err != nil {
		return err
	}

	userA, err := h.userRepo.GetByID(ctx, idA)
	if err != nil {
		return err
	}
	userB, err := h.userRepo.GetByID(ctx, idB)
	if err != nil {
		return err
	}

	diff := entity.Difficulty(difficulty)

	// Pick random scenario from database matching the exact difficulty
	scenario, err := h.scenarioRepo.GetRandom(ctx, diff, nil)
	if err != nil {
		log.Warn().Err(err).Str("difficulty", difficulty).Msg("ws: could not fetch random scenario, using default")
		scenario = &entity.Scenario{
			ID:                     uuid.New(),
			Title:                  "Leave Request Under Deadline Pressure",
			Difficulty:             diff,
			BackgroundContext:      "The team is 3 weeks from a major product release. Junior requests 2 days leave.",
			RoleAObjective:         "Secure 2 days leave while keeping a collaborative relationship.",
			RoleAConstraints:       []string{"Do not reveal a personal emergency", "Remain professional"},
			RoleBObjective:         "Limit leave to 1 day due to strict release deadlines.",
			RoleBConstraints:       []string{"Do not demoralize the junior dev", "Protect sprint delivery"},
			PrepDurationSeconds:    180,
			SessionDurationSeconds: 360,
		}
	}

	roomID := "room_" + uuid.New().String()[:8]

	// Create Postgres Session
	session := &entity.Session{
		RoomID:     roomID,
		ScenarioID: scenario.ID,
		Difficulty: diff,
		State:      entity.SessionStateWaiting,
	}
	if err := h.sessionRepo.Create(ctx, session); err != nil {
		return err
	}

	// Add Participants (A = Seat A, B = Seat B)
	_ = h.sessionRepo.AddParticipant(ctx, &entity.SessionParticipant{
		SessionID: session.ID,
		UserID:    userA.ID,
		RoleID:    uuid.New(),
		Seat:      "A",
	})
	_ = h.sessionRepo.AddParticipant(ctx, &entity.SessionParticipant{
		SessionID: session.ID,
		UserID:    userB.ID,
		RoleID:    uuid.New(),
		Seat:      "B",
	})

	// Create Redis Room State
	redisRoom := &entity.Room{
		ID:    roomID,
		State: entity.RoomStateWaiting,
		PeerA: entity.Peer{
			UserID:      userA.ID.String(),
			DisplayName: userA.DisplayName,
			EloRating:   userA.EloRating,
			Seat:        "A",
		},
		PeerB: entity.Peer{
			UserID:      userB.ID.String(),
			DisplayName: userB.DisplayName,
			EloRating:   userB.EloRating,
			Seat:        "B",
		},
		ScenarioID: scenario.ID.String(),
		Difficulty: diff,
	}
	_ = h.roomState.Create(ctx, redisRoom)

	// Create In-Memory Room Manager
	rm := NewRoomManager(
		roomID,
		userA, userB,
		diff,
		scenario,
		h.sessionRepo,
		h.scenarioRepo,
		h.roomState,
		h.webrtcCache,
		h.evalUseCase,
	)
	rm.SessionID = session.ID

	h.mu.Lock()
	h.rooms[roomID] = rm
	clientA := h.clients[userA.ID.String()]
	clientB := h.clients[userB.ID.String()]
	h.mu.Unlock()

	log.Info().Str("room_id", roomID).Str("userA", userA.DisplayName).Str("userB", userB.DisplayName).Msg("ws: match created")

	// If clients already connected via WebSocket, attach them
	if clientA != nil {
		rm.AttachClient(clientA)
	}
	if clientB != nil {
		rm.AttachClient(clientB)
	}

	return nil
}

// Register adds an authenticated client to the hub.
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.UserID] = client
	log.Debug().Str("user_id", client.UserID).Str("conn_id", client.ID).Msg("ws: client registered")

	// Check if client belongs to an existing active room
	for _, rm := range h.rooms {
		if rm.UserA.ID.String() == client.UserID || rm.UserB.ID.String() == client.UserID {
			go rm.AttachClient(client)
			break
		}
	}
}

// Unregister removes a client upon disconnect.
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if cur, ok := h.clients[client.UserID]; ok && cur.ID == client.ID {
		delete(h.clients, client.UserID)
		log.Debug().Str("user_id", client.UserID).Msg("ws: client unregistered")

		roomID := client.GetRoomID()
		if rm, ok := h.rooms[roomID]; ok {
			go rm.HandleDisconnect(client.UserID)
		}
	}
}

// HandleMessage routes incoming framed events from clients.
func (h *Hub) HandleMessage(client *Client, env *Envelope) {
	switch env.Event {
	case EventPing:
		client.Send(EventPong, PongPayload{
			ServerTime:    time.Now().UTC().Format(time.RFC3339),
			LatencyHintMs: 15,
		})

	case EventJoinQueue:
		var qp struct {
			Difficulty string `json:"difficulty"`
		}
		diff := "medium"
		if err := json.Unmarshal(env.Payload, &qp); err == nil && qp.Difficulty != "" {
			diff = strings.ToLower(qp.Difficulty)
		}
		uID, err := uuid.Parse(client.UserID)
		if err == nil {
			u, err2 := h.userRepo.GetByID(context.Background(), uID)
			if err2 == nil {
				_ = h.queueRepo.Enqueue(context.Background(), client.UserID, u.EloRating, diff)
				log.Info().Str("user_id", client.UserID).Str("difficulty", diff).Msg("ws: user joined matchmaking queue via WS")
			}
		}

	case EventLeaveQueue:
		_ = h.queueRepo.Remove(context.Background(), client.UserID)
		log.Info().Str("user_id", client.UserID).Msg("ws: user left matchmaking queue via WS")

	case EventPrepReady:
		room := h.getRoomForClient(client)
		if room != nil {
			room.HandlePrepReady(client.UserID)
		}

	case EventSignal:
		var sigPayload SignalEnvelopePayload
		if err := json.Unmarshal(env.Payload, &sigPayload); err == nil {
			room := h.getRoomByID(sigPayload.RoomID)
			if room != nil {
				room.HandleSignal(client.UserID, sigPayload.Signal)
			}
		}

	case EventSessionEnd:
		var endPayload SessionEndPayload
		reason := "user_ended"
		if err := json.Unmarshal(env.Payload, &endPayload); err == nil && endPayload.Reason != "" {
			reason = endPayload.Reason
		}
		room := h.getRoomForClient(client)
		if room != nil {
			room.HandleSessionEnd(client.UserID, reason)
		}

	default:
		log.Warn().Str("event", string(env.Event)).Str("user_id", client.UserID).Msg("ws: unhandled event")
	}
}

func (h *Hub) HandleBinaryAudio(client *Client, data []byte) {
	roomID := client.GetRoomID()
	if roomID == "" || len(data) == 0 || h.audioBuffer == nil {
		return
	}

	go func() {
		_ = h.audioBuffer.AppendChunk(context.Background(), roomID, client.UserID, 0, data)
	}()
}

func (h *Hub) getRoomForClient(client *Client) *RoomManager {
	roomID := client.GetRoomID()
	if roomID == "" {
		return nil
	}
	return h.getRoomByID(roomID)
}

func (h *Hub) getRoomByID(roomID string) *RoomManager {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rooms[roomID]
}

// GetConnectedClientCount returns the number of active WS connections.
func (h *Hub) GetConnectedClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}