CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id VARCHAR(50) UNIQUE NOT NULL,
    scenario_id UUID NOT NULL REFERENCES scenarios(id),
    difficulty VARCHAR(10) NOT NULL CHECK (difficulty IN ('easy','medium','hard')),
    state VARCHAR(20) NOT NULL DEFAULT 'idle',
    spin_seed BIGINT,
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE session_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    role_id UUID NOT NULL REFERENCES roles(id),
    seat VARCHAR(1) NOT NULL CHECK (seat IN ('A','B')),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(session_id, user_id),
    UNIQUE(session_id, seat)
);

CREATE INDEX idx_sessions_room_id ON sessions(room_id);
CREATE INDEX idx_sessions_state ON sessions(state);
CREATE INDEX idx_session_participants_session_id ON session_participants(session_id);
CREATE INDEX idx_session_participants_user_id ON session_participants(user_id);
