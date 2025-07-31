-- Migration: 002_add_users_and_auth.sql
-- Description: Add users table and update existing tables with user associations
-- Created: 2025-07-31

-- Users table: stores user information and authentication data
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    email TEXT,
    avatar_url TEXT,
    provider TEXT NOT NULL DEFAULT 'discord', -- 'discord', 'github', etc.
    provider_id TEXT NOT NULL, -- External provider user ID
    provider_data JSONB, -- Additional provider-specific data
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    modified_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE
);

-- Unique constraint on provider + provider_id combination
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_provider_id ON users(provider, provider_id);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- User sessions table for session management
CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ip_address TEXT,
    user_agent TEXT
);

CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);

-- Add user_id column to decks table
ALTER TABLE decks ADD COLUMN IF NOT EXISTS user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_decks_user_id ON decks(user_id);

-- Add user_id column to games table  
ALTER TABLE games ADD COLUMN IF NOT EXISTS user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_games_user_id ON games(user_id);

-- Create a default user for development/testing (user_id = 1)
INSERT INTO users (id, username, display_name, email, provider, provider_id, provider_data) 
VALUES (
    1, 
    'dev_user', 
    'Development User', 
    'dev@localhost', 
    'dev', 
    'dev_1', 
    '{"dev_mode": true}'::jsonb
) ON CONFLICT (id) DO NOTHING;

-- Reset the sequence to ensure next user gets ID > 1
SELECT setval('users_id_seq', GREATEST(1, (SELECT MAX(id) FROM users)), true);

-- Update existing decks and games to belong to dev user if they don't have an owner
UPDATE decks SET user_id = 1 WHERE user_id IS NULL;
UPDATE games SET user_id = 1 WHERE user_id IS NULL;

-- Add triggers for users table
DROP TRIGGER IF EXISTS update_users_modified_at ON users;
CREATE TRIGGER update_users_modified_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_at();

-- Record this migration
INSERT INTO schema_migrations (version) VALUES ('002_add_users_and_auth') 
ON CONFLICT (version) DO NOTHING;