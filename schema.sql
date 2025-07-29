-- Decks table: stores all deck information
CREATE TABLE decks (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    cards JSONB NOT NULL, -- CardID -> Count mapping
    created_at TIMESTAMP DEFAULT NOW(),
    modified_at TIMESTAMP DEFAULT NOW()
);

-- Games table: stores game metadata and logs
CREATE TABLE games (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE, -- Auto-generated, can be updated later as metadata
    player1_deck TEXT NOT NULL REFERENCES decks(name),
    player2_deck TEXT NOT NULL REFERENCES decks(name),
    seed INTEGER,
    log_content TEXT NOT NULL DEFAULT '', -- The actual game log
    status TEXT NOT NULL DEFAULT 'created' CHECK (status IN ('created', 'in_progress', 'completed')),
    winner INTEGER CHECK (winner IN (1, 2)),
    turns INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    modified_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for better performance
CREATE INDEX idx_games_player1_deck ON games(player1_deck);
CREATE INDEX idx_games_player2_deck ON games(player2_deck);
CREATE INDEX idx_games_status ON games(status);
CREATE INDEX idx_games_created_at ON games(created_at);

-- Function to update modified_at timestamp
CREATE OR REPLACE FUNCTION update_modified_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers to auto-update modified_at
CREATE TRIGGER update_decks_modified_at
    BEFORE UPDATE ON decks
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_at();

CREATE TRIGGER update_games_modified_at
    BEFORE UPDATE ON games
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_at();