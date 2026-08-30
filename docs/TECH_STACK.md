# Finalized Technology Stack

> **Project:** Real-Time Scenario-Based Roleplay Simulator
> **Version:** 1.1.0 | **Status:** LOCKED — Phase 1 finalized 2026-08-30
> All implementation work MUST use the versions and packages defined in this document.

---

## Table of Contents

1. [Stack Decision Summary](#1-stack-decision-summary)
2. [Backend — Go Services](#2-backend--go-services)
3. [Frontend — Next.js](#3-frontend--nextjs)
4. [Data Layer](#4-data-layer)
5. [AI Services](#5-ai-services)
6. [Infrastructure & Deployment](#6-infrastructure--deployment)
7. [Go Module Dependency Reference](#7-go-module-dependency-reference)
8. [Frontend Package Reference](#8-frontend-package-reference)
9. [Decision Log](#9-decision-log)

---

## 1. Stack Decision Summary

| Layer                   | Technology                                      | Version / Tag          |
|-------------------------|-------------------------------------------------|------------------------|
| Backend language        | Go                                              | 1.23+                  |
| HTTP router             | `go-chi/chi`                                    | v5                     |
| WebSocket library       | `nhooyr.io/websocket`                           | v1.8.x                 |
| Database driver         | `jackc/pgx`                                     | v5                     |
| SQL codegen             | `sqlc`                                          | v1.27+                 |
| DB migration runner     | `golang-migrate/migrate`                        | v4                     |
| JWT                     | `golang-jwt/jwt`                                | v5                     |
| Redis client            | `redis/go-redis`                                | v9                     |
| Auth strategy           | JWT (RS256) + Redis session store               | —                      |
| Primary database        | PostgreSQL                                      | 16                     |
| Cache / Queue / Pub-Sub | Redis                                           | 7.x                    |
| STT provider (primary)  | Deepgram Nova-2                                 | API v1                 |
| STT provider (fallback) | OpenAI Whisper                                  | whisper-1              |
| LLM evaluation (primary)| Google Gemini 1.5 Pro                           | gemini-1.5-pro         |
| LLM evaluation (fast)   | Google Gemini 1.5 Flash                         | gemini-1.5-flash       |
| Frontend framework      | Next.js (App Router) + TypeScript               | 14+                    |
| Styling                 | Tailwind CSS                                    | v3                     |
| Frontend state          | Zustand                                         | v4                     |
| Deployment (MVP)        | Docker Compose on single Linux VPS              | Docker 26+             |
| Reverse proxy           | Caddy                                           | v2 (auto-TLS)          |
| Observability           | OpenTelemetry → Prometheus + Grafana            | —                      |
| Container runtime       | Docker Engine                                   | 26+                    |

---

## 2. Backend — Go Services

### 2.1 HTTP Router — `go-chi/chi` v5

**Rationale:** Fully compatible with `net/http` stdlib, zero external dependencies, composable middleware, lightweight. No lock-in — any `http.Handler` works directly.

```
Pattern:
  r := chi.NewRouter()
  r.Use(middleware.Logger)
  r.Use(middleware.Recoverer)
  r.Use(authMiddleware)
  r.Post("/auth/login", handler.Login)
  r.Route("/users", func(r chi.Router) {
    r.Get("/me", handler.GetMe)
    r.Patch("/me", handler.UpdateMe)
  })
```

**Key middleware used:**
- `chi/middleware.Logger` — structured request logging
- `chi/middleware.Recoverer` — panic recovery
- `chi/middleware.RealIP` — real client IP behind NGINX
- `chi/middleware.Compress` — gzip response compression
- Custom: `JWTAuth`, `RateLimit`, `CORS`

---

### 2.2 WebSocket Library — `nhooyr.io/websocket` v1.8.x

**Rationale:** Actively maintained (unlike archived `gorilla/websocket`), context-aware connection lifecycle (clean cancellation), idiomatic Go API, supports binary messages for audio chunk streaming.

```
Upgrade pattern:
  conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
      OriginPatterns: []string{"roleplay-sim.io"},
  })

Read/Write with context:
  ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
  defer cancel()
  _, msg, err := conn.Read(ctx)
  conn.Write(ctx, websocket.MessageText, payload)
```

---

### 2.3 Database Access — `sqlc` + `pgx/v5`

**Rationale:** Write plain SQL (already spec'd in `DATABASE.md`), run `sqlc generate`, receive fully type-safe Go structs and query functions. Zero ORM magic, zero N+1 risk, full PostgreSQL feature access.

**Workflow:**
```
1. Write SQL queries in queries/*.sql
2. Run: sqlc generate
3. Get: db/query.sql.go  (type-safe functions)
        db/models.go     (struct types)

Example generated call:
  user, err := q.GetUserByEmail(ctx, email)
  session, err := q.CreateSession(ctx, db.CreateSessionParams{
      ScenarioID: scenarioID,
      RoomID:     roomID,
      Difficulty: db.DifficultyMedium,
  })
```

**sqlc.yaml configuration:**
```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/repository/postgres/queries/"
    schema: "migrations/"
    gen:
      go:
        package: "db"
        out: "internal/repository/postgres/db"
        emit_json_tags: true
        emit_pointers_for_null_fields: true
        emit_enum_valid_method: true
```

---

### 2.4 Authentication — JWT (RS256) + Redis Session Store

**Rationale:** Pure stateless JWT is fast but cannot be revoked before expiry. Redis session store adds a lookup step (~1ms) on every request to enable instant revocation, active session tracking, and multi-device sign-out.

**Flow:**
```
Login:
  1. Validate credentials
  2. Generate JWT (RS256, 15-min expiry, claims: user_id, role, session_id)
  3. Store session in Redis: SET session:<session_id> <user_id> EX 604800 (7 days)
  4. Return access_token + Set-Cookie: refresh_token (HttpOnly, SameSite=Strict)

Auth Middleware (every request):
  1. Parse JWT → extract session_id
  2. EXISTS session:<session_id> in Redis → if 0, reject 401
  3. If valid, inject user context into request

Logout / Revocation:
  1. DEL session:<session_id> from Redis
  2. Clear refresh_token cookie
```

**Redis session key schema:**
```
Key:   session:<session_id>           TTL: 7 days (refresh window)
Value: <user_id>

Key:   user:sessions:<user_id>        TTL: 7 days
Value: SET of session_ids             (for "sign out all devices")
```

---

### 2.5 Redis Client — `go-redis/redis` v9

**Rationale:** Most actively maintained Go Redis client, supports Redis 7 streams, pipelines, pub/sub, and all data structures used in the architecture spec.

---

### 2.6 Migration Runner — `golang-migrate/migrate` v4

**Rationale:** Standard, battle-tested. Supports PostgreSQL via `pgx`. Simple CLI usage: `migrate -path migrations/ -database $DATABASE_URL up`.

---

## 3. Frontend — Next.js

### 3.1 Framework — Next.js 14 App Router + TypeScript

- **App Router** for all routes (no Pages Router)
- **Server Components** for data-fetching pages (profile, leaderboard, history)
- **Client Components** (`"use client"`) for all interactive UI (room, spin, WebRTC)

### 3.2 State Management — Zustand v4

**Rationale:** Minimal boilerplate, no Provider wrapping, works perfectly with Next.js App Router Client Components, easy to slice per domain.

**Store slices:**
```typescript
// stores/auth.store.ts     — user identity, token, login/logout
// stores/matchmaking.store.ts — queue status, room_id
// stores/room.store.ts     — room state machine (mirrors server FSM)
// stores/session.store.ts  — prep timer, session timer, spin result
// stores/evaluation.store.ts — scores, rubric results
// stores/webrtc.store.ts   — peer connection state, stream refs
```

### 3.3 WebSocket Hook

A singleton `useWebSocket` hook wraps the WS connection, auto-reconnects, and dispatches incoming events to Zustand stores:

```typescript
// hooks/useWebSocket.ts
// Connects once on mount, reconnects with exponential backoff
// Dispatches events to appropriate Zustand stores
// Provides send(event, payload) helper
```

### 3.4 WebRTC Hook

```typescript
// hooks/useWebRTC.ts
// Manages RTCPeerConnection lifecycle
// Handles offer/answer/ICE via useWebSocket's SIGNAL channel
// Exposes localStream, remoteStream, connectionState
```

---

## 4. Data Layer

### 4.1 PostgreSQL 16

- Driver: `pgx/v5` (direct, no ORM)
- Query generation: `sqlc`
- Connection pooling: `pgxpool` (built into pgx/v5)
- Pool config: `MaxConns: runtime.NumCPU() * 4`, `MinConns: 2`

### 4.2 Redis 7

- Client: `go-redis/redis` v9
- Use cases: matchmaking queue, room state hash, WebRTC signaling cache,
  audio stream buffer, rate limiting, leaderboard sorted set, pub/sub,
  **JWT session store**

---

## 5. AI Services

### 5.1 STT — Deepgram Nova-2 (Primary)

**Rationale:** Nova-2 natively supports **multichannel audio** in a single API call, returning per-channel transcripts with word-level timestamps. This eliminates the need for two separate STT calls and simplifies the transcript merge step. Accuracy is on par with Whisper for English speech.

**API Endpoint:** `POST https://api.deepgram.com/v1/listen`

**Request configuration:**
```
Authorization: Token <DEEPGRAM_API_KEY>
Content-Type: audio/webm

Query parameters:
  model=nova-2
  multichannel=true         ← key feature: per-channel transcription
  channels=2                ← assembled dual-channel webm
  punctuate=true
  utterances=true
  words=true                ← word-level timestamps
  language=en
  smart_format=true
```

**Response structure:**
```json
{
  "results": {
    "channels": [
      {
        "alternatives": [{
          "transcript": "I understand the release pressure...",
          "words": [
            { "word": "I", "start": 0.0, "end": 0.12, "confidence": 0.99 },
            { "word": "understand", "start": 0.14, "end": 0.82, "confidence": 0.98 }
          ]
        }]
      },
      {
        "alternatives": [{
          "transcript": "I appreciate that. What exactly do you need?",
          "words": [
            { "word": "I", "start": 0.35, "end": 0.42, "confidence": 0.99 }
          ]
        }]
      }
    ]
  }
}
```

**Key advantage:** Single API call returns both channels. The server assembles the two mono WebM streams into a stereo WebM (left = User A, right = User B) before sending to Deepgram.

**Audio assembly for dual-channel:**
```
User A mono (audio:stream:<room_id>:<user_a_id>) ──┐
                                                    ├──► ffmpeg merge → stereo webm → Deepgram
User B mono (audio:stream:<room_id>:<user_b_id>) ──┘
```

**ffmpeg merge command (server-side):**
```bash
ffmpeg -i user_a.webm -i user_b.webm \
  -filter_complex "[0:a][1:a]amerge=inputs=2,pan=stereo|c0=c0|c1=c1[out]" \
  -map "[out]" -c:a libopus dual_channel.webm
```

---

### 5.2 STT — OpenAI Whisper (Fallback)

Used if Deepgram API is unavailable or returns an error after 2 retries.

**Endpoint:** `POST https://api.openai.com/v1/audio/transcriptions`

Whisper does not support multichannel — two separate sequential calls are made (one per mono channel), then interleaved.

```
model: whisper-1
response_format: verbose_json
timestamp_granularities: ["word", "segment"]
language: en
temperature: 0
```

---

### 5.3 LLM Evaluation — Google Gemini 1.5 Pro / Flash

**Rationale:** Gemini supports structured JSON output via `responseMimeType: "application/json"` with a provided `responseSchema`. This enforces the rubric JSON schema at the API level — the model cannot output non-conforming responses.

**Model selection strategy:**
- **Gemini 1.5 Flash** (`gemini-1.5-flash`): Used for straightforward evaluations (easy/medium difficulty). Faster (~3–5s), cheaper.
- **Gemini 1.5 Pro** (`gemini-1.5-pro`): Used for hard difficulty or when Flash output fails JSON validation. Better reasoning on complex transcripts.

**Go SDK:** `google.golang.org/genai` (official Google Gen AI Go SDK)

**API call configuration:**
```go
config := &genai.GenerateContentConfig{
    ResponseMIMEType: "application/json",
    ResponseSchema:   rubricOutputSchema, // *genai.Schema matching JSON schema
    Temperature:      genai.Ptr[float32](0.2),
    TopP:             genai.Ptr[float32](0.8),
    MaxOutputTokens:  2048,
}

result, err := client.Models.GenerateContent(
    ctx,
    "gemini-1.5-flash",    // or "gemini-1.5-pro" for hard tier
    genai.Text(userPrompt),
    config,
)
```

**Model routing logic:**
```
if difficulty == "hard" OR previous_flash_attempt_failed:
    use gemini-1.5-pro
else:
    use gemini-1.5-flash

if gemini_pro_fails (after 3 retries):
    apply fallback evaluation (score=50, no Elo change)
```

**`llm_model_used` field values (in `evaluations` table):**
- `"gemini-1.5-flash-001"`
- `"gemini-1.5-pro-001"`

---

## 6. Infrastructure & Deployment

### 6.1 Deployment MVP — Docker Compose on Linux VPS

**Rationale:** Simplest possible production topology. A single VPS with Docker Compose handles all services. Caddy provides automatic TLS via Let's Encrypt. No Kubernetes complexity at MVP stage.

**Recommended VPS spec:**
```
Provider:  Hetzner CX32 (or equivalent)
CPU:       4 vCPU
RAM:       8 GB
Disk:      80 GB SSD
OS:        Ubuntu 22.04 LTS
Location:  Choose closest to target user region
```

**`docker-compose.yml` services:**
```yaml
services:
  caddy:          # Reverse proxy + auto-TLS (port 80/443)
  backend:        # Go binary (port 8080 internal)
  frontend:       # Next.js (port 3000 internal, served via Caddy)
  postgres:       # PostgreSQL 16 (port 5432 internal only)
  redis:          # Redis 7 (port 6379 internal only)
```

**Port exposure:**
- `80/443` → Caddy (public)
- All other ports → internal Docker network only (never exposed to host)

### 6.2 Reverse Proxy — Caddy v2

**Rationale:** Caddy provides automatic TLS certificate provisioning and renewal via Let's Encrypt with zero configuration. No need to manage Certbot or NGINX TLS config manually.

**Caddyfile:**
```
roleplay-sim.io {
    reverse_proxy /v1/ws* backend:8080   # WebSocket pass-through
    reverse_proxy /v1/*   backend:8080   # REST API
    reverse_proxy *       frontend:3000  # Next.js
}
```

WebSocket upgrade is handled automatically by Caddy's reverse proxy.

### 6.3 TURN Server — Coturn

```yaml
# In docker-compose.yml
  coturn:
    image: coturn/coturn:latest
    ports:
      - "3478:3478/udp"   # STUN/TURN
      - "3478:3478/tcp"
      - "5349:5349/udp"   # TURN over TLS
      - "5349:5349/tcp"
    environment:
      TURN_SECRET: ${TURN_SECRET}
```

### 6.4 Environment Variables

```bash
# Backend
DATABASE_URL=postgresql://user:pass@postgres:5432/roleplay_db?sslmode=disable
REDIS_URL=redis://redis:6379
JWT_PRIVATE_KEY_PATH=/run/secrets/jwt_private.pem
JWT_PUBLIC_KEY_PATH=/run/secrets/jwt_public.pem
DEEPGRAM_API_KEY=<secret>
GEMINI_API_KEY=<secret>
OPENAI_API_KEY=<secret>          # whisper fallback only
TURN_SECRET=<secret>
APP_ENV=production
LOG_LEVEL=info
PORT=8080

# Frontend
NEXT_PUBLIC_API_URL=https://roleplay-sim.io/v1
NEXT_PUBLIC_WS_URL=wss://roleplay-sim.io/v1/ws
NEXT_PUBLIC_TURN_URL=turn:roleplay-sim.io:3478
```

### 6.5 Observability

| Component      | Tool                                     |
|----------------|------------------------------------------|
| Instrumentation| OpenTelemetry Go SDK                     |
| Tracing        | OTEL → Grafana Tempo                     |
| Metrics        | `prometheus/client_golang` → Grafana     |
| Logging        | `rs/zerolog` (structured JSON)           |
| Dashboards     | Grafana Cloud free tier                  |

---

## 7. Go Module Dependency Reference

```
go.mod — primary dependencies

github.com/go-chi/chi/v5          v5.x.x    HTTP router
nhooyr.io/websocket               v1.8.x    WebSocket
github.com/jackc/pgx/v5           v5.x.x    PostgreSQL driver + pgxpool
github.com/go-redis/redis/v9      v9.x.x    Redis client
github.com/golang-jwt/jwt/v5      v5.x.x    JWT RS256
github.com/golang-migrate/migrate/v4 v4.x.x DB migrations
github.com/google/uuid            v1.x.x    UUID generation
google.golang.org/genai           v0.x.x    Gemini AI SDK
github.com/rs/zerolog             v1.x.x    Structured logging
go.opentelemetry.io/otel          v1.x.x    Observability
github.com/prometheus/client_golang v1.x.x  Metrics
github.com/sethvargo/go-envconfig  v0.x.x  Env config parsing
github.com/stretchr/testify       v1.x.x    Test assertions
github.com/testcontainers/testcontainers-go v0.x.x Integration tests
github.com/deepgram-sdk/deepgram-go-sdk v1.x.x Deepgram STT

dev tools (not in go.mod):
sqlc        v1.27+    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
golangci-lint          Static analysis
```

---

## 8. Frontend Package Reference

```
package.json — primary dependencies

next                     14.x.x    Framework
react / react-dom        18.x.x    UI runtime
typescript               5.x.x     Type safety
tailwindcss              3.x.x     Styling
zustand                  4.x.x     State management
@tanstack/react-query    5.x.x     Server state / API caching
axios                    1.x.x     HTTP client
react-hook-form          7.x.x     Form management
zod                      3.x.x     Schema validation (forms + API responses)
recharts                 2.x.x     Radar chart for rubric scores
date-fns                 3.x.x     Date formatting
clsx / tailwind-merge    latest    Conditional class names

dev dependencies:
eslint + @typescript-eslint
prettier
husky + lint-staged
```

---

## 9. Decision Log

| Decision               | Choice                          | Rejected Alternatives              | Rationale                                                       |
|------------------------|---------------------------------|------------------------------------|-----------------------------------------------------------------|
| HTTP Router            | `chi` v5                        | `gin`, `echo`, `fiber`             | Stdlib-compatible, zero deps, clean routing                     |
| WebSocket              | `nhooyr.io/websocket`           | `gorilla/websocket` (archived)     | Actively maintained, context-aware lifecycle                    |
| DB Access              | `sqlc` + `pgx/v5`              | `gorm`, raw `pgx`                  | Type-safe generated code from SQL, zero ORM overhead            |
| Auth Strategy          | JWT + Redis session store       | Pure stateless JWT                 | Enables instant revocation, multi-device sign-out               |
| Frontend State         | Zustand v4                      | Redux Toolkit, Jotai, Context API  | Minimal boilerplate, App Router compatible, easy store slicing  |
| STT Primary            | Deepgram Nova-2                 | Whisper API, Google STT            | Native multichannel in 1 call, word timestamps, high accuracy   |
| STT Fallback           | OpenAI Whisper                  | Google STT, AssemblyAI             | Widely adopted, simple API, proven accuracy                     |
| LLM Primary            | Gemini 1.5 Pro / Flash          | GPT-4o, Claude Sonnet              | Native JSON schema enforcement, competitive accuracy, cost      |
| LLM Model Routing      | Flash → Pro escalation          | Single model                       | Flash for speed/cost on easy/medium; Pro for hard/complex       |
| Reverse Proxy          | Caddy v2                        | NGINX + Certbot, Traefik           | Auto TLS with zero config, WebSocket support built-in           |
| Deployment MVP         | Docker Compose on VPS           | Fly.io, Railway, GKE               | Full control, cheapest, no vendor lock-in, easy to understand   |
| Migration Tool         | `golang-migrate`                | `goose`, `atlas`                   | Standard choice, pgx support, CLI + library usage               |
