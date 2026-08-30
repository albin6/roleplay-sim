package entity

import (
	"time"

	"github.com/google/uuid"
)

// RoleContext groups roles into a workplace context (e.g., "IT Team", "Finance Team").
type RoleContext struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
}

// Role represents an individual role within a context (e.g., "Junior Developer").
type Role struct {
	ID             uuid.UUID
	ContextID      uuid.UUID
	Name           string
	Slug           string
	HierarchyLevel int // 1 (lowest) to 4 (highest)
	Description    *string
	IsActive       bool
	CreatedAt      time.Time
}

// RolePair is a matched pair of roles for a session.
type RolePair struct {
	Context  RoleContext
	RoleA    Role // lower hierarchy
	RoleB    Role // higher hierarchy
}
