# Database & State Design

> **Project:** Real-Time Scenario-Based Roleplay Simulator
> **Version:** 1.0.0 | **Phase:** 1 — Documentation

---

## Table of Contents

1. [PostgreSQL Schema Overview](#1-postgresql-schema-overview)
2. [Table Definitions](#2-table-definitions)
3. [Entity Relationship Diagram](#3-entity-relationship-diagram)
4. [Indexes & Performance Considerations](#4-indexes--performance-considerations)
5. [Redis Data Models](#5-redis-data-models)
6. [Data Retention Policy](#6-data-retention-policy)

---

## 1. PostgreSQL Schema Overview

```
postgres://roleplay_db
│
├── users                    ← Core user identity + Elo rating
├── role_contexts            ← Workplace domains (IT, HR, Sales, Legal…)
├── roles                    ← Individual roles within a context
├── scenarios                ← Scenario templates with difficulty tiers
├── sessions                 ← Each completed roleplay session
├── session_participants     ← Joining table: user ↔ session with role assignment
├── evaluations              ← AI-generated evaluation per participant
├── rubric_scores            ← Granular rubric dimension scores per evaluation
└── leaderboard_history      ← Elo rating snapshots after each session
```

All tables use `UUID` primary keys generated via `gen_random_uuid()`.
All timestamps are `TIMESTAMPTZ` stored in UTC.

---

## 2. Table Definitions

### 2.1 `users`

Stores core user identity, authentication credentials, and current Elo rating.

```sql
CREATE TABLE users (
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    username            VARCHAR(50)     NOT NULL UNIQUE,
    email               VARCHAR(255)    NOT NULL UNIQUE,
    password_hash       VARCHAR(255)    NOT NULL,               -- bcrypt
    display_name        VARCHAR(100)    NOT NULL,
    avatar_url          TEXT,
    elo_rating          NUMERIC(8, 2)   NOT NULL DEFAULT 1200.00,
    total_sessions      INTEGER         NOT NULL DEFAULT 0,
    wins                INTEGER         NOT NULL DEFAULT 0,
    losses              INTEGER         NOT NULL DEFAULT 0,
    is_active           BOOLEAN         NOT NULL DEFAULT TRUE,
    email_verified      BOOLEAN         NOT NULL DEFAULT FALSE,
    last_seen_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email       ON users (email);
CREATE INDEX idx_users_elo_rating  ON users (elo_rating DESC);
CREATE INDEX idx_users_username    ON users (username);
```

**Notes:**
- `elo_rating` default of 1200 is the standard Elo starting point.
- `wins`/`losses` are denormalized counters for fast leaderboard queries.
- Soft-delete pattern via `is_active`; no hard deletes.

---

### 2.2 `role_contexts`

Defines the workplace domain (e.g., IT Team, HR Department, Sales Floor).

```sql
CREATE TABLE role_contexts (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100) NOT NULL UNIQUE,
    slug            VARCHAR(100) NOT NULL UNIQUE,
    description     TEXT,
    icon_url        TEXT,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
```

**Seed Data Examples:**

| name               | slug               |
|--------------------|--------------------|
| IT Team            | it-team            |
| HR Department      | hr-department      |
| Sales Floor        | sales-floor        |
| Legal & Compliance | legal-compliance   |
| Product Team       | product-team       |
| Finance            | finance            |

---

### 2.3 `roles`

Individual roles within a context. Each scenario uses exactly two roles (a pair).

```sql
CREATE TABLE roles (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    context_id      UUID         NOT NULL REFERENCES role_contexts(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    slug            VARCHAR(100) NOT NULL,
    hierarchy_level SMALLINT     NOT NULL,    -- 1=junior, 2=mid, 3=senior, 4=lead/manager
    description     TEXT,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (context_id, slug)
);

CREATE INDEX idx_roles_context_id ON roles (context_id);
```

**Hierarchy Level Reference:**

| Level | Meaning                       |
|-------|-------------------------------|
| 1     | Junior / Intern / Entry-level |
| 2     | Mid-level / Associate         |
| 3     | Senior / Specialist           |
| 4     | Lead / Manager / Director     |

`hierarchy_level` is used in the Elo calculation modifier — higher-level role opponents increase expected difficulty.

---

### 2.4 `scenarios`

Scenario templates defining the conflict premise, objectives, and difficulty tier.

```sql
CREATE TABLE scenarios (
    id                       UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    context_id               UUID          NOT NULL REFERENCES role_contexts(id) ON DELETE CASCADE,
    role_a_id                UUID          NOT NULL REFERENCES roles(id),
    role_b_id                UUID          NOT NULL REFERENCES roles(id),
    title                    VARCHAR(200)  NOT NULL,
    difficulty               VARCHAR(10)   NOT NULL CHECK (difficulty IN ('easy', 'medium', 'hard')),
    prep_duration_seconds    INTEGER       NOT NULL DEFAULT 180,
    session_duration_seconds INTEGER       NOT NULL DEFAULT 360,
    background_context       TEXT          NOT NULL,
    role_a_objective         TEXT          NOT NULL,
    role_b_objective         TEXT          NOT NULL,
    role_a_constraints       TEXT[],
    role_b_constraints       TEXT[],
    difficulty_rationale     TEXT,
    tags                     TEXT[],
    is_active                BOOLEAN       NOT NULL DEFAULT TRUE,
    play_count               INTEGER       NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_scenarios_context_difficulty ON scenarios (context_id, difficulty) WHERE is_active = TRUE;
CREATE INDEX idx_scenarios_difficulty         ON scenarios (difficulty) WHERE is_active = TRUE;
```

**Example Scenario Record:**

```
title:              "Leave Request Under Deadline Pressure"
difficulty:         "medium"
background_context: "The team is 3 weeks from a major product release. Sprint
                     velocity has dropped. Junior developer has a personal
                     commitment requiring 2 days away next week."
role_a_objective:   "Secure 2 days of approved leave without damaging your
                     relationship with the team lead."
role_b_objective:   "Limit absence to 1 day maximum while preserving team
                     morale and protecting the release timeline."
```

---

### 2.5 `sessions`

Each completed (or in-progress) roleplay session between two users.

```sql
CREATE TABLE sessions (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    scenario_id         UUID         NOT NULL REFERENCES scenarios(id),
    room_id             VARCHAR(100) NOT NULL UNIQUE,
    state               VARCHAR(20)  NOT NULL DEFAULT 'created'
                            CHECK (state IN (
                                'created', 'spinning', 'prep', 'signaling',
                                'live', 'evaluating', 'complete', 'abandoned'
                            )),
    difficulty          VARCHAR(10)  NOT NULL CHECK (difficulty IN ('easy', 'medium', 'hard')),
    prep_started_at     TIMESTAMPTZ,
    session_started_at  TIMESTAMPTZ,
    session_ended_at    TIMESTAMPTZ,
    eval_completed_at   TIMESTAMPTZ,
    transcript_a_url    TEXT,
    transcript_b_url    TEXT,
    abandonment_reason  VARCHAR(100),
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_state        ON sessions (state);
CREATE INDEX idx_sessions_scenario_id  ON sessions (scenario_id);
CREATE INDEX idx_sessions_room_id      ON sessions (room_id);
CREATE INDEX idx_sessions_created_at   ON sessions (created_at DESC);
```

---

### 2.6 `session_participants`

Join table linking each user to a session with their assigned role.

```sql
CREATE TABLE session_participants (
    id              UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      UUID           NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    user_id         UUID           NOT NULL REFERENCES users(id),
    role_id         UUID           NOT NULL REFERENCES roles(id),
    seat            CHAR(1)        NOT NULL CHECK (seat IN ('A', 'B')),
    elo_before      NUMERIC(8, 2)  NOT NULL,
    elo_after       NUMERIC(8, 2),
    elo_delta       NUMERIC(8, 2),
    joined_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    UNIQUE (session_id, user_id),
    UNIQUE (session_id, seat)
);

CREATE INDEX idx_session_participants_user_id    ON session_participants (user_id);
CREATE INDEX idx_session_participants_session_id ON session_participants (session_id);
```

---

### 2.7 `evaluations`

AI-generated evaluation result for each participant in a session.

```sql
CREATE TABLE evaluations (
    id                  UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id          UUID           NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    participant_id      UUID           NOT NULL REFERENCES session_participants(id),
    overall_score       NUMERIC(5, 2)  NOT NULL CHECK (overall_score BETWEEN 0 AND 100),
    objective_achieved  BOOLEAN        NOT NULL,
    llm_model_used      VARCHAR(100)   NOT NULL,
    prompt_version      VARCHAR(20)    NOT NULL,
    raw_transcript      TEXT           NOT NULL,
    summary_feedback    TEXT           NOT NULL,
    strengths           TEXT[],
    areas_for_improvement TEXT[],
    raw_llm_response    JSONB,
    stt_duration_ms     INTEGER,
    llm_duration_ms     INTEGER,
    created_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    UNIQUE (session_id, participant_id)
);

CREATE INDEX idx_evaluations_session_id     ON evaluations (session_id);
CREATE INDEX idx_evaluations_participant_id ON evaluations (participant_id);
```

---

### 2.8 `rubric_scores`

Granular dimension-level scores within an evaluation.

```sql
CREATE TABLE rubric_scores (
    id              UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    evaluation_id   UUID           NOT NULL REFERENCES evaluations(id) ON DELETE CASCADE,
    dimension       VARCHAR(50)    NOT NULL,
    score           NUMERIC(5, 2)  NOT NULL CHECK (score BETWEEN 0 AND 10),
    weight          NUMERIC(4, 3)  NOT NULL,
    justification   TEXT           NOT NULL,
    evidence_quotes TEXT[],
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rubric_scores_evaluation_id ON rubric_scores (evaluation_id);
```

**Rubric Dimensions:**

| Dimension               | Weight | Description                                                |
|-------------------------|--------|------------------------------------------------------------|
| `communication_clarity` | 0.20   | Clear, structured, jargon-appropriate communication        |
| `active_listening`      | 0.15   | Paraphrasing, follow-up questions, acknowledgement          |
| `negotiation_strategy`  | 0.20   | Tactical framing, BATNA awareness, mutual gains             |
| `emotional_regulation`  | 0.15   | Tone management, staying composed under pressure            |
| `empathy`               | 0.10   | Perspective-taking, validation of other party's concerns    |
| `objective_alignment`   | 0.20   | Degree to which the user achieved their stated goal         |

---

### 2.9 `leaderboard_history`

Immutable append-only log of every Elo rating change.

```sql
CREATE TABLE leaderboard_history (
    id              UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID           NOT NULL REFERENCES users(id),
    session_id      UUID           NOT NULL REFERENCES sessions(id),
    elo_before      NUMERIC(8, 2)  NOT NULL,
    elo_after       NUMERIC(8, 2)  NOT NULL,
    elo_delta       NUMERIC(8, 2)  NOT NULL,
    rank_before     INTEGER,
    rank_after      INTEGER,
    difficulty      VARCHAR(10)    NOT NULL,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_leaderboard_history_user_id    ON leaderboard_history (user_id, created_at DESC);
CREATE INDEX idx_leaderboard_history_session_id ON leaderboard_history (session_id);
```

---

## 3. Entity Relationship Diagram

```
users
  │
  ├── session_participants ──► sessions ──► scenarios
  │         │                                │
  │         └──► evaluations             role_contexts
  │                   │                      │
  │                   └──► rubric_scores    roles ◄──────────────────┐
  │                                          │                        │
  └── leaderboard_history                    └──────────────────── scenarios
                                               (role_a_id, role_b_id)
```

**Cardinalities:**

| Relationship                        | Cardinality |
|-------------------------------------|-------------|
| user → session_participants         | 1 : N       |
| session → session_participants      | 1 : 2       |
| session_participants → evaluations  | 1 : 1       |
| evaluations → rubric_scores         | 1 : 6       |
| user → leaderboard_history          | 1 : N       |
| scenario → sessions                 | 1 : N       |
| role_context → roles                | 1 : N       |
| role_context → scenarios            | 1 : N       |

---

## 4. Indexes & Performance Considerations

### 4.1 Critical Query Patterns

| Query                                    | Primary Index                                   |
|------------------------------------------|-------------------------------------------------|
| Fetch user by email (login)              | `idx_users_email`                               |
| Leaderboard top 100                      | `idx_users_elo_rating` (DESC)                   |
| User session history                     | `idx_session_participants_user_id`              |
| Random scenario by difficulty            | `idx_scenarios_difficulty` + `TABLESAMPLE`      |
| Evaluate session post-completion         | `idx_evaluations_session_id`                    |
| Elo history for a user                   | `idx_leaderboard_history_user_id`               |

### 4.2 Partitioning Strategy (Future)

- `leaderboard_history` → partition by `RANGE (created_at)` monthly
- `sessions` → partition by `RANGE (created_at)` monthly at scale

### 4.3 Connection Pooling

- Use **PgBouncer** in `transaction` mode
- Pool size: `(num_cores × 2) + 1` per backend instance
- Max connections: 100 (PostgreSQL `max_connections`)

---

## 5. Redis Data Models

### 5.1 Matchmaking Queue

```
Type:    Sorted Set
Key:     matchmaking:queue
Score:   Unix timestamp (enqueue time, for FIFO)
Member:  <user_id>
TTL:     None (managed by matchmaking worker)
```

### 5.2 User Matchmaking Metadata

```
Type:    Hash
Key:     matchmaking:user:<user_id>
Fields:
  elo_rating      → "1250.50"
  preferred_diff  → "medium"
  enqueued_at     → "1722345600"
  context_pref    → "it-team"
TTL:     300 seconds
```

### 5.3 Active Room State

```
Type:    Hash
Key:     room:<room_id>
Fields:
  state              → "LIVE"
  user_a_id          → "<uuid>"
  user_b_id          → "<uuid>"
  user_a_ws_node     → "node-1"
  user_b_ws_node     → "node-2"
  scenario_id        → "<uuid>"
  role_a_id          → "<uuid>"
  role_b_id          → "<uuid>"
  difficulty         → "medium"
  spin_seed          → "4829173"
  prep_expires_at    → "1722345780"
  session_expires_at → "1722346140"
  session_id         → "<uuid>"
TTL:     3600 seconds (safety net)
```

### 5.4 WebRTC Signaling Cache

```
Type:    String (JSON)
Key:     webrtc:<room_id>:offer:<user_id>    TTL: 60s
Key:     webrtc:<room_id>:answer:<user_id>   TTL: 60s
Key:     webrtc:<room_id>:ice:<user_id>      TTL: 60s
```

### 5.5 Session Timer

```
Type:    String (integer)
Key:     session:timer:<room_id>
Value:   Unix timestamp of session end
TTL:     session_duration_seconds + 60
```

### 5.6 Rate Limiting

```
Type:    String (counter)
Key:     ratelimit:ws:<user_id>     TTL: 1 second
Key:     ratelimit:rest:<ip>        TTL: 60 seconds
```

### 5.7 Global Leaderboard Cache

```
Type:    Sorted Set
Key:     leaderboard:global
Score:   elo_rating (float64)
Member:  <user_id>
TTL:     None (persistent)

ZADD leaderboard:global <elo> <user_id>
ZREVRANGE leaderboard:global 0 99 WITHSCORES
ZRANK leaderboard:global <user_id>
```

### 5.8 Pub/Sub Channels

```
Channel: room:<room_id>:events
  Purpose: Cross-node WebSocket event broadcasting

Channel: matchmaking:new_room
  Purpose: Notify WS hub instances of newly created rooms
  Message: { "room_id": "...", "user_a": "...", "user_b": "..." }

Channel: evaluation:complete
  Purpose: Notify WS hub that AI evaluation finished
  Message: { "room_id": "...", "session_id": "..." }
```

---

## 6. Data Retention Policy

| Data Type              | Retention                         | Reasoning                                   |
|------------------------|-----------------------------------|---------------------------------------------|
| Raw audio recordings   | Deleted immediately after STT     | Privacy; only transcripts retained          |
| Transcripts            | 90 days                           | Dispute resolution; evaluation audit        |
| Evaluations & scores   | Indefinite                        | Core product value; user history            |
| Leaderboard history    | Indefinite                        | Analytics; user progression tracking        |
| Session metadata       | Indefinite                        | Product analytics                           |
| Redis room state       | Deleted on CLOSED                 | Transient; no persistence needed            |
| Redis matchmaking data | TTL 300s                          | Auto-expired if unmatched                   |
| WebRTC cache           | TTL 60s                           | Ephemeral signaling data                    |
