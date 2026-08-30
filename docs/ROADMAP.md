# Implementation Roadmap

> **Project:** Real-Time Scenario-Based Roleplay Simulator
> **Version:** 1.0.0 | **Phase:** 1 — Documentation

---

## Table of Contents

1. [Roadmap Philosophy](#1-roadmap-philosophy)
2. [Phase Overview](#2-phase-overview)
3. [Phase 1 — Architecture & Documentation](#3-phase-1--architecture--documentation)
4. [Phase 2 — Backend Core](#4-phase-2--backend-core)
5. [Phase 3 — Real-Time Engine (WebSocket + WebRTC)](#5-phase-3--real-time-engine-websocket--webrtc)
6. [Phase 4 — AI Evaluation Pipeline](#6-phase-4--ai-evaluation-pipeline)
7. [Phase 5 — Frontend Implementation](#7-phase-5--frontend-implementation)
8. [Phase 6 — Testing, Observability & Deployment](#8-phase-6--testing-observability--deployment)
9. [Milestone Summary Table](#9-milestone-summary-table)
10. [Risk Register](#10-risk-register)

---

## 1. Roadmap Philosophy

- **Documentation First:** All architecture decisions are frozen in docs before any code is written.
- **Inside-Out:** Build domain → use cases → repositories → delivery in that order.
- **Vertical Slices:** Each phase delivers a runnable, testable slice of functionality.
- **No Premature Optimization:** Ship correct behavior first; optimize after profiling.
- **Feature Flags:** All non-trivial features are gated behind feature flags from day one.

---

## 2. Phase Overview

```
Phase 1 ──► Phase 2 ──► Phase 3 ──► Phase 4 ──► Phase 5 ──► Phase 6
  Docs       Backend     Real-Time    AI Eval     Frontend    Deploy
  ✅ Now      ~2 weeks   ~3 weeks     ~2 weeks    ~3 weeks    ~2 weeks
```

**Total estimated timeline: ~12 weeks** (solo or small team; adjust for team size)

---

## 3. Phase 1 — Architecture & Documentation

**Goal:** Complete all design documentation before a single line of production code is written.

**Status:** ✅ In Progress

### Milestones

#### M1.1 — Project Scaffold & Documentation
- [x] Create `AGENT.md` root context file
- [x] Create `docs/ARCHITECTURE.md` — system design, folder structure, state machines
- [x] Create `docs/DATABASE.md` — PostgreSQL schema + Redis data models
- [x] Create `docs/API_SPECS.md` — REST + WebSocket contracts
- [x] Create `docs/AI_EVALUATION.md` — STT pipeline, LLM prompts, Elo formula
- [x] Create `docs/ROADMAP.md` — this file

#### M1.2 — Repository Bootstrap (Post-Review)
- [ ] Initialize Go module: `go mod init github.com/org/roleplay-sim`
- [ ] Initialize Next.js 14 project in `frontend/`
- [ ] Set up `docker-compose.yml` with PostgreSQL, Redis, and Go backend
- [ ] Set up `Makefile` with `make dev`, `make test`, `make migrate`, `make build`
- [ ] Create `.github/workflows/ci.yml` for lint + test pipeline
- [ ] Set up `golangci-lint` configuration
- [ ] Configure `eslint` + `prettier` for frontend

**Deliverable:** Repository is runnable with `make dev`; both DB and Redis are reachable from the Go server.

---

## 4. Phase 2 — Backend Core

**Goal:** Implement authentication, user management, scenario seeding, and the matchmaking queue without any real-time components.

**Prerequisites:** Phase 1 complete and reviewed.

### Milestones

#### M2.1 — Domain Layer
- [ ] Define `entity/user.go` — User, EloRating, UserState
- [ ] Define `entity/session.go` — Session, SessionState enum
- [ ] Define `entity/room.go` — Room, RoomState, Peer
- [ ] Define `entity/scenario.go` — Scenario, Difficulty, ConflictObjective
- [ ] Define `entity/role.go` — Role, RoleContext, HierarchyLevel
- [ ] Define `entity/evaluation.go` — Evaluation, RubricScore, Transcript
- [ ] Define all repository interfaces in `domain/repository/`
- [ ] Define all sentinel domain errors in `domain/errors/`
- [ ] Write unit tests for all entity value objects and validation logic

#### M2.2 — Database Migrations
- [ ] Write migration `001_create_users.sql`
- [ ] Write migration `002_create_role_contexts_and_roles.sql`
- [ ] Write migration `003_create_scenarios.sql`
- [ ] Write migration `004_create_sessions_and_participants.sql`
- [ ] Write migration `005_create_evaluations_and_rubric_scores.sql`
- [ ] Write migration `006_create_leaderboard_history.sql`
- [ ] Integrate `golang-migrate` or `goose` as migration runner
- [ ] Seed script: `role_contexts`, `roles`, and initial `scenarios` (min 5 per difficulty tier)

#### M2.3 — Repository Layer (PostgreSQL + sqlc)
- [ ] Configure `sqlc.yaml` targeting `migrations/` schema and `internal/repository/postgres/queries/`
- [ ] Write SQL query files in `internal/repository/postgres/queries/` for all domain operations
- [ ] Run `sqlc generate` → produces type-safe Go in `internal/repository/postgres/db/`
- [ ] Implement `postgres/user_repo.go` — wraps sqlc calls: GetUserByEmail, CreateUser, UpdateEloRating
- [ ] Implement `postgres/session_repo.go` — CreateSession, GetSession, UpdateSessionState
- [ ] Implement `postgres/scenario_repo.go` — GetRandomScenarioByDifficulty, GetScenarioByID
- [ ] Implement `postgres/evaluation_repo.go` — CreateEvaluation, CreateRubricScores, GetEvaluationBySession
- [ ] Write integration tests for each repository using `testcontainers-go` (real PostgreSQL 16)

#### M2.4 — Repository Layer (Redis)
- [ ] Implement `redis/matchmaking_queue_repo.go` — Enqueue, Dequeue, Remove, Size
- [ ] Implement `redis/room_state_repo.go` — Create, Get, Update, Delete
- [ ] Implement `redis/webrtc_cache_repo.go` — StoreSDP, GetSDP, StoreICE, GetICE
- [ ] Write integration tests using `testcontainers-go` with real Redis

#### M2.5 — Auth Use Cases & HTTP Handler
- [ ] Implement `usecase/auth/register_usecase.go` — bcrypt hashing, validation, DB write
- [ ] Implement `usecase/auth/login_usecase.go` — credential verification, JWT (RS256) issuance
- [ ] Implement `pkg/jwt/jwt.go` — RS256 token generation and validation (`golang-jwt/jwt` v5)
- [ ] Implement Redis session store — `SET session:<id> <user_id> EX 604800` on login; `DEL` on logout
- [ ] Implement `delivery/http/handler/auth_handler.go` — POST /auth/register, /auth/login, /auth/refresh, /auth/logout
- [ ] Implement `delivery/http/middleware/auth_middleware.go` — JWT parse → Redis EXISTS session check → 401 if missing
- [ ] Implement `delivery/http/middleware/rate_limit_middleware.go`
- [ ] Write HTTP handler tests using `httptest`

#### M2.6 — User & Leaderboard HTTP Handlers
- [ ] Implement `usecase/leaderboard/fetch_leaderboard_usecase.go`
- [ ] Implement `delivery/http/handler/user_handler.go`
- [ ] Implement `delivery/http/handler/leaderboard_handler.go`
- [ ] Write handler tests

#### M2.7 — Matchmaking REST Handler
- [ ] Implement `usecase/matchmaking/enqueue_usecase.go`
- [ ] Implement `delivery/http/handler/session_handler.go`
- [ ] Write tests

**Deliverable:** Full REST API is functional. All endpoints exercisable via Postman/Bruno.

---

## 5. Phase 3 — Real-Time Engine (WebSocket + WebRTC)

**Goal:** Implement the complete WebSocket hub, room state machine, and WebRTC signaling relay.

**Prerequisites:** Phase 2 complete.

### Milestones

#### M3.1 — WebSocket Infrastructure
- [ ] Implement `delivery/ws/hub.go` — client registry, room map, register/unregister channels
- [ ] Implement `delivery/ws/client.go` — readPump, writePump, connection lifecycle
- [ ] Implement `delivery/ws/handler.go` — WebSocket upgrade + JWT validation
- [ ] Implement WebSocket event envelope parsing and routing
- [ ] Handle `PING`/`PONG` heartbeat with configurable intervals
- [ ] Implement graceful disconnect with 30-second reconnect window
- [ ] Write tests for hub registration and broadcast behavior

#### M3.2 — Matchmaking Worker
- [ ] Implement background goroutine matchmaking worker (polls every 500ms)
- [ ] Implement `usecase/matchmaking/match_usecase.go` — ZPOPMIN pair, room creation
- [ ] Broadcast `ROOM_READY` via Redis Pub/Sub to correct WS hub nodes
- [ ] Handle edge case: user disconnects while in queue (auto-dequeue on timeout)
- [ ] Write tests for matchmaking pair formation

#### M3.3 — Room State Machine
- [ ] Implement `delivery/ws/room.go` — room broadcast, state transitions
- [ ] Implement `SPIN_START` / `SPIN_RESULT` flow with deterministic seed generation
- [ ] Implement `usecase/session/spin_wheel_usecase.go`
- [ ] Implement `usecase/session/assign_scenario_usecase.go`
- [ ] Handle `SPIN_ACK` and `SCENARIO_ACK` (requires both peers to ack)
- [ ] Implement server-side prep countdown timer (goroutine per room)
- [ ] Broadcast `PREP_TIMER_TICK` every second; handle `PREP_READY` early exit
- [ ] Implement session countdown timer; broadcast `SESSION_TIMER_TICK`
- [ ] Handle `SESSION_END` from both peers OR server timer expiry
- [ ] Transition to `EVALUATING` state and trigger AI pipeline goroutine
- [ ] Write state machine transition tests

#### M3.4 — WebRTC Signaling Relay
- [ ] Implement `SIGNAL` event handler — relay offer/answer/ice to peer
- [ ] Cache SDP/ICE in Redis for late-joining peers (TTL 60s)
- [ ] Implement `pkg/webrtc/signaling.go` relay helpers
- [ ] Write integration test: simulate two WS clients completing full signaling handshake
- [ ] Configure STUN/TURN server addresses in environment config

#### M3.5 — Cross-Node Broadcasting (Multi-instance)
- [ ] Implement Redis Pub/Sub subscriber in hub for `room:<room_id>:events`
- [ ] Forward Redis messages to local WS clients
- [ ] Implement `matchmaking:new_room` subscriber
- [ ] Write test: two clients on different hub instances complete full room lifecycle

**Deliverable:** Two browser clients can be matched, spin wheels, view scenario, complete prep phase, establish WebRTC P2P call, and end the session correctly.

---

## 6. Phase 4 — AI Evaluation Pipeline

**Goal:** Full post-session AI evaluation: audio capture, STT, LLM scoring, and Elo updates.

**Prerequisites:** Phase 3 complete.

### Milestones

#### M4.1 — Client-Side Audio Capture (Frontend)
- [ ] Implement `MediaRecorder` capturing local mic stream in Next.js
- [ ] Stream audio chunks via WebSocket binary messages every 5 seconds
- [ ] Tie `MediaRecorder` lifecycle to session LIVE → EVALUATING transition

#### M4.2 — Server-Side Audio Buffer & Stereo Assembly
- [ ] Implement WebSocket binary message handler for audio chunks
- [ ] Store chunks in Redis Stream `audio:stream:<room_id>:<user_id>`
- [ ] Implement audio buffer assembly: concatenate chunks into mono `.webm` per user
- [ ] Implement ffmpeg stereo merge: User A (left) + User B (right) → `dual_channel.webm`
- [ ] Ensure ffmpeg is available in the Docker backend image

#### M4.3 — STT Integration (Deepgram Nova-2 primary / Whisper fallback)
- [ ] Implement `pkg/ai/deepgram/client.go` — Deepgram Nova-2 multichannel API wrapper
  - [ ] POST stereo WebM with `multichannel=true&channels=2&words=true`
  - [ ] Parse per-channel word arrays with timestamps
- [ ] Implement `pkg/ai/whisper/client.go` — OpenAI Whisper fallback (two sequential mono calls)
- [ ] Implement `usecase/evaluation/transcribe_usecase.go`:
  - [ ] Try Deepgram first; fall back to Whisper after 2 failed retries
  - [ ] Implement word-level interleaving: merge-sort Channel A + Channel B words by `start` timestamp
- [ ] Implement transcript quality checks (word count, confidence threshold)
- [ ] Write unit tests for interleaving merge-sort algorithm
- [ ] Write integration test with mock Deepgram response fixture

#### M4.4 — LLM Evaluation (Gemini 1.5 Flash / Pro)
- [ ] Implement `pkg/ai/gemini/client.go` — Gemini client using `google.golang.org/genai` SDK
  - [ ] Configure `ResponseMIMEType: "application/json"` + `ResponseSchema`
  - [ ] Implement model routing: Flash for easy/medium → Pro for hard or Flash validation failure
- [ ] Implement `usecase/evaluation/evaluate_usecase.go` — parallel dual evaluation (goroutines)
- [ ] Implement JSON output validation against rubric schema (section 4.3 of AI_EVALUATION.md)
- [ ] Implement exponential backoff retry (max 3 attempts per model tier)
- [ ] Implement fallback evaluation when both Flash and Pro fail
- [ ] Write unit tests for prompt assembly with various scenario/difficulty inputs
- [ ] Write integration test with mock Gemini response fixture

#### M4.5 — Elo Calculation
- [ ] Implement `pkg/elo/calculator.go` — full formula with all K-factor modifiers
- [ ] Implement `usecase/evaluation/update_elo_usecase.go`
- [ ] Write unit tests for Elo calculator (all K-factor combinations)
- [ ] Verify Elo conservation (sum of deltas near 0 in balanced scenarios)

#### M4.6 — Pipeline Orchestration
- [ ] Implement evaluation worker goroutine triggered by `SESSION_END`
- [ ] Implement idempotency check
- [ ] Implement `evaluation:complete` Redis Pub/Sub publish
- [ ] Subscribe WS hub to `evaluation:complete`; broadcast `EVALUATION_READY`
- [ ] Delete audio buffers from Redis post-evaluation
- [ ] Write end-to-end test: session → evaluation → Elo update → WS broadcast

**Deliverable:** After a completed session, both users receive full evaluation with rubric scores, feedback, and updated Elo ratings within 90 seconds.

---

## 7. Phase 5 — Frontend Implementation

**Goal:** Complete Next.js 14 frontend application.

**Prerequisites:** Phase 3 complete (API and WS contracts finalized).

### Milestones

#### M5.1 — Foundation & Auth
- [ ] Set up Next.js 14 App Router project structure
- [ ] Set up Zustand stores: `auth.store.ts`, `matchmaking.store.ts`, `room.store.ts`, `session.store.ts`, `evaluation.store.ts`, `webrtc.store.ts`
- [ ] Implement auth pages: `/login`, `/register`
- [ ] Implement `useAuth` hook backed by Zustand `auth.store.ts`
- [ ] Implement JWT access token in-memory management + auto-refresh via `POST /auth/refresh`
- [ ] Implement Next.js middleware for protected route redirects
- [ ] Set up `axios` API client with auth header interceptors + 401 auto-refresh interceptor

#### M5.2 — WebSocket Client
- [ ] Implement `hooks/useWebSocket.ts` — singleton WS connection with exponential-backoff reconnect
- [ ] Dispatch incoming server events to the correct Zustand store (room, session, evaluation)
- [ ] Map all 10+ server→client events to Zustand state transitions
- [ ] Implement PING heartbeat (every 30s), surface latency to UI

#### M5.3 — Matchmaking UI
- [ ] Build `/play` page — matchmaking entry point
- [ ] Implement difficulty + context preference selector
- [ ] Show real-time queue position from WS updates
- [ ] Animate waiting state; transition to room on `ROOM_READY`

#### M5.4 — Room & Spin UI
- [ ] Build `/room/[room_id]` page — main game room
- [ ] Implement dual spinning wheel animation driven by `spin_seed`
- [ ] Reveal role context and role assignment after spin
- [ ] Display scenario brief and private objective after `SCENARIO_ASSIGN`

#### M5.5 — Prep Phase UI
- [ ] Display `PREP_TIMER_TICK` countdown with visual urgency (color shift < 30s)
- [ ] Display "Peer is ready" indicator when `peer_ready = true`
- [ ] "I'm Ready" button → sends `PREP_READY`

#### M5.6 — WebRTC Video/Audio UI
- [ ] Implement WebRTC PeerConnection management hook (`useWebRTC`)
- [ ] Handle offer creation (initiator), answer creation (receiver)
- [ ] Handle ICE candidate gathering and relay via SIGNAL events
- [ ] Display local and remote video/audio streams
- [ ] Display session countdown timer
- [ ] Mute/unmute and camera toggle controls
- [ ] "End Session" button → sends `SESSION_END`

#### M5.7 — Evaluation Results UI
- [ ] Build evaluation results overlay on `EVALUATION_READY`
- [ ] Display overall score with animated counter
- [ ] Display rubric dimension radar/spider chart
- [ ] Display Elo delta with win/loss indicator
- [ ] Show strengths and improvement areas
- [ ] Link to `/history/:session_id` for full detail

#### M5.8 — Profile & Leaderboard UI
- [ ] Build `/profile` page — user stats, session history list
- [ ] Build `/leaderboard` page — paginated global leaderboard
- [ ] Build `/history/[session_id]` — detailed session evaluation view

**Deliverable:** Full end-to-end user experience is functional in the browser.

---

## 8. Phase 6 — Testing, Observability & Deployment

**Goal:** Production-ready hardening, monitoring, and deployment.

**Prerequisites:** Phase 5 complete.

### Milestones

#### M6.1 — Test Coverage
- [ ] Achieve >= 80% unit test coverage on all Go packages
- [ ] Full integration test suite for all API endpoints
- [ ] WebSocket end-to-end test suite (minimum 10 scenarios)
- [ ] Load test with `k6`: simulate 200 concurrent rooms; measure p99 latency
- [ ] Chaos test: kill Redis mid-session; verify graceful degradation
- [ ] Security scan with `gosec` and `npm audit`

#### M6.2 — Observability
- [ ] Instrument all HTTP handlers with OpenTelemetry traces
- [ ] Instrument WebSocket events with trace spans
- [ ] Expose Prometheus metrics: active rooms, queue depth, eval latency, WS connections
- [ ] Set up Grafana dashboards: system health, game metrics, Elo distribution
- [ ] Configure structured JSON logging with `zerolog` or `slog`
- [ ] Set up distributed tracing with Jaeger or Google Cloud Trace

#### M6.3 — Infrastructure
- [ ] Write production `docker-compose.yml` with all services
- [ ] Write Kubernetes manifests: Deployment, Service, HPA, ConfigMap, Secret
- [ ] Set up PgBouncer for PostgreSQL connection pooling
- [ ] Set up Redis Sentinel or Redis Cluster for HA
- [ ] Configure Coturn TURN server
- [ ] Set up NGINX with TLS termination and WebSocket proxy config
- [ ] Configure CDN for frontend static assets

#### M6.4 — CI/CD
- [ ] GitHub Actions: lint → test → build → push Docker image
- [ ] Staging environment: automatic deploy on `main` branch push
- [ ] Production environment: manual gate + canary deploy
- [ ] Database migration step in deployment pipeline

#### M6.5 — Security Hardening
- [ ] Penetration test checklist: OWASP Top 10
- [ ] Rate limiting audit
- [ ] JWT secret rotation procedure documented
- [ ] GDPR compliance checklist: data retention, right-to-erasure endpoint
- [ ] Dependency vulnerability scan (Dependabot / Snyk)

**Deliverable:** System is deployed to production, monitored, and ready for public beta.

---

## 9. Milestone Summary Table

| Phase | Milestone | Description                                    | Est. Duration |
|-------|-----------|------------------------------------------------|---------------|
| 1     | M1.1      | Documentation complete                         | 1 day         |
| 1     | M1.2      | Repository bootstrap                           | 1 day         |
| 2     | M2.1      | Domain layer + entities                        | 3 days        |
| 2     | M2.2      | Database migrations + seed data                | 2 days        |
| 2     | M2.3      | PostgreSQL repositories                        | 3 days        |
| 2     | M2.4      | Redis repositories                             | 2 days        |
| 2     | M2.5      | Auth use cases + HTTP handlers                 | 3 days        |
| 2     | M2.6      | User & leaderboard handlers                    | 2 days        |
| 2     | M2.7      | Matchmaking REST handlers                      | 1 day         |
| 3     | M3.1      | WebSocket hub + client infrastructure          | 3 days        |
| 3     | M3.2      | Matchmaking worker (real-time)                 | 2 days        |
| 3     | M3.3      | Room state machine                             | 4 days        |
| 3     | M3.4      | WebRTC signaling relay                         | 2 days        |
| 3     | M3.5      | Cross-node broadcasting (Redis Pub/Sub)        | 2 days        |
| 4     | M4.1      | Client-side audio capture (FE)                 | 1 day         |
| 4     | M4.2      | Server-side audio buffer (Redis Streams)       | 1 day         |
| 4     | M4.3      | STT integration + transcript processing        | 3 days        |
| 4     | M4.4      | LLM evaluation + JSON validation               | 3 days        |
| 4     | M4.5      | Elo calculation                                | 1 day         |
| 4     | M4.6      | Pipeline orchestration + WS broadcast          | 2 days        |
| 5     | M5.1      | Next.js foundation + auth                      | 2 days        |
| 5     | M5.2      | WebSocket client                               | 2 days        |
| 5     | M5.3      | Matchmaking UI                                 | 1 day         |
| 5     | M5.4      | Room & spin UI                                 | 2 days        |
| 5     | M5.5      | Prep phase UI                                  | 1 day         |
| 5     | M5.6      | WebRTC video/audio UI                          | 4 days        |
| 5     | M5.7      | Evaluation results UI                          | 2 days        |
| 5     | M5.8      | Profile & leaderboard UI                       | 2 days        |
| 6     | M6.1      | Test coverage + load testing                   | 3 days        |
| 6     | M6.2      | Observability stack                            | 2 days        |
| 6     | M6.3      | Infrastructure (Docker, K8s, TURN, NGINX)      | 3 days        |
| 6     | M6.4      | CI/CD pipeline                                 | 2 days        |
| 6     | M6.5      | Security hardening                             | 2 days        |
|       | **Total** | **All phases**                                 | **~60 days**  |

---

## 10. Risk Register

| Risk                                      | Likelihood | Impact | Mitigation                                                        |
|-------------------------------------------|------------|--------|-------------------------------------------------------------------|
| WebRTC NAT traversal failure rate > 10%   | Medium     | High   | Deploy Coturn TURN server; test diverse network topologies        |
| STT transcription quality < 70% accuracy  | Medium     | High   | Test varied accents/audio quality; tune Whisper temperature       |
| LLM hallucinated rubric scores            | Medium     | High   | Strict JSON schema validation; fallback score on parse failure    |
| Redis memory exhaustion under load        | Low        | High   | Set `maxmemory-policy allkeys-lru`; monitor with Prometheus       |
| Matchmaking queue imbalance (odd users)   | High       | Low    | Re-enqueue unpaired users; add wait-time Elo relaxation           |
| WebSocket connection drops on mobile      | High       | Medium | Implement reconnect with exponential backoff; 30s window          |
| LLM API rate limits / cost overrun        | Medium     | Medium | Implement caching for identical scenarios; budget alerts          |
| Audio sync drift between channels         | Low        | Medium | Timestamp-based merge; allow +-200ms alignment tolerance          |
| Cross-instance WS message delivery loss   | Low        | High   | Use Redis Streams (not just Pub/Sub) for at-least-once delivery   |
| Elo inflation / rating deflation          | Medium     | Medium | Monitor Elo distribution weekly; apply deflation correction       |
