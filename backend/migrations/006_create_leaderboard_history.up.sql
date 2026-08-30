CREATE TABLE leaderboard_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    elo_before DECIMAL(10,2) NOT NULL,
    elo_after DECIMAL(10,2) NOT NULL,
    elo_delta DECIMAL(10,2) NOT NULL,
    rank_before INTEGER,
    rank_after INTEGER,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_leaderboard_history_user_id ON leaderboard_history(user_id);
CREATE INDEX idx_leaderboard_history_recorded_at ON leaderboard_history(recorded_at);
