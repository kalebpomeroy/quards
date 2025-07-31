package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// UserContextKey is the key used to store user in request context
type UserContextKey string

const (
	UserKey UserContextKey = "user"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	devMode   bool
	devUserID int
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware() *AuthMiddleware {
	devMode := os.Getenv("ENVIRONMENT") == "development" || os.Getenv("AUTH_DEV_MODE") == "true"
	devUserID := 1 // Default dev user ID

	if devUserIDStr := os.Getenv("AUTH_DEV_USER_ID"); devUserIDStr != "" {
		if id, err := strconv.Atoi(devUserIDStr); err == nil {
			devUserID = id
		}
	}

	return &AuthMiddleware{
		devMode:   devMode,
		devUserID: devUserID,
	}
}

// RequireAuth is middleware that requires authentication
func (a *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := a.authenticateRequest(r)
		if err != nil {
			writeAuthError(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Add user to request context
		ctx := context.WithValue(r.Context(), UserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth is middleware that optionally authenticates the user
func (a *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := a.authenticateRequest(r) // Ignore error for optional auth

		// Add user to request context (may be nil)
		ctx := context.WithValue(r.Context(), UserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// authenticateRequest attempts to authenticate a request
func (a *AuthMiddleware) authenticateRequest(r *http.Request) (*User, error) {
	// Dev mode bypass
	if a.devMode {
		if user, err := LoadUserByID(a.devUserID); err == nil {
			return user, nil
		}
		// If dev user doesn't exist, fall through to normal auth
	}

	// Try session token from cookie
	if cookie, err := r.Cookie("session_token"); err == nil {
		if user, err := ValidateSession(cookie.Value); err == nil {
			return user, nil
		}
	}

	// Try Authorization header
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if user, err := ValidateSession(token); err == nil {
				return user, nil
			}
		}
	}

	return nil, fmt.Errorf("no valid authentication found")
}

// GetUserFromContext extracts the user from request context
func GetUserFromContext(r *http.Request) *User {
	if user, ok := r.Context().Value(UserKey).(*User); ok {
		return user
	}
	return nil
}

// GetUserIDFromContext extracts the user ID from request context
func GetUserIDFromContext(r *http.Request) int {
	if user := GetUserFromContext(r); user != nil {
		return user.ID
	}
	return 0
}

// writeAuthError writes an authentication error response
func writeAuthError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, message)))
}