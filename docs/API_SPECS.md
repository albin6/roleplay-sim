# API & WebSocket Contracts

> **Project:** Real-Time Scenario-Based Roleplay Simulator
> **Version:** 1.0.0 | **Phase:** 1 — Documentation
> **Base URL:** `https://api.roleplay-sim.io/v1`
> **WebSocket URL:** `wss://api.roleplay-sim.io/v1/ws`

---

## Table of Contents

1. [Authentication](#1-authentication)
2. [REST API Endpoints](#2-rest-api-endpoints)
3. [WebSocket Protocol](#3-websocket-protocol)
4. [Error Codes](#4-error-codes)
5. [Rate Limits](#5-rate-limits)

---

## 1. Authentication

All REST endpoints (except `/auth/register` and `/auth/login`) require a Bearer JWT in the `Authorization` header:

```
Authorization: Bearer <access_token>
```

WebSocket connections require the token as a query parameter on the upgrade request:

```
wss://api.roleplay-sim.io/v1/ws?token=<access_token>
```

**Token Lifecycle:**

| Token Type    | Expiry | Storage         |
|---------------|--------|-----------------|
| access_token  | 15 min | Memory (JS)     |
| refresh_token | 7 days | HttpOnly cookie |

---

## 2. REST API Endpoints

### 2.1 Auth

#### `POST /auth/register`

Register a new user account.

**Request Body:**
```json
{
  "username": "janedoe42",
  "email": "jane@example.com",
  "password": "Str0ng!Pass",
  "display_name": "Jane Doe"
}
```

**Validation Rules:**
- `username`: 3–50 chars, alphanumeric + underscore only
- `email`: valid RFC 5322 format
- `password`: min 8 chars, at least 1 uppercase, 1 digit, 1 special char
- `display_name`: 2–100 chars

**Response `201 Created`:**
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "janedoe42",
    "email": "jane@example.com",
    "display_name": "Jane Doe",
    "elo_rating": 1200.00,
    "created_at": "2026-08-29T21:11:27Z"
  },
  "access_token": "<jwt>",
  "expires_in": 900
}
```

---

#### `POST /auth/login`

**Request Body:**
```json
{
  "email": "jane@example.com",
  "password": "Str0ng!Pass"
}
```

**Response `200 OK`:** Same structure as register response.

---

#### `POST /auth/refresh`

Rotate access token using the HttpOnly refresh cookie. No body required.

**Response `200 OK`:**
```json
{
  "access_token": "<new_jwt>",
  "expires_in": 900
}
```

---

#### `POST /auth/logout`

Revoke refresh token. **Response `204 No Content`**

---

### 2.2 Users

#### `GET /users/me`

Fetch the authenticated user's full profile.

**Response `200 OK`:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "janedoe42",
  "display_name": "Jane Doe",
  "avatar_url": "https://cdn.roleplay-sim.io/avatars/janedoe42.webp",
  "elo_rating": 1342.50,
  "total_sessions": 27,
  "wins": 18,
  "losses": 9,
  "rank": 42,
  "created_at": "2026-01-15T09:00:00Z"
}
```

---

#### `PATCH /users/me`

Update profile fields. All fields optional.

**Request Body:**
```json
{
  "display_name": "Jane D.",
  "avatar_url": "https://example.com/avatar.webp"
}
```

**Response `200 OK`:** Updated user object.

---

#### `GET /users/me/history`

Fetch session history.

**Query Parameters:** `page` (default 1), `limit` (default 20, max 50), `sort` (`asc`/`desc`)

**Response `200 OK`:**
```json
{
  "data": [
    {
      "session_id": "...",
      "scenario_title": "Leave Request Under Deadline Pressure",
      "difficulty": "medium",
      "role_played": "Junior Developer",
      "opponent_display_name": "alex_senior",
      "overall_score": 78.50,
      "objective_achieved": true,
      "elo_delta": 12.30,
      "played_at": "2026-08-28T14:30:00Z"
    }
  ],
  "pagination": { "page": 1, "limit": 20, "total": 27, "total_pages": 2 }
}
```

---

#### `GET /users/me/history/:session_id`

Fetch detailed evaluation for a specific session.

**Response `200 OK`:**
```json
{
  "session_id": "...",
  "scenario": {
    "title": "Leave Request Under Deadline Pressure",
    "difficulty": "medium",
    "background_context": "..."
  },
  "my_role": "Junior Developer",
  "opponent_role": "Team Lead",
  "session_started_at": "2026-08-28T14:35:00Z",
  "session_ended_at": "2026-08-28T14:41:00Z",
  "evaluation": {
    "overall_score": 78.50,
    "objective_achieved": true,
    "summary_feedback": "You demonstrated strong preparation...",
    "strengths": ["Clear framing of your request"],
    "areas_for_improvement": ["Could have offered a compromise earlier"],
    "rubric_scores": [
      {
        "dimension": "communication_clarity",
        "score": 8.2,
        "weight": 0.20,
        "justification": "Your opening statement was concise and well-framed.",
        "evidence_quotes": ["'I understand the release pressure...'"]
      }
    ]
  },
  "elo_before": 1330.20,
  "elo_after": 1342.50,
  "elo_delta": 12.30
}
```

---

### 2.3 Matchmaking

#### `POST /matchmaking/enqueue`

**Request Body:**
```json
{
  "preferred_difficulty": "medium",
  "preferred_context": "it-team"
}
```

**Response `200 OK`:**
```json
{
  "status": "queued",
  "position": 3,
  "estimated_wait_seconds": 15
}
```

---

#### `DELETE /matchmaking/dequeue`

Cancel matchmaking. **Response `204 No Content`**

---

#### `GET /matchmaking/status`

**Response `200 OK`:**
```json
{
  "status": "queued",
  "position": 2,
  "room_id": null
}
```

`status` values: `"queued"` | `"matched"` | `"not_queued"`

---

### 2.4 Sessions

#### `GET /sessions/:session_id`

**Response `200 OK`:**
```json
{
  "id": "...",
  "state": "complete",
  "difficulty": "hard",
  "scenario_title": "Budget Reallocation Showdown",
  "session_started_at": "2026-08-29T14:00:00Z",
  "session_ended_at": "2026-08-29T14:06:30Z",
  "participants": [
    { "display_name": "janedoe42", "role": "Finance Director", "seat": "A" },
    { "display_name": "alex_senior", "role": "Product Manager", "seat": "B" }
  ]
}
```

---

### 2.5 Leaderboard

#### `GET /leaderboard`

**Query Parameters:** `page` (default 1), `limit` (default 50, max 100)

**Response `200 OK`:**
```json
{
  "data": [
    {
      "rank": 1,
      "user_id": "...",
      "display_name": "TopPlayer",
      "avatar_url": "...",
      "elo_rating": 1875.20,
      "total_sessions": 142,
      "wins": 98
    }
  ],
  "my_rank": 42,
  "my_elo": 1342.50,
  "pagination": { "page": 1, "limit": 50, "total": 8420, "total_pages": 169 }
}
```

---

### 2.6 Scenarios — Admin

All admin endpoints require `role: admin` in JWT. Prefix: `/admin`

| Method | Path                    | Description                                   |
|--------|-------------------------|-----------------------------------------------|
| GET    | /admin/scenarios        | List scenarios (filters: difficulty, context) |
| POST   | /admin/scenarios        | Create scenario                               |
| PUT    | /admin/scenarios/:id    | Update scenario                               |
| DELETE | /admin/scenarios/:id    | Soft-delete (is_active = false)               |
| GET    | /admin/role-contexts    | List all role contexts                        |
| POST   | /admin/role-contexts    | Create role context                           |
| GET    | /admin/roles            | List roles (filter by context_id)             |
| POST   | /admin/roles            | Create role                                   |

---

## 3. WebSocket Protocol

### 3.1 Connection Handshake

```
Client → Server:
  GET wss://api.roleplay-sim.io/v1/ws?token=<access_token>
  Headers: Upgrade: websocket, Connection: Upgrade, Sec-WebSocket-Version: 13

Server → Client (success): HTTP 101 Switching Protocols
Server → Client (auth fail): HTTP 401 Unauthorized
```

On success, server immediately sends:

```json
{
  "event": "CONNECTED",
  "payload": {
    "connection_id": "conn_a1b2c3",
    "server_time": "2026-08-29T21:11:27Z"
  }
}
```

---

### 3.2 Envelope Format

Every WebSocket message (both directions):

```json
{
  "event":     "<EVENT_NAME>",
  "payload":   {},
  "timestamp": "<ISO-8601>",
  "seq":       1
}
```

| Field       | Description                                          |
|-------------|------------------------------------------------------|
| `event`     | SCREAMING_SNAKE_CASE identifier                      |
| `payload`   | Event-specific data                                  |
| `timestamp` | ISO-8601 UTC                                         |
| `seq`       | Monotonically increasing (per sender)                |

---

### 3.3 Client → Server Events

#### `JOIN_QUEUE`
```json
{ "event": "JOIN_QUEUE", "payload": {}, "timestamp": "...", "seq": 1 }
```

#### `LEAVE_QUEUE`
```json
{ "event": "LEAVE_QUEUE", "payload": {}, "timestamp": "...", "seq": 2 }
```

#### `SPIN_ACK`
```json
{
  "event": "SPIN_ACK",
  "payload": { "room_id": "room_x9k2m4" },
  "timestamp": "...", "seq": 3
}
```

#### `SCENARIO_ACK`
```json
{
  "event": "SCENARIO_ACK",
  "payload": { "room_id": "room_x9k2m4" },
  "timestamp": "...", "seq": 4
}
```

#### `PREP_READY`
```json
{
  "event": "PREP_READY",
  "payload": { "room_id": "room_x9k2m4" },
  "timestamp": "...", "seq": 5
}
```

#### `SESSION_END`
```json
{
  "event": "SESSION_END",
  "payload": {
    "room_id": "room_x9k2m4",
    "reason": "timer_expired"
  },
  "timestamp": "...", "seq": 6
}
```

`reason` values: `"timer_expired"` | `"user_ended"`

#### `PING`
```json
{
  "event": "PING",
  "payload": { "client_time": "2026-08-29T21:12:00Z" },
  "timestamp": "...", "seq": 7
}
```

---

### 3.4 Server → Client Events

#### `ROOM_READY`
```json
{
  "event": "ROOM_READY",
  "payload": {
    "room_id": "room_x9k2m4",
    "peer_display_name": "alex_senior",
    "peer_avatar_url": "https://cdn.roleplay-sim.io/avatars/alex.webp",
    "peer_elo_rating": 1410.00
  },
  "timestamp": "...", "seq": 1
}
```

#### `SPIN_START`
```json
{
  "event": "SPIN_START",
  "payload": {
    "room_id": "room_x9k2m4",
    "spin_seed": 4829173,
    "animation_duration_ms": 3500
  },
  "timestamp": "...", "seq": 2
}
```

#### `SPIN_RESULT`
```json
{
  "event": "SPIN_RESULT",
  "payload": {
    "room_id": "room_x9k2m4",
    "context": { "id": "...", "name": "IT Team", "slug": "it-team" },
    "your_role": {
      "id": "...", "name": "Junior Developer",
      "hierarchy_level": 1,
      "description": "A developer 6 months into their first professional role."
    },
    "peer_role": { "id": "...", "name": "Team Lead", "hierarchy_level": 4 },
    "difficulty": "medium"
  },
  "timestamp": "...", "seq": 3
}
```

#### `SCENARIO_ASSIGN`
```json
{
  "event": "SCENARIO_ASSIGN",
  "payload": {
    "room_id": "room_x9k2m4",
    "scenario_id": "...",
    "title": "Leave Request Under Deadline Pressure",
    "difficulty": "medium",
    "background_context": "The team is 3 weeks from a major product release...",
    "your_objective": "Secure 2 days of approved leave without damaging your relationship with the team lead.",
    "your_constraints": [
      "You cannot reveal a personal emergency — it must be framed as a general request.",
      "You must stay professional even if the lead becomes dismissive."
    ],
    "prep_duration_seconds": 180,
    "session_duration_seconds": 360
  },
  "timestamp": "...", "seq": 4
}
```

**Note:** `your_objective` and `your_constraints` are private — each peer receives only their own.

#### `PREP_TIMER_TICK`
```json
{
  "event": "PREP_TIMER_TICK",
  "payload": {
    "room_id": "room_x9k2m4",
    "seconds_remaining": 147,
    "peer_ready": false
  },
  "timestamp": "...", "seq": 5
}
```

#### `PREP_END`
```json
{
  "event": "PREP_END",
  "payload": {
    "room_id": "room_x9k2m4",
    "initiator_seat": "A"
  },
  "timestamp": "...", "seq": 6
}
```

`initiator_seat` = seat of the user who creates the WebRTC offer.

#### `SESSION_TIMER_TICK`
```json
{
  "event": "SESSION_TIMER_TICK",
  "payload": {
    "room_id": "room_x9k2m4",
    "seconds_remaining": 312,
    "phase": "live"
  },
  "timestamp": "...", "seq": 7
}
```

#### `EVALUATION_READY`
```json
{
  "event": "EVALUATION_READY",
  "payload": {
    "room_id": "room_x9k2m4",
    "session_id": "...",
    "your_score": {
      "overall_score": 78.50,
      "objective_achieved": true,
      "elo_delta": 12.30,
      "elo_new": 1342.50,
      "summary_feedback": "You demonstrated strong preparation...",
      "strengths": ["Clear framing of your request"],
      "areas_for_improvement": ["Could have offered a compromise earlier"],
      "rubric_scores": [
        {
          "dimension": "communication_clarity",
          "score": 8.2,
          "weight": 0.20,
          "justification": "..."
        }
      ]
    },
    "peer_score": {
      "overall_score": 82.10,
      "objective_achieved": true,
      "elo_delta": -12.30
    }
  },
  "timestamp": "...", "seq": 8
}
```

#### `PEER_DISCONNECTED`
```json
{
  "event": "PEER_DISCONNECTED",
  "payload": {
    "room_id": "room_x9k2m4",
    "reconnect_window_seconds": 30
  },
  "timestamp": "...", "seq": 9
}
```

#### `ROOM_CLOSED`
```json
{
  "event": "ROOM_CLOSED",
  "payload": {
    "room_id": "room_x9k2m4",
    "reason": "session_complete"
  },
  "timestamp": "...", "seq": 10
}
```

`reason` values: `"session_complete"` | `"peer_timeout"` | `"admin_close"`

#### `ERROR`
```json
{
  "event": "ERROR",
  "payload": {
    "code": "ROOM_NOT_FOUND",
    "message": "The specified room does not exist or has expired.",
    "retryable": false
  },
  "timestamp": "...", "seq": 11
}
```

#### `PONG`
```json
{
  "event": "PONG",
  "payload": {
    "server_time": "2026-08-29T21:12:00Z",
    "latency_hint_ms": 23
  },
  "timestamp": "...", "seq": 12
}
```

---

### 3.5 WebRTC Signaling Events

These relay SDP offers/answers and ICE candidates between peers via the server.

#### `SIGNAL` — Offer (Client → Server → Peer)
```json
{
  "event": "SIGNAL",
  "payload": {
    "room_id": "room_x9k2m4",
    "signal": { "type": "offer", "sdp": "v=0\r\no=- 4148...\r\n..." }
  },
  "timestamp": "...", "seq": 8
}
```

#### `SIGNAL` — Answer
```json
{
  "event": "SIGNAL",
  "payload": {
    "room_id": "room_x9k2m4",
    "signal": { "type": "answer", "sdp": "v=0\r\no=- 9271...\r\n..." }
  },
  "timestamp": "...", "seq": 9
}
```

#### `SIGNAL` — ICE Candidate
```json
{
  "event": "SIGNAL",
  "payload": {
    "room_id": "room_x9k2m4",
    "signal": {
      "type": "ice",
      "candidate": {
        "candidate": "candidate:1234 1 udp 2130706431 192.168.1.1 54321 typ host",
        "sdpMid": "0",
        "sdpMLineIndex": 0,
        "usernameFragment": "abc123"
      }
    }
  },
  "timestamp": "...", "seq": 10
}
```

**Server relay behavior:**
- Server identifies the other peer in the room and forwards the `SIGNAL` event verbatim.
- Server does NOT inspect or modify SDP content.
- If receiver is not connected, server stores in Redis cache (TTL: 60s).

---

## 4. Error Codes

### REST Error Response Format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable description.",
    "details": [
      { "field": "email", "issue": "Must be a valid email address." }
    ]
  }
}
```

### REST Error Code Reference

| HTTP Status | Code                  | Description                                       |
|-------------|-----------------------|---------------------------------------------------|
| 400         | `VALIDATION_ERROR`    | Request body/params failed validation             |
| 400         | `ALREADY_QUEUED`      | User is already in the matchmaking queue          |
| 401         | `UNAUTHORIZED`        | Missing or invalid JWT                            |
| 401         | `TOKEN_EXPIRED`       | Access token expired; use refresh endpoint        |
| 403         | `FORBIDDEN`           | Authenticated but insufficient permissions        |
| 404         | `USER_NOT_FOUND`      | User does not exist                               |
| 404         | `SESSION_NOT_FOUND`   | Session does not exist                            |
| 404         | `ROOM_NOT_FOUND`      | Room does not exist or has expired                |
| 409         | `CONFLICT`            | Resource conflict (duplicate username etc.)        |
| 422         | `UNPROCESSABLE`       | Semantically invalid request                      |
| 429         | `RATE_LIMITED`        | Too many requests                                 |
| 500         | `INTERNAL_ERROR`      | Unexpected server error                           |
| 503         | `SERVICE_UNAVAILABLE` | Downstream service unreachable                    |

### WebSocket Error Codes

| Code               | Description                                      | Retryable |
|--------------------|--------------------------------------------------|-----------|
| `ROOM_NOT_FOUND`   | Room ID invalid or expired                       | No        |
| `ROOM_FULL`        | Room already has 2 participants                  | No        |
| `INVALID_STATE`    | Event sent in wrong room state                   | No        |
| `SIGNAL_FAILED`    | Failed to relay WebRTC signal to peer            | Yes       |
| `EVAL_TIMEOUT`     | AI evaluation exceeded timeout (90s)             | Yes       |
| `PEER_TIMEOUT`     | Peer did not reconnect within the 30s window     | No        |
| `RATE_LIMITED`     | Too many events sent                             | Yes       |

---

## 5. Rate Limits

| Endpoint / Event            | Limit      | Window     |
|-----------------------------|------------|------------|
| `POST /auth/login`          | 10 req     | 1 minute   |
| `POST /auth/register`       | 5 req      | 1 minute   |
| `POST /matchmaking/enqueue` | 3 req      | 1 minute   |
| All other REST endpoints    | 60 req     | 1 minute   |
| WebSocket events (per user) | 30 events  | 1 second   |
| `SIGNAL` events (per user)  | 50 events  | 10 seconds |
| `PING` heartbeat            | 1 event    | 5 seconds  |

Rate limit headers on REST responses:

```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 47
X-RateLimit-Reset: 1722345660
```

When exceeded: `429 Too Many Requests` with `Retry-After` header.
