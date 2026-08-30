# AGENT.md — Master Project Context

> **Real-Time Scenario-Based Roleplay Simulator**
> Last Updated: 2026-08-30 | Phase: 1 — Architecture & Documentation (Stack LOCKED)

---

## Project Purpose

A high-concurrency, peer-to-peer (P2P) workplace roleplay platform that enables candidates and students to master **negotiation**, **communication**, and **conflict-resolution** skills through structured, timed, AI-evaluated roleplay sessions with real human counterparts.

---

## Core Workflow (End-to-End)

```
[User A] ──┐
           ├──► Matchmaking Queue (Redis) ──► Room Created
[User B] ──┘
                      │
                      ▼
            Synchronized Role Spin
        (Dual Wheels: Context + Role Pair)
                      │
                      ▼
          Dynamic Scenario Assignment
           (Easy / Medium / Hard tier)
                      │
                      ▼
           Prep Phase (2–5 min timer)
                      │
                      ▼
      Live WebRTC Audio/Video Session (5–6 min)
                      │
                      ▼
     Dual-Channel Audio Capture + Transcription
     (Deepgram Nova-2 multichannel / Whisper fallback)
                      │
                      ▼
     LLM Rubric Evaluation → Elo Score Update
     (Gemini 1.5 Flash → Pro escalation)
                      │
                      ▼
              Leaderboard & Feedback
```

---

## Technology Stack (LOCKED — 2026-08-30)

> Full rationale and dependency versions: see `docs/TECH_STACK.md`

| Layer                    | Technology                               |
|--------------------------|------------------------------------------|
| Backend language         | Go 1.23+                                 |
| HTTP router              | `go-chi/chi` v5                          |
| WebSocket library        | `nhooyr.io/websocket` v1.8.x             |
| Database driver          | `jackc/pgx` v5                           |
| SQL code generation      | `sqlc` v1.27+                            |
| DB migration runner      | `golang-migrate/migrate` v4              |
| Auth strategy            | JWT (RS256) + Redis session store        |
| JWT library              | `golang-jwt/jwt` v5                      |
| Redis client             | `go-redis/redis` v9                      |
| Primary database         | PostgreSQL 16                            |
| Cache / Queue / Pub-Sub  | Redis 7                                  |
| STT provider (primary)   | Deepgram Nova-2 (multichannel)           |
| STT provider (fallback)  | OpenAI Whisper API (whisper-1)           |
| LLM evaluation (primary) | Google Gemini 1.5 Pro / Flash            |
| Frontend framework       | Next.js 14 App Router + TypeScript       |
| Styling                  | Tailwind CSS v3                          |
| Frontend state           | Zustand v4                               |
| Reverse proxy            | Caddy v2 (auto-TLS)                      |
| Deployment (MVP)         | Docker Compose on Linux VPS              |
| Structured logging       | `rs/zerolog`                             |
| Observability            | OpenTelemetry → Prometheus + Grafana     |

---

## Repository Layout (Planned)

```
/
├── AGENT.md                        ← This file
├── docs/
│   ├── TECH_STACK.md               ← Finalized stack, dep versions, decision log
│   ├── ARCHITECTURE.md             ← System design, diagrams, folder structure
│   ├── DATABASE.md                 ← PostgreSQL schema + Redis data models
│   ├── API_SPECS.md                ← REST + WebSocket contracts
│   ├── AI_EVALUATION.md            ← STT pipeline, LLM prompts, Elo formula
│   └── ROADMAP.md                  ← Phased implementation milestones
├── cmd/
│   └── server/
│       └── main.go                 ← Binary entrypoint
├── internal/
│   ├── domain/                     ← Entities, value objects, domain errors
│   ├── usecase/                    ← Application business logic
│   ├── repository/
│   │   ├── postgres/
│   │   │   ├── db/                 ← sqlc-generated code (DO NOT EDIT MANUALLY)
│   │   │   └── queries/            ← Hand-written .sql query files
│   │   └── redis/                  ← Redis repository implementations
│   └── delivery/
│       ├── http/                   ← chi router + REST handlers
│       └── ws/                     ← WebSocket hub & handlers (nhooyr.io/websocket)
├── pkg/
│   ├── webrtc/                     ← Signaling helpers
│   ├── ai/
│   │   ├── deepgram/               ← Deepgram Nova-2 STT client
│   │   ├── whisper/                ← OpenAI Whisper fallback client
│   │   └── gemini/                 ← Gemini 1.5 Pro/Flash evaluation client
│   ├── elo/                        ← Rating calculation utilities
│   ├── jwt/                        ← RS256 token generation & validation
│   └── validator/
├── migrations/                     ← golang-migrate SQL files
├── sqlc.yaml                       ← sqlc configuration
├── frontend/                       ← Next.js 14 App Router
│   ├── app/
│   ├── components/
│   ├── hooks/
│   │   ├── useWebSocket.ts         ← WS singleton + event dispatcher
│   │   └── useWebRTC.ts            ← RTCPeerConnection manager
│   ├── stores/                     ← Zustand stores (auth, room, session, etc.)
│   └── lib/
├── docker-compose.yml
├── Caddyfile
├── Makefile
└── .env.example
```

---

## Phase Overview

| Phase | Focus                                        | Status          |
|-------|----------------------------------------------|-----------------|
| 1     | Architecture, Documentation, Stack Lock      | ✅ Complete      |
| 2     | Backend Core (Domain, Auth, Matchmaking REST)| 🔒 Pending      |
| 3     | WebRTC Signaling + Session State Machine     | 🔒 Pending      |
| 4     | AI Pipeline (Deepgram STT + Gemini Eval)     | 🔒 Pending      |
| 5     | Frontend Implementation (Next.js + Zustand)  | 🔒 Pending      |
| 6     | Testing, Observability, Docker Compose Deploy| 🔒 Pending      |

---

## Key Design Constraints

- **Concurrency:** Target 1,000+ simultaneous active rooms.
- **Latency:** WebSocket event round-trip < 100ms; signaling < 200ms.
- **Isolation:** Each room is a fully isolated state machine — no shared mutable state across rooms.
- **Fairness:** Scenario assignment must be statistically uniform across difficulty tiers.
- **Evaluation Integrity:** Audio capture and transcription happen server-side after session ends.
- **GDPR / Privacy:** Raw audio deleted immediately after Deepgram transcription; only text transcripts retained (90 days).
- **Auth Revocation:** All sessions tracked in Redis; logout takes effect within one request cycle.

---

## Agent Instructions

> When implementing any phase, agents MUST:
> 1. Reference `docs/TECH_STACK.md` for exact package names and versions.
> 2. Reference `docs/ARCHITECTURE.md` for folder structure and component boundaries.
> 3. Reference `docs/DATABASE.md` before writing any repository layer code or SQL.
> 4. Reference `docs/API_SPECS.md` before writing any HTTP handler or WebSocket handler.
> 5. Reference `docs/AI_EVALUATION.md` before implementing the AI pipeline.
> 6. Follow Clean Architecture dependency rules: **Domain ← UseCase ← Repository/Delivery**.
> 7. Write tests alongside each component (unit + integration).
> 8. Never write raw SQL inside use-case or domain layers — use sqlc-generated repository functions.
> 9. `internal/repository/postgres/db/` is sqlc-generated — NEVER edit files in that directory manually.
> 10. Use `zerolog` for all logging — never `fmt.Println` in production code paths.
