package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserRole defines the role of a user in the platform.
type UserRole string

const (
	UserRolePlayer UserRole = "player"
	UserRoleAdmin  UserRole = "admin"
)

// User is the core user domain entity.
type User struct {
	ID            uuid.UUID `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	DisplayName   string    `json:"display_name"`
	AvatarURL     *string   `json:"avatar_url"`
	EloRating     float64   `json:"elo_rating"`
	TotalSessions int       `json:"total_sessions"`
	Wins          int       `json:"wins"`
	Losses        int       `json:"losses"`
	Role          UserRole  `json:"role"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// IsAdmin returns true if the user has admin role.
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// Rank returns the user's current leaderboard rank (populated externally from Redis).
type UserRank struct {
	User
	Rank int64
}
