package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"quards/internal/database"
)

// User represents a user in the system
type User struct {
	ID           int                    `json:"id"`
	Username     string                 `json:"username"`
	DisplayName  string                 `json:"displayName"`
	Email        *string                `json:"email,omitempty"`
	AvatarURL    *string                `json:"avatarUrl,omitempty"`
	Provider     string                 `json:"provider"`
	ProviderID   string                 `json:"providerId"`
	ProviderData map[string]interface{} `json:"providerData,omitempty"`
	IsActive     bool                   `json:"isActive"`
	CreatedAt    time.Time              `json:"createdAt"`
	ModifiedAt   time.Time              `json:"modifiedAt"`
	LastLoginAt  *time.Time             `json:"lastLoginAt,omitempty"`
}

// UserSession represents a user session
type UserSession struct {
	ID           int       `json:"id"`
	UserID       int       `json:"userId"`
	SessionToken string    `json:"sessionToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	CreatedAt    time.Time `json:"createdAt"`
	IPAddress    *string   `json:"ipAddress,omitempty"`
	UserAgent    *string   `json:"userAgent,omitempty"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Username     string                 `json:"username"`
	DisplayName  string                 `json:"displayName"`
	Email        *string                `json:"email,omitempty"`
	AvatarURL    *string                `json:"avatarUrl,omitempty"`
	Provider     string                 `json:"provider"`
	ProviderID   string                 `json:"providerId"`
	ProviderData map[string]interface{} `json:"providerData,omitempty"`
}

// CreateUser creates a new user in the database
func CreateUser(req *CreateUserRequest) (*User, error) {
	db := database.GetDB()

	// Convert provider data to JSON
	var providerDataJSON []byte
	var err error
	if req.ProviderData != nil {
		providerDataJSON, err = json.Marshal(req.ProviderData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal provider data: %w", err)
		}
	}

	var userID int
	err = db.QueryRow(`
		INSERT INTO users (username, display_name, email, avatar_url, provider, provider_id, provider_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		req.Username, req.DisplayName, req.Email, req.AvatarURL, req.Provider, req.ProviderID, providerDataJSON).Scan(&userID)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return LoadUserByID(userID)
}

// LoadUserByID loads a user by their ID
func LoadUserByID(userID int) (*User, error) {
	db := database.GetDB()

	var user User
	var providerDataJSON []byte
	err := db.QueryRow(`
		SELECT id, username, display_name, email, avatar_url, provider, provider_id, 
		       provider_data, is_active, created_at, modified_at, last_login_at
		FROM users WHERE id = $1`, userID).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Email, &user.AvatarURL,
		&user.Provider, &user.ProviderID, &providerDataJSON, &user.IsActive,
		&user.CreatedAt, &user.ModifiedAt, &user.LastLoginAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %d", userID)
		}
		return nil, fmt.Errorf("failed to load user: %w", err)
	}

	// Parse provider data JSON
	if providerDataJSON != nil {
		err = json.Unmarshal(providerDataJSON, &user.ProviderData)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal provider data: %w", err)
		}
	}

	return &user, nil
}

// LoadUserByProvider loads a user by their provider and provider ID
func LoadUserByProvider(provider, providerID string) (*User, error) {
	db := database.GetDB()

	var userID int
	err := db.QueryRow(`
		SELECT id FROM users WHERE provider = $1 AND provider_id = $2`,
		provider, providerID).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found with provider %s and ID %s", provider, providerID)
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return LoadUserByID(userID)
}

// UpdateLastLogin updates the user's last login timestamp
func UpdateLastLogin(userID int) error {
	db := database.GetDB()

	_, err := db.Exec(`
		UPDATE users SET last_login_at = NOW() WHERE id = $1`, userID)

	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}