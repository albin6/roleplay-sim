CREATE TABLE scenarios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    context_id UUID NOT NULL REFERENCES role_contexts(id),
    title VARCHAR(255) NOT NULL,
    difficulty VARCHAR(10) NOT NULL CHECK (difficulty IN ('easy','medium','hard')),
    background_context TEXT NOT NULL,
    role_a_objective TEXT NOT NULL,
    role_a_constraints JSONB NOT NULL DEFAULT '[]',
    role_b_objective TEXT NOT NULL,
    role_b_constraints JSONB NOT NULL DEFAULT '[]',
    prep_duration_seconds INTEGER NOT NULL DEFAULT 180,
    session_duration_seconds INTEGER NOT NULL DEFAULT 360,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_scenarios_context_id_difficulty ON scenarios(context_id, difficulty);
