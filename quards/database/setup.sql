-- States table: stores all known states for a given seed
CREATE TABLE states (
    seed_id TEXT NOT NULL,
    state_signature TEXT NOT NULL,
    state_json JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (seed_id, state_signature)
);

CREATE INDEX IF NOT EXISTS idx_states_seed_id ON states(seed_id);
CREATE INDEX IF NOT EXISTS idx_states_signature ON states(state_signature);

-- Edges table: one row per (seed, parent, action); child is NULL until resolved
CREATE TABLE edges (
    parent_signature TEXT NOT NULL,
    id SERIAL PRIMARY KEY,


    name TEXT NOT NULL,
    params JSONB,

    child_signature TEXT,
    status TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'CLOSED', 'ERROR')),
    
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_edges_status ON edges(status);
CREATE INDEX IF NOT EXISTS idx_edges_child_signature ON edges(child_signature);
CREATE INDEX IF NOT EXISTS idx_edges_child_signature ON edges(parent_signature);