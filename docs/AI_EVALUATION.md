# AI Pipeline & Prompt Engineering Spec

> **Project:** Real-Time Scenario-Based Roleplay Simulator
> **Version:** 1.0.0 | **Phase:** 1 — Documentation

---

## Table of Contents

1. [Pipeline Overview](#1-pipeline-overview)
2. [Dual-Channel Audio Capture & STT Workflow](#2-dual-channel-audio-capture--stt-workflow)
3. [Transcript Processing & Speaker Diarization](#3-transcript-processing--speaker-diarization)
4. [LLM Evaluation — Prompt Engineering](#4-llm-evaluation--prompt-engineering)
5. [Rubric Dimensions & Scoring Guide](#5-rubric-dimensions--scoring-guide)
6. [Elo Rating Calculation](#6-elo-rating-calculation)
7. [Pipeline Reliability & Error Handling](#7-pipeline-reliability--error-handling)
8. [Prompt Version Control](#8-prompt-version-control)

---

## 1. Pipeline Overview

The AI evaluation pipeline is triggered immediately when a session reaches the `EVALUATING` state. It runs asynchronously as a background worker and notifies the WebSocket hub via Redis Pub/Sub when complete.

```
Session Ends (state → EVALUATING)
          │
          ▼
  ┌─────────────────────────────────────────────────────────────────────────┐
  │                     AI EVALUATION WORKER (Go goroutine)                │
  │                                                                         │
  │  Step 1:  Retrieve dual-channel audio chunks from Redis Streams        │
  │  Step 2:  Merge mono streams → stereo WebM (ffmpeg, server-side)       │
  │  Step 3:  Send stereo WebM to Deepgram Nova-2 (single multichannel call)│
  │           Response: Channel A words + Channel B words (timestamped)    │
  │  Step 4:  Interleave words by timestamp → Interleaved Dialogue         │
  │  Step 5:  LLM Evaluation (parallel for User A and User B)              │
  │           gemini-1.5-flash (easy/medium) or gemini-1.5-pro (hard)     │
  │           Evaluate(context, scenario, transcript, role=A) → Score A    │
  │           Evaluate(context, scenario, transcript, role=B) → Score B    │
  │  Step 6:  Validate JSON outputs against response schema                │
  │  Step 7:  Persist evaluations to PostgreSQL                            │
  │  Step 8:  Calculate Elo delta for both users                           │
  │  Step 9:  Update PostgreSQL (users.elo_rating, leaderboard_history)    │
  │           Update Redis leaderboard sorted set                           │
  │  Step 10: Delete raw audio buffers from Redis                          │
  │  Step 11: Publish evaluation:complete to Redis Pub/Sub                 │
  └─────────────────────────────────────────────────────────────────────────┘
          │
          ▼
  WS Hub receives notification → broadcasts EVALUATION_READY to both peers
```

**Timeout Budgets:**

| Step                       | Timeout     |
|----------------------------|-------------|
| ffmpeg stereo merge        | 10 seconds  |
| Deepgram STT (multichannel)| 30 seconds  |
| LLM evaluation per user    | 45 seconds  |
| Total pipeline             | 90 seconds  |

---

## 2. Dual-Channel Audio Capture & STT Workflow

### 2.1 Audio Capture Strategy

Audio is captured **client-side** using the `MediaRecorder` API from each user's local microphone stream (independent of the WebRTC peer connection).

```
User A Browser
  │
  ├── WebRTC stream ──► P2P to User B
  │
  └── MediaRecorder (local mic only)
        │ Sends audio chunks every 5 seconds via WebSocket binary message
        ▼
      Go WS Server ──► Assembles chunks into per-user Redis Stream buffer
```

**Audio Format:**
- Codec: `audio/webm;codecs=opus`
- Sample Rate: 48,000 Hz
- Channels: 1 (mono — local microphone only)
- Chunk interval: 5,000ms
- Max session audio: ~6 min × ~97 kbps opus ≈ ~4.4 MB per user

### 2.2 Audio Buffer Storage (Redis Streams)

```
Key:    audio:stream:<room_id>:<user_id>
Type:   Redis Stream (XADD)
Fields:
  chunk_index → integer (0, 1, 2, ...)
  data        → base64-encoded audio bytes
  timestamp   → millisecond offset from session start

XADD audio:stream:<room_id>:<user_id> * chunk_index 0 data <base64> timestamp 0
```

After evaluation completes: `DEL audio:stream:<room_id>:<user_id>`

### 2.3 STT API Integration — Deepgram Nova-2 (Primary)

**Rationale:** Deepgram Nova-2 supports **multichannel audio** natively in a single API call — it returns independent per-channel transcripts with word-level timestamps. This avoids two separate API calls and makes the merge step trivial (timestamp ordering instead of guessing speaker from a single mix).

**Pre-processing — Stereo Assembly:**

The two mono WebM streams from Redis are merged into a single stereo WebM file server-side using ffmpeg before sending to Deepgram:

```bash
# Channel 0 = User A (left), Channel 1 = User B (right)
ffmpeg -i user_a.webm -i user_b.webm \
  -filter_complex "[0:a][1:a]amerge=inputs=2,pan=stereo|c0=c0|c1=c1[out]" \
  -map "[out]" -c:a libopus dual_channel.webm
```

**API Endpoint:** `POST https://api.deepgram.com/v1/listen`

**Request configuration:**
```
Authorization: Token <DEEPGRAM_API_KEY>
Content-Type: audio/webm

Query parameters:
  model=nova-2
  multichannel=true       ← returns per-channel transcripts
  channels=2
  punctuate=true
  utterances=true
  words=true              ← word-level timestamps for interleaving
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
          "transcript": "I understand the release pressure and I want to be transparent.",
          "confidence": 0.98,
          "words": [
            { "word": "I",           "start": 0.00, "end": 0.12, "confidence": 0.99 },
            { "word": "understand",  "start": 0.14, "end": 0.82, "confidence": 0.98 },
            { "word": "the",         "start": 0.84, "end": 0.94, "confidence": 0.99 },
            { "word": "release",     "start": 0.96, "end": 1.40, "confidence": 0.97 },
            { "word": "pressure",    "start": 1.42, "end": 1.90, "confidence": 0.98 }
          ]
        }]
      },
      {
        "alternatives": [{
          "transcript": "I appreciate that. What exactly do you need?",
          "confidence": 0.97,
          "words": [
            { "word": "I",          "start": 2.10, "end": 2.22, "confidence": 0.99 },
            { "word": "appreciate", "start": 2.24, "end": 2.80, "confidence": 0.97 }
          ]
        }]
      }
    ]
  }
}
```

**Key advantage of multichannel:** Channel 0 words and Channel 1 words carry absolute timestamps from session start. The interleaving merge simply sorts all words across both channels by `start` time — no ambiguity about who spoke when.

### 2.4 STT API Integration — OpenAI Whisper (Fallback)

Activated if Deepgram returns an error after 2 retries, or if the Deepgram circuit breaker is open.

**Endpoint:** `POST https://api.openai.com/v1/audio/transcriptions`

Whisper does not support multichannel — two separate sequential calls are made, one per mono channel, then interleaved.

```
model:                    whisper-1
response_format:          verbose_json
timestamp_granularities:  ["word", "segment"]
language:                 en
temperature:              0
```

**Whisper response structure (verbose_json):**
```json
{
  "task": "transcribe",
  "language": "english",
  "duration": 360.5,
  "text": "I understand the release pressure...",
  "segments": [
    {
      "id": 0,
      "start": 0.0,
      "end": 3.2,
      "text": "I understand the release pressure and I want to be transparent.",
      "words": [
        { "word": "I", "start": 0.0, "end": 0.2 },
        { "word": "understand", "start": 0.2, "end": 0.8 }
      ]
    }
  ]
}
```

---

## 3. Transcript Processing & Speaker Diarization

### 3.1 Interleaving Algorithm

Since each channel is transcribed independently (channels are already separated by user), the merge algorithm:

1. Parse segment arrays from both STT responses
2. Assign speaker labels: Channel A → `[ROLE_A]`, Channel B → `[ROLE_B]`
3. Merge both segment arrays into a single list sorted by `start` timestamp
4. Produce the interleaved transcript string

**Interleaved Transcript Format:**
```
[00:00] [Junior Developer] I understand the release pressure, and I want to be transparent about my situation.
[00:04] [Team Lead] I appreciate that. What exactly do you need?
[00:07] [Junior Developer] I need two days off next week — Wednesday and Thursday.
[00:11] [Team Lead] Two days is a significant ask right now given where we are in the sprint.
...
```

### 3.2 Quality Checks

| Check                             | Action on Failure                             |
|-----------------------------------|-----------------------------------------------|
| Transcript < 50 words total       | Flag as `low_quality`; reduce eval weight     |
| User's channel < 10 words         | Score that user 0; mark `evaluation_skipped`  |
| STT confidence < 0.6 avg segment  | Add `[LOW_CONFIDENCE]` warning to prompt      |
| Duration mismatch > 30s           | Log warning; use actual transcript length     |

---

## 4. LLM Evaluation — Prompt Engineering (Gemini 1.5 Pro / Flash)

The evaluation runs **twice** — once per participant — with each call receiving the full interleaved transcript but evaluating from only one role's perspective.

**Model routing:**

| Condition                                   | Model used           |
|---------------------------------------------|----------------------|
| Difficulty is `easy` or `medium`            | `gemini-1.5-flash`   |
| Difficulty is `hard`                        | `gemini-1.5-pro`     |
| Flash output fails JSON schema validation   | Retry with `gemini-1.5-pro` |
| Pro fails after 3 retries                   | Apply fallback evaluation |

**Go SDK:** `google.golang.org/genai` (official Google Gen AI Go SDK)

Structured JSON output is enforced at the API level via `ResponseMIMEType: "application/json"` and `ResponseSchema` — the model cannot emit non-conforming responses.

```go
config := &genai.GenerateContentConfig{
    ResponseMIMEType: "application/json",
    ResponseSchema:   rubricOutputSchema,  // *genai.Schema matching section 4.3
    Temperature:      genai.Ptr[float32](0.2),
    TopP:             genai.Ptr[float32](0.8),
    MaxOutputTokens:  2048,
}

model := "gemini-1.5-flash"
if difficulty == "hard" || flashFailed {
    model = "gemini-1.5-pro"
}

result, err := client.Models.GenerateContent(ctx, model, genai.Text(userPrompt), config)
```

### 4.1 System Prompt

```
You are an expert workplace communication coach and trained evaluator for professional roleplay simulations.

Your role is to evaluate a participant's performance in a timed workplace roleplay conversation based strictly on the provided rubric dimensions. You have deep expertise in negotiation theory, conflict resolution, organizational psychology, and professional communication.

EVALUATION PRINCIPLES:
1. Be objective and evidence-based. Every score MUST be grounded in specific quotes from the transcript.
2. Be calibrated. A score of 10/10 should be rare and reserved for exemplary, professional-grade performance.
3. Be constructive. Feedback should be actionable and specific, not generic.
4. Stay in scope. Only evaluate the evaluated participant's performance, not the peer's.
5. Follow the JSON schema exactly. Do not include any text outside the JSON object.

SCORING SCALE (per dimension, 0-10):
0-2:  Poor — fundamental failure in this area
3-4:  Below average — significant gaps, inconsistent execution
5-6:  Average — competent but unremarkable; common mistakes present
7-8:  Good — above average; minor issues only
9-10: Excellent — professional-grade, highly effective execution
```

---

### 4.2 User Prompt Template

```
SCENARIO CONTEXT:
Title: {{scenario_title}}
Difficulty: {{difficulty}} ({{difficulty_rationale}})
Background: {{background_context}}

EVALUATED PARTICIPANT:
Role: {{evaluated_role_name}} (Hierarchy Level {{hierarchy_level}}/4)
Private Objective: {{evaluated_role_objective}}
Constraints: {{evaluated_role_constraints | join(", ")}}

PEER PARTICIPANT:
Role: {{peer_role_name}} (Hierarchy Level {{peer_hierarchy_level}}/4)
Peer's Objective (for context): {{peer_role_objective}}

SESSION TRANSCRIPT ({{session_duration_seconds}}s session):
{{interleaved_transcript}}

RUBRIC DIMENSIONS TO EVALUATE:
{{#each rubric_dimensions}}
- {{dimension_key}} (weight: {{weight}}): {{description}}
{{/each}}

EVALUATION TASK:
Evaluate ONLY the [{{evaluated_role_name}}] participant.
Produce a JSON object conforming exactly to the schema below.
Base every score on specific evidence from the transcript.
```

---

### 4.3 JSON Output Schema

The LLM is instructed to return **only** a valid JSON object. JSON mode / structured output is enabled at the API call level.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "required": [
    "overall_score",
    "objective_achieved",
    "objective_achievement_reasoning",
    "summary_feedback",
    "strengths",
    "areas_for_improvement",
    "rubric_scores"
  ],
  "properties": {
    "overall_score": {
      "type": "number",
      "minimum": 0,
      "maximum": 100,
      "description": "Weighted aggregate: sum(score_i * weight_i) * 10"
    },
    "objective_achieved": {
      "type": "boolean"
    },
    "objective_achievement_reasoning": {
      "type": "string",
      "minLength": 50,
      "maxLength": 500
    },
    "summary_feedback": {
      "type": "string",
      "minLength": 100,
      "maxLength": 600,
      "description": "2-3 paragraph coaching narrative. Tone: professional, encouraging, direct."
    },
    "strengths": {
      "type": "array",
      "items": { "type": "string", "minLength": 20, "maxLength": 200 },
      "minItems": 1,
      "maxItems": 5
    },
    "areas_for_improvement": {
      "type": "array",
      "items": { "type": "string", "minLength": 20, "maxLength": 200 },
      "minItems": 1,
      "maxItems": 5
    },
    "rubric_scores": {
      "type": "array",
      "minItems": 6,
      "maxItems": 6,
      "items": {
        "type": "object",
        "required": ["dimension", "score", "weight", "justification", "evidence_quotes"],
        "properties": {
          "dimension": {
            "type": "string",
            "enum": [
              "communication_clarity",
              "active_listening",
              "negotiation_strategy",
              "emotional_regulation",
              "empathy",
              "objective_alignment"
            ]
          },
          "score": { "type": "number", "minimum": 0, "maximum": 10 },
          "weight": { "type": "number", "enum": [0.10, 0.15, 0.20] },
          "justification": {
            "type": "string",
            "minLength": 50,
            "maxLength": 400
          },
          "evidence_quotes": {
            "type": "array",
            "items": { "type": "string", "minLength": 10, "maxLength": 300 },
            "minItems": 1,
            "maxItems": 4
          }
        }
      }
    }
  }
}
```

**Example LLM Output:**

```json
{
  "overall_score": 78.50,
  "objective_achieved": true,
  "objective_achievement_reasoning": "The Junior Developer successfully negotiated a compromise of 1.5 days, which while not the full 2 days requested, secured more than the Team Lead's opening offer and preserved the working relationship.",
  "summary_feedback": "Your performance demonstrated solid preparation and a genuine commitment to transparency, which established goodwill early in the conversation...",
  "strengths": [
    "Clear and respectful framing of the request in the opening 30 seconds.",
    "Effective use of empathy when acknowledging the release pressure before stating needs."
  ],
  "areas_for_improvement": [
    "Consider introducing a concrete compromise earlier to shift from positional to interest-based negotiation.",
    "When the Team Lead raised the sprint velocity concern, a more direct acknowledgement before pivoting would have shown stronger active listening."
  ],
  "rubric_scores": [
    {
      "dimension": "communication_clarity",
      "score": 8.2,
      "weight": 0.20,
      "justification": "The participant communicated their request in clear, professional language with logical structure: context → request → rationale. Minor issue: informal filler words used three times.",
      "evidence_quotes": [
        "I understand the release pressure, and I want to be transparent about my situation.",
        "What I'm asking for is two consecutive days — Wednesday and Thursday specifically."
      ]
    }
  ]
}
```

---

## 5. Rubric Dimensions & Scoring Guide

| Dimension               | Weight | High Score Indicators                               | Low Score Indicators                       |
|-------------------------|--------|-----------------------------------------------------|--------------------------------------------|
| `communication_clarity` | 0.20   | Structured, concise, jargon-appropriate phrasing    | Rambling, vague, or confusing language     |
| `active_listening`      | 0.15   | Paraphrasing, follow-up Qs, referencing peer words  | Ignoring peer's points, talking over       |
| `negotiation_strategy`  | 0.20   | BATNA awareness, interest-based framing, anchoring  | Positional bargaining, no compromise       |
| `emotional_regulation`  | 0.15   | Calm under pressure, measured tone throughout       | Frustration, defensive or heated language  |
| `empathy`               | 0.10   | Validates peer's constraints, perspective-taking    | Dismissive of peer's concerns              |
| `objective_alignment`   | 0.20   | Measurable progress toward stated private objective | Failed to move outcome strategically       |

**Overall Score Formula:**

```
overall_score = sum(score_i * weight_i) * 10

Where:
  score_i  in [0, 10]
  weight_i in {0.10, 0.15, 0.20}
  sum(weight_i) = 1.00
  overall_score in [0, 100]
```

**Example Calculation:**
```
communication_clarity:  8.2 * 0.20 = 1.640
active_listening:        7.0 * 0.15 = 1.050
negotiation_strategy:    7.5 * 0.20 = 1.500
emotional_regulation:    8.5 * 0.15 = 1.275
empathy:                 7.0 * 0.10 = 0.700
objective_alignment:     8.0 * 0.20 = 1.600
                                      ──────
Sum:                                  7.765
Overall Score:                7.765 * 10 = 77.65
```

---

## 6. Elo Rating Calculation

### 6.1 Standard Elo Formula

```
Expected score for User A:
  E_A = 1 / (1 + 10^((R_B - R_A) / 400))

New rating:
  R_A' = R_A + K * (S_A - E_A)

Where:
  R_A, R_B = current Elo ratings
  S_A      = actual outcome (weighted, see below)
  E_A      = expected outcome
  K        = effective K-factor
```

### 6.2 Outcome Determination

| User A Objective | User B Objective | S_A | S_B |
|------------------|------------------|-----|-----|
| Achieved         | Not Achieved     | 1.0 | 0.0 |
| Not Achieved     | Achieved         | 0.0 | 1.0 |
| Both Achieved    | Both Achieved    | 0.5 | 0.5 |
| Neither Achieved | Neither Achieved | 0.5 | 0.5 |

### 6.3 K-Factor Adjustments

```
K_effective = K_base * K_difficulty * K_role_hierarchy * K_experience

K_base = 32 (standard)

K_difficulty:
  easy   → 0.75
  medium → 1.00
  hard   → 1.25

K_role_hierarchy (applied to lower-hierarchy player's K only):
  hierarchy_diff = |level_A - level_B|
  diff = 0 → 1.00
  diff = 1 → 1.05
  diff = 2 → 1.10
  diff = 3 → 1.15

K_experience:
  total_sessions < 10  → 1.50  (new player, high volatility)
  total_sessions 10-50 → 1.00  (normal)
  total_sessions > 50  → 0.80  (veteran, stable rating)
```

### 6.4 Score-Weighted Modifier

```
score_bonus = (overall_score / 100 - 0.5) * 0.2
Range: [-0.10, +0.10]

S_A_weighted = clamp(S_A + score_bonus_A, 0.0, 1.0)
```

A user who wins with score 90/100 gains more Elo than one who wins with 55/100.

### 6.5 Full Calculation Example

```
Setup:
  User A: R=1200, sessions=5,  hierarchy=1 (junior), score=78.50, achieved=true
  User B: R=1350, sessions=30, hierarchy=4 (lead),   score=82.10, achieved=true

Step 1 — Expected scores:
  E_A = 1 / (1 + 10^((1350-1200)/400)) = 1 / (1 + 2.371) = 0.297
  E_B = 1 - 0.297 = 0.703

Step 2 — Raw outcomes (both achieved → S = 0.5 each):
  S_A = 0.5, S_B = 0.5

Step 3 — Score-weighted modifier:
  score_bonus_A = (78.50/100 - 0.5) * 0.2 = 0.285 * 0.2 = 0.057
  S_A_weighted  = clamp(0.5 + 0.057) = 0.557

  score_bonus_B = (82.10/100 - 0.5) * 0.2 = 0.321 * 0.2 = 0.064
  S_B_weighted  = clamp(0.5 + 0.064) = 0.564

Step 4 — K-factors:
  K_A = 32 * 1.00 (medium) * 1.15 (hierarchy_diff=3, junior upside) * 1.50 (new player)
      = 55.2

  K_B = 32 * 1.00 (medium) * 1.00 (senior, no upside) * 1.00 (30 sessions)
      = 32

Step 5 — New ratings:
  R_A' = 1200 + 55.2 * (0.557 - 0.297) = 1200 + 55.2 * 0.260 = 1200 + 14.35 = 1214.35
  R_B' = 1350 + 32  * (0.564 - 0.703) = 1350 + 32 * (-0.139) = 1350 - 4.45  = 1345.55
```

### 6.6 Rating Floor & Ceiling

```
Minimum Elo:  800   (hard floor — cannot drop below)
Maximum Elo:  No ceiling
Starting Elo: 1200
```

---

## 7. Pipeline Reliability & Error Handling

### 7.1 Retry Policy

| Step         | Max Retries | Backoff          | Failure Action                              |
|--------------|-------------|------------------|---------------------------------------------|
| STT request  | 3           | Exponential      | Mark transcript unavailable                 |
| LLM request  | 3           | Exponential      | Assign fallback score (50/100)              |
| DB write     | 5           | Linear 1s        | Alert + manual review queue                 |

### 7.2 Fallback Evaluation

If LLM evaluation fails after all retries:

```json
{
  "overall_score": 50.00,
  "objective_achieved": null,
  "summary_feedback": "Automated evaluation was unavailable for this session. Your session has been flagged for manual review. No Elo change will be applied.",
  "is_fallback": true
}
```

When `is_fallback = true`, **no Elo change is applied** to either user.

### 7.3 Idempotency

- Worker checks for an existing `evaluations` row before processing.
- If `(session_id, participant_id)` already exists, worker skips without error.
- Prevents double-evaluation on worker crash-and-restart.

---

## 8. Prompt Version Control

All prompt templates are versioned for evaluation reproducibility.

Every `evaluations` row stores:
- `llm_model_used`: e.g., `"gemini-1.5-flash-001"` or `"gemini-1.5-pro-001"`
- `stt_provider`: `"deepgram-nova-2"` or `"openai-whisper-1"` (for audit trail)
- `prompt_version`: e.g., `"v1.2"`
- `raw_llm_response`: full JSON response for audit trail

**`llm_model_used` valid values:**

| Value                    | When used                                     |
|--------------------------|-----------------------------------------------|
| `gemini-1.5-flash-001`   | Easy / medium difficulty, primary path        |
| `gemini-1.5-pro-001`     | Hard difficulty, or Flash JSON validation fail|

**Version Changelog:**

| Version | Date       | Changes                                                  |
|---------|------------|----------------------------------------------------------|
| v1.0    | 2026-09-01 | Initial prompt; 6 rubric dimensions; Gemini 1.5 Flash/Pro|
| v1.1    | TBD        | Add `evidence_quotes` requirement to all dimensions      |
| v1.2    | TBD        | Tighten scoring calibration; add difficulty instructions |

When a prompt version is updated:
1. Old version is archived (not deleted).
2. New sessions use the new version.
3. Historical evaluations remain tagged with their original version.

