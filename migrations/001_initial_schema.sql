-- Migration: 001_initial_schema.sql
-- Description: Create initial database schema for quards application
-- Created: 2025-07-31

-- Decks table: stores all deck information
CREATE TABLE IF NOT EXISTS decks (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT DEFAULT '',
    cards JSONB NOT NULL, -- CardID -> Count mapping  
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    modified_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Games table: stores game metadata and logs
CREATE TABLE IF NOT EXISTS games (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE, -- Auto-generated, can be updated later as metadata
    player1_deck TEXT NOT NULL, -- Deck names (stored as names for log compatibility)
    player2_deck TEXT NOT NULL, -- Deck names (stored as names for log compatibility)
    seed INTEGER,
    log_content TEXT NOT NULL DEFAULT '', -- The actual game log
    status TEXT NOT NULL DEFAULT 'created' CHECK (status IN ('created', 'in_progress', 'completed')),
    winner INTEGER CHECK (winner IN (1, 2)),
    turns INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    modified_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_decks_name ON decks(name);
CREATE INDEX IF NOT EXISTS idx_decks_created_at ON decks(created_at);

CREATE INDEX IF NOT EXISTS idx_games_player1_deck ON games(player1_deck);
CREATE INDEX IF NOT EXISTS idx_games_player2_deck ON games(player2_deck);
CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);
CREATE INDEX IF NOT EXISTS idx_games_created_at ON games(created_at);
CREATE INDEX IF NOT EXISTS idx_games_seed ON games(seed);
CREATE INDEX IF NOT EXISTS idx_games_winner ON games(winner);

-- Function to update modified_at timestamp
CREATE OR REPLACE FUNCTION update_modified_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers to auto-update modified_at
DROP TRIGGER IF EXISTS update_decks_modified_at ON decks;
CREATE TRIGGER update_decks_modified_at
    BEFORE UPDATE ON decks
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_at();

DROP TRIGGER IF EXISTS update_games_modified_at ON games;
CREATE TRIGGER update_games_modified_at
    BEFORE UPDATE ON games
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_at();

-- Insert a migrations tracking table to keep track of applied migrations
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Record this migration
INSERT INTO schema_migrations (version) VALUES ('001_initial_schema') 
ON CONFLICT (version) DO NOTHING;