package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// SupabaseAuth handles Supabase authentication
type SupabaseAuth struct {
	URL string
	Key string
}

// NewSupabaseAuth creates a new Supabase auth instance
func NewSupabaseAuth(url, key string) *SupabaseAuth {
	return &SupabaseAuth{
		URL: url,
		Key: key,
	}
}

// User represents an authenticated user
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// VerifyToken verifies a Supabase JWT token and returns the user
func (s *SupabaseAuth) VerifyToken(token string) (*User, error) {
	if s.URL == "" || s.Key == "" {
		return nil, fmt.Errorf("supabase not configured")
	}

	// Call Supabase API to verify token
	req, err := http.NewRequest("GET", s.URL+"/auth/v1/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("apikey", s.Key)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token verification failed: %s", string(body))
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	return &user, nil
}

// ExtractToken extracts the bearer token from Authorization header
func ExtractToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is empty")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}
