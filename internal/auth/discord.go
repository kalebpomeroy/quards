package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"quards/internal/database"
)

// DiscordConfig holds Discord OAuth configuration
type DiscordConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// DiscordUser represents user data from Discord API
type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	GlobalName    string `json:"global_name"`
	Avatar        string `json:"avatar"`
	Email         string `json:"email"`
	Verified      bool   `json:"verified"`
}

// NewDiscordConfig creates Discord OAuth configuration from environment variables
func NewDiscordConfig() *DiscordConfig {
	return &DiscordConfig{
		ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
		ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("DISCORD_REDIRECT_URL"),
	}
}

// IsConfigured returns true if Discord OAuth is properly configured
func (d *DiscordConfig) IsConfigured() bool {
	return d.ClientID != "" && d.ClientSecret != "" && d.RedirectURL != ""
}

// GetAuthURL returns the Discord OAuth authorization URL
func (d *DiscordConfig) GetAuthURL(state string) string {
	baseURL := "https://discord.com/api/oauth2/authorize"
	params := url.Values{
		"client_id":     {d.ClientID},
		"redirect_uri":  {d.RedirectURL},
		"response_type": {"code"},
		"scope":         {"identify email"},
		"state":         {state},
	}
	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// ExchangeCodeForToken exchanges authorization code for access token
func (d *DiscordConfig) ExchangeCodeForToken(code string) (string, error) {
	tokenURL := "https://discord.com/api/oauth2/token"
	
	data := url.Values{
		"client_id":     {d.ClientID},
		"client_secret": {d.ClientSecret},
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {d.RedirectURL},
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code for token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("discord token exchange failed: %s", string(body))
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	return tokenResponse.AccessToken, nil
}

// GetUserInfo fetches user information from Discord API
func (d *DiscordConfig) GetUserInfo(accessToken string) (*DiscordUser, error) {
	userURL := "https://discord.com/api/users/@me"
	
	req, err := http.NewRequest("GET", userURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discord user info request failed: %s", string(body))
	}

	var discordUser DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&discordUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &discordUser, nil
}

// CreateOrUpdateUser creates or updates a user from Discord data
func (d *DiscordConfig) CreateOrUpdateUser(discordUser *DiscordUser) (*User, error) {
	// Try to find existing user
	existingUser, err := LoadUserByProvider("discord", discordUser.ID)
	if err == nil {
		// User exists, update last login
		UpdateLastLogin(existingUser.ID)
		return existingUser, nil
	}

	// Create new user
	username := discordUser.Username
	displayName := discordUser.GlobalName
	if displayName == "" {
		displayName = fmt.Sprintf("%s#%s", discordUser.Username, discordUser.Discriminator)
	}

	var email *string
	if discordUser.Email != "" {
		email = &discordUser.Email
	}

	var avatarURL *string
	if discordUser.Avatar != "" {
		avatar := fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", discordUser.ID, discordUser.Avatar)
		avatarURL = &avatar
	}

	// Ensure username is unique
	username = d.ensureUniqueUsername(username)

	providerData := map[string]interface{}{
		"discord_username":      discordUser.Username,
		"discord_discriminator": discordUser.Discriminator,
		"discord_global_name":   discordUser.GlobalName,
		"verified":              discordUser.Verified,
	}

	createReq := &CreateUserRequest{
		Username:     username,
		DisplayName:  displayName,
		Email:        email,
		AvatarURL:    avatarURL,
		Provider:     "discord",
		ProviderID:   discordUser.ID,
		ProviderData: providerData,
	}

	user, err := CreateUser(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	UpdateLastLogin(user.ID)
	return user, nil
}

// ensureUniqueUsername ensures the username is unique by appending numbers if necessary
func (d *DiscordConfig) ensureUniqueUsername(baseUsername string) string {
	// Clean username (Discord usernames can have special characters)
	username := strings.ToLower(strings.ReplaceAll(baseUsername, " ", "_"))
	
	// Check if base username is available
	if _, err := LoadUserByID(0); err != nil { // This will always fail, just checking if function works
		// Try the base username first
		if !d.usernameExists(username) {
			return username
		}

		// Try with numbers
		for i := 1; i <= 999; i++ {
			candidate := fmt.Sprintf("%s_%d", username, i)
			if !d.usernameExists(candidate) {
				return candidate
			}
		}
	}

	// Fallback to username with timestamp if all else fails
	return fmt.Sprintf("%s_%d", username, int64(1000000))
}

// usernameExists checks if a username already exists
func (d *DiscordConfig) usernameExists(username string) bool {
	db := database.GetDB()
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", username).Scan(&count)
	return err == nil && count > 0
}