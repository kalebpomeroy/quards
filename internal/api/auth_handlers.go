package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"quards/internal/auth"
)

// LoginDiscordHandler initiates Discord OAuth flow
func LoginDiscordHandler(w http.ResponseWriter, r *http.Request) {
	discordConfig := auth.NewDiscordConfig()
	
	if !discordConfig.IsConfigured() {
		writeError(w, "Discord OAuth not configured", http.StatusServiceUnavailable)
		return
	}

	// Generate state parameter for CSRF protection
	state, err := generateState()
	if err != nil {
		writeError(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	// Store state in session/cookie for validation
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to Discord OAuth
	authURL := discordConfig.GetAuthURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// CallbackDiscordHandler handles Discord OAuth callback
func CallbackDiscordHandler(w http.ResponseWriter, r *http.Request) {
	discordConfig := auth.NewDiscordConfig()

	if !discordConfig.IsConfigured() {
		writeError(w, "Discord OAuth not configured", http.StatusServiceUnavailable)
		return
	}

	// Validate state parameter
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		writeError(w, "Missing state cookie", http.StatusBadRequest)
		return
	}

	if r.URL.Query().Get("state") != stateCookie.Value {
		writeError(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Handle OAuth error
	if errorCode := r.URL.Query().Get("error"); errorCode != "" {
		errorDesc := r.URL.Query().Get("error_description")
		writeError(w, fmt.Sprintf("OAuth error: %s - %s", errorCode, errorDesc), http.StatusBadRequest)
		return
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	accessToken, err := discordConfig.ExchangeCodeForToken(code)
	if err != nil {
		writeError(w, fmt.Sprintf("Failed to exchange code for token: %v", err), http.StatusInternalServerError)
		return
	}

	// Get user info from Discord
	discordUser, err := discordConfig.GetUserInfo(accessToken)
	if err != nil {
		writeError(w, fmt.Sprintf("Failed to get user info: %v", err), http.StatusInternalServerError)
		return
	}

	// Create or update user
	user, err := discordConfig.CreateOrUpdateUser(discordUser)
	if err != nil {
		writeError(w, fmt.Sprintf("Failed to create/update user: %v", err), http.StatusInternalServerError)
		return
	}

	// Create session
	userAgent := r.Header.Get("User-Agent")
	ipAddress := getClientIP(r)
	session, err := auth.CreateSession(user.ID, &ipAddress, &userAgent)
	if err != nil {
		writeError(w, fmt.Sprintf("Failed to create session: %v", err), http.StatusInternalServerError)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.SessionToken,
		Path:     "/",
		MaxAge:   int(auth.SessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to frontend with success
	http.Redirect(w, r, "/?auth=success", http.StatusTemporaryRedirect)
}

// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get session token from cookie
	sessionCookie, err := r.Cookie("session_token")
	if err == nil {
		// Delete session from database
		auth.DeleteSession(sessionCookie.Value)
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	writeResponse(w, map[string]string{"message": "Logged out successfully"})
}

// MeHandler returns current user information
func MeHandler(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r)
	if user == nil {
		writeError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	writeResponse(w, user)
}

// getClientIP extracts client IP address from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in case of multiple
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx != -1 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

// generateState generates a random state parameter for OAuth
func generateState() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}