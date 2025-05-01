-- States table: stores all known states for a given seed
CREATE TABLE states (
    seed TEXT NOT NULL,
    game TEXT NOT NULL,
    state_signature TEXT NOT NULL,
    state_json JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (seed, state_signature)
);

CREATE INDEX IF NOT EXISTS idx_states_seed ON states(seed);
CREATE INDEX IF NOT EXISTS idx_states_signature ON states(state_signature);

-- Edges table: one row per (seed, parent, action); child is NULL until resolved
CREATE TABLE edges (
    id SERIAL PRIMARY KEY,
    seed TEXT NOT NULL,
    parent_signature TEXT NOT NULL,
    turn INT,

    name TEXT NOT NULL,
    params JSONB,

    child_signature TEXT,
    status TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'CLOSED', 'ERROR', 'SOLVED')),
    
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_edges_status ON edges(status);
CREATE INDEX IF NOT EXISTS idx_edges_child_signature ON edges(child_signature);
CREATE INDEX IF NOT EXISTS idx_edges_child_signature ON edges(parent_signature);