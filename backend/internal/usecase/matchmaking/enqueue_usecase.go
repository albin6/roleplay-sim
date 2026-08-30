package matchmaking

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	domainerrors "github.com/roleplay-sim/backend/internal/domain/errors"
	"github.com/roleplay-sim/backend/internal/domain/repository"
)

type EnqueueInput struct {
	UserID     uuid.UUID
	EloRating  float64
	Difficulty string
	Context    string
}

type EnqueueResult struct {
	Position      int64
	QueueSize     int64
}

type EnqueueUseCase struct {
	queue    repository.MatchmakingQueueRepository
	userRepo repository.UserRepository
}

func NewEnqueueUseCase(queue repository.MatchmakingQueueRepository, userRepo repository.UserRepository) *EnqueueUseCase {
	return &EnqueueUseCase{queue: queue, userRepo: userRepo}
}

func (uc *EnqueueUseCase) Execute(ctx context.Context, in EnqueueInput) (*EnqueueResult, error) {
	userIDStr := in.UserID.String()

	// Check if already queued
	queued, err := uc.queue.IsQueued(ctx, userIDStr)
	if err != nil {
		return nil, fmt.Errorf("enqueue: check queued: %w", err)
	}
	if queued {
		return nil, domainerrors.ErrAlreadyQueued
	}

	if err := uc.queue.Enqueue(ctx, userIDStr, in.EloRating, in.Difficulty); err != nil {
		return nil, fmt.Errorf("enqueue: %w", err)
	}

	size, err := uc.queue.Size(ctx, in.Difficulty)
	if err != nil {
		return nil, fmt.Errorf("enqueue: get size: %w", err)
	}

	return &EnqueueResult{
		Position:  size,
		QueueSize: size,
	}, nil
}

type DequeueUseCase struct {
	queue repository.MatchmakingQueueRepository
}

func NewDequeueUseCase(queue repository.MatchmakingQueueRepository) *DequeueUseCase {
	return &DequeueUseCase{queue: queue}
}

func (uc *DequeueUseCase) Execute(ctx context.Context, userID uuid.UUID) error {
	queued, err := uc.queue.IsQueued(ctx, userID.String())
	if err != nil {
		return fmt.Errorf("dequeue: check: %w", err)
	}
	if !queued {
		return domainerrors.ErrNotQueued
	}
	return uc.queue.Remove(ctx, userID.String())
}
