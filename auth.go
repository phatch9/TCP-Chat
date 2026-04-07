package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a registered user
type User struct {
	username     string
	passwordHash string
	createdAt    time.Time
	lastLogin    time.Time
}

// Session represents an active user session
type Session struct {
	token     string
	username  string
	createdAt time.Time
	lastUsed  time.Time
}

// AuthManager handles user registration, login, and session management
type AuthManager struct {
	users    map[string]*User    // username -> User
	sessions map[string]*Session // token -> Session
	mu       sync.RWMutex
}

// NewAuthManager creates a new authentication manager
func NewAuthManager() *AuthManager {
	return &AuthManager{
		users:    make(map[string]*User),
		sessions: make(map[string]*Session),
	}
}

// Register creates a new user account
func (am *AuthManager) Register(username, password string) error {
	// Validate inputs
	if len(username) < 2 || len(username) > 20 {
		return fmt.Errorf("username must be 2-20 characters")
	}

	if len(strings.TrimSpace(username)) == 0 {
		return fmt.Errorf("username cannot be empty or spaces only")
	}

	if strings.Contains(username, " ") {
		return fmt.Errorf("username cannot contain spaces")
	}

	if len(password) < 4 {
		return fmt.Errorf("password must be at least 4 characters")
	}

	if len(password) > 100 {
		return fmt.Errorf("password must be less than 100 characters")
	}

	// Check if user already exists
	am.mu.RLock()
	_, exists := am.users[strings.ToLower(username)]
	am.mu.RUnlock()

	if exists {
		return fmt.Errorf("username '%s' already taken", username)
	}

	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %v", err)
	}

	// Create user
	user := &User{
		username:     strings.ToLower(username),
		passwordHash: string(hashedPassword),
		createdAt:    time.Now(),
	}

	// Store user
	am.mu.Lock()
	am.users[user.username] = user
	am.mu.Unlock()

	log.Printf("new user registered: %s", username)
	return nil
}

// Login authenticates a user and creates a session
func (am *AuthManager) Login(username, password string) (string, error) {
	username = strings.ToLower(username)

	// Get user
	am.mu.RLock()
	user, exists := am.users[username]
	am.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("invalid username or password")
	}

	// Check password
	err := bcrypt.CompareHashAndPassword([]byte(user.passwordHash), []byte(password))
	if err != nil {
		return "", fmt.Errorf("invalid username or password")
	}

	// Generate session token
	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("error creating session: %v", err)
	}

	// Create session
	session := &Session{
		token:     token,
		username:  username,
		createdAt: time.Now(),
		lastUsed:  time.Now(),
	}

	// Store session
	am.mu.Lock()
	am.sessions[token] = session
	user.lastLogin = time.Now()
	am.mu.Unlock()

	log.Printf("user logged in: %s", username)
	return token, nil
}

// Logout invalidates a session
func (am *AuthManager) Logout(token string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	session, exists := am.sessions[token]
	if !exists {
		return fmt.Errorf("invalid session")
	}

	delete(am.sessions, token)
	log.Printf("user logged out: %s", session.username)
	return nil
}

// ValidateSession checks if a token is valid
func (am *AuthManager) ValidateSession(token string) (string, error) {
	am.mu.RLock()
	session, exists := am.sessions[token]
	am.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("invalid or expired session")
	}

	// Update last used time
	am.mu.Lock()
	session.lastUsed = time.Now()
	am.mu.Unlock()

	return session.username, nil
}

// GetUser retrieves user information
func (am *AuthManager) GetUser(username string) (*User, error) {
	username = strings.ToLower(username)

	am.mu.RLock()
	user, exists := am.users[username]
	am.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// UserExists checks if a username is registered
func (am *AuthManager) UserExists(username string) bool {
	username = strings.ToLower(username)

	am.mu.RLock()
	_, exists := am.users[username]
	am.mu.RUnlock()

	return exists
}

// GetAllSessions returns all active sessions (for admin purposes)
func (am *AuthManager) GetAllSessions() []string {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var tokens []string
	for token := range am.sessions {
		tokens = append(tokens, token)
	}
	return tokens
}

// GetSessionCount returns the number of active sessions
func (am *AuthManager) GetSessionCount() int {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return len(am.sessions)
}

// GetUserCount returns the number of registered users
func (am *AuthManager) GetUserCount() int {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return len(am.users)
}

// generateToken creates a random session token
func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CleanupExpiredSessions removes sessions older than maxAge
func (am *AuthManager) CleanupExpiredSessions(maxAge time.Duration) int {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()
	removed := 0

	for token, session := range am.sessions {
		if now.Sub(session.lastUsed) > maxAge {
			delete(am.sessions, token)
			removed++
		}
	}

	if removed > 0 {
		log.Printf("cleaned up %d expired sessions", removed)
	}

	return removed
}
