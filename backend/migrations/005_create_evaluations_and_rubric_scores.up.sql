CREATE TABLE evaluations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    participant_id UUID NOT NULL REFERENCES session_participants(id) ON DELETE CASCADE,
    overall_score DECIMAL(5,2),
    objective_achieved BOOLEAN,
    objective_achievement_reasoning TEXT,
    summary_feedback TEXT,
    strengths JSONB NOT NULL DEFAULT '[]',
    areas_for_improvement JSONB NOT NULL DEFAULT '[]',
    elo_delta DECIMAL(10,2),
    elo_rating_before DECIMAL(10,2),
    elo_rating_after DECIMAL(10,2),
    llm_model_used VARCHAR(100),
    stt_provider VARCHAR(50),
    prompt_version VARCHAR(20),
    raw_llm_response JSONB,
    is_fallback BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(session_id, participant_id)
);

CREATE TABLE rubric_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    evaluation_id UUID NOT NULL REFERENCES evaluations(id) ON DELETE CASCADE,
    dimension VARCHAR(50) NOT NULL CHECK (dimension IN ('communication_clarity','active_listening','negotiation_strategy','emotional_regulation','empathy','objective_alignment')),
    score DECIMAL(4,2) NOT NULL CHECK (score BETWEEN 0 AND 10),
    weight DECIMAL(4,2) NOT NULL,
    justification TEXT NOT NULL,
    evidence_quotes JSONB NOT NULL DEFAULT '[]',
    UNIQUE(evaluation_id, dimension)
);

CREATE INDEX idx_evaluations_session_id ON evaluations(session_id);
CREATE INDEX idx_evaluations_participant_id ON evaluations(participant_id);
