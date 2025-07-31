package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"quards/internal/database"
)

// SessionDuration is how long sessions last
const SessionDuration = 30 * 24 * time.Hour // 30 days

// CreateSession creates a new user session
func CreateSession(userID int, ipAddress, userAgent *string) (*UserSession, error) {
	db := database.GetDB()

	// Generate random session token
	token, err := generateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	expiresAt := time.Now().Add(SessionDuration)

	var sessionID int
	err = db.QueryRow(`
		INSERT INTO user_sessions (user_id, session_token, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		userID, token, expiresAt, ipAddress, userAgent).Scan(&sessionID)

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &UserSession{
		ID:           sessionID,
		UserID:       userID,
		SessionToken: token,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}, nil
}

// ValidateSession validates a session token and returns the associated user
func ValidateSession(sessionToken string) (*User, error) {
	db := database.GetDB()

	var userID int
	var expiresAt time.Time

	err := db.QueryRow(`
		SELECT user_id, expires_at FROM user_sessions 
		WHERE session_token = $1 AND expires_at > NOW()`,
		sessionToken).Scan(&userID, &expiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid or expired session")
		}
		return nil, fmt.Errorf("failed to validate session: %w", err)
	}

	return LoadUserByID(userID)
}

// DeleteSession deletes a session (logout)
func DeleteSession(sessionToken string) error {
	db := database.GetDB()

	_, err := db.Exec(`DELETE FROM user_sessions WHERE session_token = $1`, sessionToken)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions from the database
func CleanupExpiredSessions() error {
	db := database.GetDB()

	_, err := db.Exec(`DELETE FROM user_sessions WHERE expires_at < NOW()`)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	return nil
}

// generateSessionToken generates a cryptographically secure random session token
func generateSessionToken() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 256 bits
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}