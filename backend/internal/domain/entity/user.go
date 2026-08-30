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
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	DisplayName  string
	AvatarURL    *string
	EloRating    float64
	TotalSessions int
	Wins         int
	Losses       int
	Role         UserRole
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
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
