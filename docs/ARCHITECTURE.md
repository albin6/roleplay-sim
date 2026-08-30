# System Architecture Document

> **Project:** Real-Time Scenario-Based Roleplay Simulator
> **Version:** 1.1.0 | **Phase:** 1 — Documentation | **Stack:** LOCKED 2026-08-30

---

## Table of Contents

1. [High-Level System Overview](#1-high-level-system-overview)
2. [Component Architecture](#2-component-architecture)
3. [Go Backend — Clean Architecture Folder Structure](#3-go-backend--clean-architecture-folder-structure)
4. [WebRTC Signaling Flow](#4-webrtc-signaling-flow)
5. [WebSocket State Machine Lifecycle](#5-websocket-state-machine-lifecycle)
6. [Redis Matchmaking Queue & Room State Management](#6-redis-matchmaking-queue--room-state-management)
7. [Concurrency & Scalability Model](#7-concurrency--scalability-model)
8. [Security Architecture](#8-security-architecture)

---

## 1. High-Level System Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            CLIENT LAYER                                      │
│  ┌──────────────────────────────┐   ┌──────────────────────────────────┐   │
│  │   Browser (User A)           │   │   Browser (User B)               │   │
│  │   Next.js 14 + TypeScript    │   │   Next.js 14 + TypeScript        │   │
│  │   WebRTC PeerConnection      │◄──┼──►WebRTC PeerConnection          │   │
│  └──────────────┬───────────────┘   └─────────────────┬────────────────┘   │
│                 │ WebSocket + HTTPS                    │ WebSocket + HTTPS  │
└─────────────────┼──────────────────────────────────────┼────────────────────┘
                  │                                      │
┌─────────────────▼──────────────────────────────────────▼────────────────────┐
│                          API GATEWAY / LOAD BALANCER                         │
│              (NGINX / Traefik — TLS termination, rate limiting)              │
└──────────────────────────────────┬──────────────────────────────────────────┘
                                   │
┌──────────────────────────────────▼──────────────────────────────────────────┐
│                          GO BACKEND CLUSTER                                   │
│                                                                               │
│  ┌─────────────────────┐  ┌─────────────────────┐  ┌──────────────────────┐ │
│  │  HTTP REST Server   │  │  WebSocket Hub      │  │  Background Workers  │ │
│  │  (Auth, User, Score)│  │  (Signaling, Room)  │  │  (AI Pipeline, Elo)  │ │
│  └──────────┬──────────┘  └──────────┬──────────┘  └──────────┬───────────┘ │
│             │                        │                          │             │
│  ┌──────────▼────────────────────────▼──────────────────────────▼───────────┐│
│  │                        USE CASE / SERVICE LAYER                           ││
│  │  MatchmakingUseCase │ SessionUseCase │ EvaluationUseCase │ UserUseCase   ││
│  └──────────┬──────────────────────────────────────────────────────────────┘│
│             │                                                                 │
│  ┌──────────▼──────────────────────────────────────────────────────────────┐ │
│  │                        REPOSITORY LAYER                                  │ │
│  │   PostgresUserRepo │ PostgresSessionRepo │ RedisRoomRepo │ RedisQueueRepo│ │
│  └──────────┬──────────────────────────────────────────────┬───────────────┘ │
└─────────────┼──────────────────────────────────────────────┼─────────────────┘
              │                                              │
┌─────────────▼──────────────────┐     ┌────────────────────▼─────────────────┐
│        PostgreSQL 16           │     │            Redis 7                    │
│   (Users, Sessions, Roles,     │     │   (Matchmaking Queue, Room State,     │
│    Scenarios, Evaluations,     │     │    WebRTC Cache, Pub/Sub Channels)    │
│    Leaderboard)                │     │                                       │
└────────────────────────────────┘     └───────────────────────────────────────┘
              │
┌─────────────▼──────────────────────────────────────────────────────────────┐
│                          EXTERNAL AI SERVICES                                │
│   ┌─────────────────────────┐       ┌─────────────────────────────────────┐ │
│   │  Whisper STT Service    │       │  LLM Evaluation API                 │ │
│   │  (Dual-channel audio    │       │  (OpenAI GPT-4o / Gemini Pro)       │ │
│   │   transcription)        │       │  (Rubric scoring, JSON output)      │ │
│   └─────────────────────────┘       └─────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Component Architecture

### 2.1 Request Flow — REST API

```
Client HTTP Request
      │
      ▼
API Gateway (NGINX) ──► Rate Limiter ──► JWT Auth Middleware
      │
      ▼
HTTP Handler (delivery/http/)
      │  validates & parses request
      ▼
Use Case (internal/usecase/)
      │  applies business rules
      ▼
Repository Interface (internal/repository/)
      │  calls concrete implementation
      ▼
PostgreSQL / Redis
```

### 2.2 Real-Time Flow — WebSocket + WebRTC

```
Client WebSocket Connect
      │
      ▼
WS Hub → RegisterClient → Assign to Room
      │
      ├──► Room does not exist → Matchmaking Queue (Redis)
      │         │
      │         └──► Second user joins → Create Room → Broadcast RoomReady
      │
      └──► Room exists → Join Room → Broadcast PeerJoined
                │
                ▼
         Spin Wheel Phase
         (Server generates seed → broadcasts to both peers)
                │
                ▼
         Scenario Assignment
         (UseCase selects tiered scenario → broadcasts payload)
                │
                ▼
         Prep Phase Timer
         (Server-side countdown broadcast every second)
                │
                ▼
         WebRTC Signaling Exchange
         ┌─────────────────────────────────┐
         │  Offer  ──────────────────────► │
         │         ◄──────────────  Answer │
         │  ICE Candidates ◄──────────────►│
         └─────────────────────────────────┘
                │
                ▼
         Live Session Timer (5–6 min)
                │
                ▼
         Session End Signal
         (Server closes room, triggers AI pipeline)
```

---

## 3. Go Backend — Clean Architecture Folder Structure

```
/
├── cmd/
│   └── server/
│       └── main.go                    ← Binary entrypoint; wires all dependencies
│
├── internal/
│   │
│   ├── domain/                        ← Pure business entities (no external deps)
│   │   ├── entity/
│   │   │   ├── user.go                ← User, UserRole, EloRating value objects
│   │   │   ├── session.go             ← Session, SessionState enum, Timer
│   │   │   ├── room.go                ← Room, RoomState, Peer
│   │   │   ├── scenario.go            ← Scenario, Difficulty, ConflictObjective
│   │   │   ├── role.go                ← Role, RolePair, Context (IT/HR/Sales etc.)
│   │   │   └── evaluation.go          ← Evaluation, RubricScore, TranscriptSegment
│   │   ├── repository/
│   │   │   ├── user_repository.go     ← UserRepository interface
│   │   │   ├── session_repository.go  ← SessionRepository interface
│   │   │   ├── room_repository.go     ← RoomRepository interface (Redis)
│   │   │   ├── scenario_repository.go ← ScenarioRepository interface
│   │   │   └── evaluation_repository.go
│   │   └── errors/
│   │       └── domain_errors.go       ← Sentinel domain errors
│   │
│   ├── usecase/                       ← Application business logic
│   │   ├── auth/
│   │   │   ├── register_usecase.go
│   │   │   └── login_usecase.go
│   │   ├── matchmaking/
│   │   │   ├── enqueue_usecase.go     ← Add user to matchmaking queue
│   │   │   └── match_usecase.go       ← Dequeue pair, create room
│   │   ├── session/
│   │   │   ├── create_session_usecase.go
│   │   │   ├── spin_wheel_usecase.go  ← Deterministic seed generation
│   │   │   ├── assign_scenario_usecase.go
│   │   │   ├── start_prep_usecase.go
│   │   │   ├── start_session_usecase.go
│   │   │   └── end_session_usecase.go ← Triggers AI pipeline
│   │   ├── evaluation/
│   │   │   ├── transcribe_usecase.go  ← STT orchestration
│   │   │   ├── evaluate_usecase.go    ← LLM rubric scoring
│   │   │   └── update_elo_usecase.go  ← Elo recalculation
│   │   └── leaderboard/
│   │       └── fetch_leaderboard_usecase.go
│   │
│   ├── repository/                    ← Concrete implementations
│   │   ├── postgres/
│   │   │   ├── user_repo.go
│   │   │   ├── session_repo.go
│   │   │   ├── scenario_repo.go
│   │   │   └── evaluation_repo.go
│   │   └── redis/
│   │       ├── matchmaking_queue_repo.go   ← ZADD/ZPOPMIN on sorted set
│   │       ├── room_state_repo.go          ← HSET room state as hash
│   │       └── webrtc_cache_repo.go        ← SDP/ICE candidate TTL cache
│   │
│   └── delivery/
│       ├── http/
│       │   ├── middleware/
│       │   │   ├── auth_middleware.go      ← JWT validation
│       │   │   ├── cors_middleware.go
│       │   │   └── rate_limit_middleware.go
│       │   ├── handler/
│       │   │   ├── auth_handler.go
│       │   │   ├── user_handler.go
│       │   │   ├── session_handler.go
│       │   │   ├── leaderboard_handler.go
│       │   │   └── health_handler.go
│       │   └── router.go                   ← chi.Router route registration
│       └── ws/
│           ├── hub.go                      ← Central WebSocket connection registry
│           ├── client.go                   ← Per-connection read/write pumps
│           ├── room.go                     ← Room-scoped broadcast logic
│           └── handler.go                  ← WS upgrade + event dispatch (nhooyr.io/websocket)
│
├── pkg/
│   ├── webrtc/
│   │   └── signaling.go                    ← SDP/ICE relay helpers
│   ├── ai/
│   │   ├── deepgram/
│   │   │   └── client.go                   ← Deepgram Nova-2 multichannel STT
│   │   ├── whisper/
│   │   │   └── client.go                   ← OpenAI Whisper fallback STT
│   │   └── gemini/
│   │       └── client.go                   ← Gemini 1.5 Pro/Flash evaluation (google.golang.org/genai)
│   ├── elo/
│   │   └── calculator.go                   ← Elo formula with role/difficulty modifier
│   ├── jwt/
│   │   └── jwt.go                          ← RS256 token generation & validation
│   └── validator/
│       └── request_validator.go
│
├── migrations/
│   ├── 001_create_users.sql
│   ├── 002_create_roles_scenarios.sql
│   ├── 003_create_sessions.sql
│   ├── 004_create_evaluations.sql
│   └── 005_create_leaderboard.sql
│
├── frontend/                               ← Next.js 14 App Router
│   ├── app/
│   ├── components/
│   ├── hooks/
│   └── lib/
│
├── docker-compose.yml
├── Makefile
└── AGENT.md
```

### Clean Architecture Dependency Rule

```
┌──────────────────────────────────────────────────────────┐
│  DOMAIN (entity, repository interfaces, domain errors)   │ ← No external deps
└─────────────────────────┬────────────────────────────────┘
                          │ depends on ↑
┌─────────────────────────▼────────────────────────────────┐
│  USE CASE (business logic, orchestration)                 │ ← Imports domain only
└─────────────────────────┬────────────────────────────────┘
                          │ depends on ↑
┌─────────────────────────▼────────────────────────────────┐
│  REPOSITORY (Postgres, Redis concrete implementations)    │ ← Implements domain interfaces
│  DELIVERY   (HTTP handlers, WebSocket hub)               │ ← Calls use cases
└──────────────────────────────────────────────────────────┘
```

**Rule:** Arrows point inward only. Domain never imports from use case, repository, or delivery.

---

## 4. WebRTC Signaling Flow

WebRTC peer connections are established with the Go server acting as a **signaling relay** only. Media flows directly P2P between browsers.

```
User A (Initiator)              Go WS Server (Relay)          User B (Receiver)
       │                               │                              │
       │── ws:connect ────────────────►│                              │
       │                               │◄─────────────── ws:connect ──│
       │                               │                              │
       │  [Server detects room full]   │                              │
       │◄────────── PEER_JOINED ───────┤──────────── PEER_JOINED ────►│
       │                               │                              │
       │  [User A creates RTCPeerConnection, getUserMedia()]          │
       │  [User A createOffer()]       │                              │
       │                               │                              │
       │── SIGNAL {type:"offer",  ────►│                              │
       │    sdp: <SDP_A>}              │──── SIGNAL {type:"offer"} ──►│
       │                               │                              │
       │                               │   [User B createAnswer()]    │
       │                               │◄─── SIGNAL {type:"answer"} ──│
       │◄── SIGNAL {type:"answer", ────┤                              │
       │     sdp: <SDP_B>}             │                              │
       │                               │                              │
       │── SIGNAL {type:"ice", ───────►│                              │
       │    candidate: <ICE_A>}        │──── SIGNAL {type:"ice"} ────►│
       │                               │◄─── SIGNAL {type:"ice"} ─────│
       │◄── SIGNAL {type:"ice"} ───────┤                              │
       │                               │                              │
       │◄══════════════ P2P WebRTC Audio/Video Stream ══════════════► │
       │           (Direct peer connection — server not involved)      │
       │                               │                              │
       │── SESSION_END ───────────────►│◄──────────── SESSION_END ────│
       │                               │                              │
       │                    [Server triggers AI pipeline]             │
```

### TURN/STUN Configuration

| Server Type | Purpose                                        | Provider             |
|-------------|------------------------------------------------|----------------------|
| STUN        | NAT traversal — discover public IP/port        | Google STUN          |
| TURN        | Relay fallback when direct P2P fails           | Coturn (self-hosted) |

---

## 5. WebSocket State Machine Lifecycle

Each room's WebSocket session progresses through a strict finite state machine:

```
                    ┌─────────────┐
                    │   IDLE      │  (No users connected)
                    └──────┬──────┘
                           │ User A connects + enqueued
                           ▼
                    ┌─────────────┐
                    │  WAITING    │  (1/2 users in matchmaking queue)
                    └──────┬──────┘
                           │ User B matched → Room Created
                           ▼
                    ┌─────────────┐
                    │  READY      │  (Both peers connected, room initialized)
                    └──────┬──────┘
                           │ Server triggers spin
                           ▼
                    ┌─────────────┐
                    │  SPINNING   │  (Wheel animation; seed broadcast)
                    └──────┬──────┘
                           │ Spin complete → scenario selected
                           ▼
                    ┌─────────────┐
                    │  SCENARIO   │  (Scenario details sent to both peers)
                    └──────┬──────┘
                           │ Prep timer started
                           ▼
                    ┌─────────────┐
                    │   PREP      │  (2–5 min countdown; both peers see timer)
                    └──────┬──────┘
                           │ Prep timer expires
                           ▼
                    ┌─────────────┐
                    │ SIGNALING   │  (WebRTC offer/answer/ICE exchange)
                    └──────┬──────┘
                           │ P2P connection established
                           ▼
                    ┌─────────────┐
                    │    LIVE     │  (5–6 min session timer running)
                    └──────┬──────┘
                           │ Timer expires OR both users send SESSION_END
                           ▼
                    ┌─────────────┐
                    │ EVALUATING  │  (Audio upload, STT, LLM eval in progress)
                    └──────┬──────┘
                           │ Evaluation complete
                           ▼
                    ┌─────────────┐
                    │  COMPLETE   │  (Scores + feedback delivered to clients)
                    └──────┬──────┘
                           │ Clients disconnect
                           ▼
                    ┌─────────────┐
                    │  CLOSED     │  (Room destroyed; Redis state purged)
                    └─────────────┘

  Failure path (any state) ──► DISCONNECTED ──► Reconnect window (30s) ──► CLOSED
```

### State Transition Table

| Current State | Event                    | Next State   | Server Action                               |
|--------------|--------------------------|--------------|----------------------------------------------|
| IDLE         | user_enqueued            | WAITING      | Add to Redis matchmaking queue               |
| WAITING      | second_user_matched      | READY        | Create room, broadcast ROOM_READY            |
| READY        | server_spin_trigger      | SPINNING     | Generate seed, broadcast SPIN_START          |
| SPINNING     | spin_complete            | SCENARIO     | Select scenario, broadcast SCENARIO_ASSIGN   |
| SCENARIO     | both_ack                 | PREP         | Start prep countdown timer                   |
| PREP         | timer_expired            | SIGNALING    | Broadcast PREP_END, initiate WebRTC          |
| SIGNALING    | p2p_established          | LIVE         | Start session countdown timer                |
| LIVE         | timer_expired OR end_cmd | EVALUATING   | Trigger AI pipeline worker                   |
| EVALUATING   | eval_complete            | COMPLETE     | Broadcast scores, update Elo                 |
| COMPLETE     | clients_disconnect       | CLOSED       | Purge room from Redis                        |
| ANY          | peer_disconnect          | DISCONNECTED | Start 30s reconnect window                   |
| DISCONNECTED | reconnect_timeout        | CLOSED       | Clean up room                                |

---

## 6. Redis Matchmaking Queue & Active Room State Management

### 6.1 Matchmaking Queue

**Data Structure:** Redis Sorted Set

```
Key: matchmaking:queue
Score: Unix timestamp (FIFO ordering)
Member: user_id
```

**Operations:**
```
ZADD matchmaking:queue <timestamp> <user_id>   ← Enqueue
ZPOPMIN matchmaking:queue 2                     ← Dequeue matched pair
ZREM matchmaking:queue <user_id>               ← Cancel/timeout
ZCARD matchmaking:queue                         ← Queue depth monitoring
```

**Matchmaking Worker Loop:**
```
Every 500ms:
  1. ZPOPMIN matchmaking:queue 2
  2. If count == 2 → CreateRoom(userA, userB)
  3. Broadcast ROOM_READY to both via pub/sub
  4. If count == 1 → ZADD back (re-enqueue) + wait
```

### 6.2 Active Room State

**Data Structure:** Redis Hash per room

```
Key: room:<room_id>

Fields:
  state              → "READY" | "SPINNING" | "PREP" | "LIVE" | ...
  user_a             → user_id
  user_b             → user_id
  scenario_id        → UUID
  difficulty         → "easy" | "medium" | "hard"
  role_a             → "junior_dev"
  role_b             → "team_lead"
  prep_expires_at    → Unix timestamp
  session_expires_at → Unix timestamp
  created_at         → Unix timestamp
```

**Operations:**
```
HSET   room:<id> field value     ← Create/update field
HGETALL room:<id>                ← Load full room state
HDEL   room:<id> field           ← Remove field
DEL    room:<id>                 ← Destroy room (on CLOSED)
EXPIRE room:<id> 3600            ← Auto-expiry safety net (1 hour)
```

### 6.3 WebRTC Session Cache

**Data Structure:** Redis String with TTL

```
Key: webrtc:<room_id>:sdp:<user_id>
Value: JSON-encoded SDP offer/answer
TTL: 60 seconds

Key: webrtc:<room_id>:ice:<user_id>
Value: JSON array of ICE candidates
TTL: 60 seconds
```

### 6.4 Pub/Sub Channels

```
Channel: room:<room_id>:events
  → Used by WS hub nodes to broadcast room events across server instances

Channel: matchmaking:notifications
  → Used by matchmaking worker to notify WS hubs of new room creations
```

### 6.5 Leaderboard Cache

**Data Structure:** Redis Sorted Set

```
Key: leaderboard:global
Score: elo_rating (float64)
Member: user_id

ZADD leaderboard:global <elo> <user_id>       ← Upsert after session
ZREVRANGE leaderboard:global 0 99 WITHSCORES  ← Top 100
ZRANK leaderboard:global <user_id>            ← User rank
```

---

## 7. Concurrency & Scalability Model

### 7.1 WebSocket Hub Architecture

```
Go Backend Instance
│
├── Hub (singleton goroutine)
│   ├── register chan *Client
│   ├── unregister chan *Client
│   └── broadcast chan RoomMessage
│
├── Room[room_id_1]
│   ├── Client A (goroutine: readPump + writePump)
│   └── Client B (goroutine: readPump + writePump)
│
└── Room[room_id_2]
    ├── Client C
    └── Client D
```

Each `Client` runs two goroutines:
- **readPump:** Blocking read from WebSocket → dispatch to Hub
- **writePump:** Blocking write to WebSocket ← receive from channel

### 7.2 Horizontal Scaling Strategy

For multi-instance deployment:
- WebSocket clients on different instances communicate via **Redis Pub/Sub**
- Room state is authoritative in Redis (not in-memory)
- Sticky sessions via load balancer (NGINX `ip_hash`) for WebSocket affinity

### 7.3 Target Capacity

| Metric                          | Target            |
|---------------------------------|-------------------|
| Concurrent active rooms         | 1,000+            |
| WebSocket connections per node  | ~2,000            |
| Nodes for 1,000 rooms           | ~1–2              |
| Matchmaking queue depth max     | 10,000 users      |
| WebSocket event latency (p99)   | < 100ms           |
| WebRTC signaling latency (p99)  | < 200ms           |

---

## 8. Security Architecture

| Concern                  | Mechanism                                                         |
|--------------------------|-------------------------------------------------------------------|
| Authentication           | JWT (RS256), short-lived access tokens (15min) + refresh tokens  |
| WebSocket Auth           | JWT validated on upgrade handshake via query param or header     |
| Rate Limiting            | NGINX + token bucket per IP (REST) and per user (WS)             |
| Room Isolation           | Each room uses unique UUID; no cross-room data access            |
| Audio Privacy            | Raw audio deleted after transcription; transcripts retained      |
| Input Validation         | Server-side validation for all WS event payloads                 |
| CORS                     | Strict allowlist for frontend origin                             |
| SQL Injection            | Parameterized queries via `pgx` driver only                      |
| XSS / CSRF               | SameSite cookies; CSRF token on state-changing REST endpoints    |
