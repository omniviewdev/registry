package registry

import (
	"context"
	"time"
)

// LoginRequest is the request body for credential-based login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is the response from a successful login.
type LoginResponse struct {
	Token   string `json:"token"`
	UserID  uint   `json:"user_id"`
	Success bool   `json:"success"`
}

// User represents the authenticated user's profile.
type User struct {
	ID            uint       `json:"id"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	Avatar        string     `json:"avatar"`
	Role          string     `json:"role"`
	EmailVerified bool       `json:"email_verified"`
	CreatedAt     time.Time  `json:"created_at"`
	LastLogin     *time.Time `json:"last_login"`
}

// Login authenticates with email and password, returning a JWT token.
func (c *Client) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	var resp LoginResponse
	err := c.post(ctx, "/v1/auth/login", &LoginRequest{
		Email:    email,
		Password: password,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetMe returns the profile of the currently authenticated user.
func (c *Client) GetMe(ctx context.Context) (*User, error) {
	var user User
	if err := c.get(ctx, "/v1/auth/me", &user); err != nil {
		return nil, err
	}
	return &user, nil
}
