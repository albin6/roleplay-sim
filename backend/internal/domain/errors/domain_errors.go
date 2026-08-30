package errors

import "errors"

// Sentinel domain errors - compare with errors.Is()
var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrSessionNotFound     = errors.New("session not found")
	ErrRoomNotFound        = errors.New("room not found")
	ErrRoomFull            = errors.New("room is full")
	ErrScenarioNotFound    = errors.New("scenario not found")
	ErrEvaluationNotFound  = errors.New("evaluation not found")
	ErrAlreadyQueued       = errors.New("user is already in the matchmaking queue")
	ErrNotQueued           = errors.New("user is not in the matchmaking queue")
	ErrInvalidRoomState    = errors.New("invalid room state for this operation")
	ErrTokenExpired        = errors.New("token has expired")
	ErrTokenInvalid        = errors.New("token is invalid")
	ErrSessionRevoked      = errors.New("session has been revoked")
	ErrEvaluationExists    = errors.New("evaluation already exists for this participant")
)
