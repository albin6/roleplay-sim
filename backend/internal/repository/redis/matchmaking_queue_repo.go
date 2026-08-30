package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const matchmakingQueueKey = "matchmaking:queue"

// MatchmakingQueueRepo manages the matchmaking Redis sorted set.
// Members are userIDs, scores are Elo ratings (for approximate skill matching).
type MatchmakingQueueRepo struct {
	client *redis.Client
}

func NewMatchmakingQueueRepo(client *redis.Client) *MatchmakingQueueRepo {
	return &MatchmakingQueueRepo{client: client}
}

func (r *MatchmakingQueueRepo) Enqueue(ctx context.Context, userID string, eloRating float64) error {
	err := r.client.ZAdd(ctx, matchmakingQueueKey, redis.Z{
		Score:  eloRating,
		Member: userID,
	}).Err()
	if err != nil {
		return fmt.Errorf("matchmaking_queue: enqueue: %w", err)
	}
	return nil
}

// Dequeue atomically pops the two users with the closest Elo ratings.
// Returns empty strings if fewer than 2 users are queued.
func (r *MatchmakingQueueRepo) Dequeue(ctx context.Context) (string, string, error) {
	// Use a Lua script for atomicity: pop 2 members
	script := redis.NewScript(`
		local count = redis.call('ZCARD', KEYS[1])
		if count < 2 then return {nil, nil} end
		local members = redis.call('ZPOPMIN', KEYS[1], 2)
		return {members[1], members[3]}
	`)
	result, err := script.Run(ctx, r.client, []string{matchmakingQueueKey}).StringSlice()
	if err != nil {
		if err == redis.Nil {
			return "", "", nil
		}
		return "", "", fmt.Errorf("matchmaking_queue: dequeue: %w", err)
	}
	if len(result) < 2 || result[0] == "" || result[1] == "" {
		return "", "", nil
	}
	return result[0], result[1], nil
}

func (r *MatchmakingQueueRepo) Remove(ctx context.Context, userID string) error {
	return r.client.ZRem(ctx, matchmakingQueueKey, userID).Err()
}

func (r *MatchmakingQueueRepo) IsQueued(ctx context.Context, userID string) (bool, error) {
	_, err := r.client.ZRank(ctx, matchmakingQueueKey, userID).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("matchmaking_queue: is_queued: %w", err)
	}
	return true, nil
}

func (r *MatchmakingQueueRepo) Size(ctx context.Context) (int64, error) {
	size, err := r.client.ZCard(ctx, matchmakingQueueKey).Result()
	if err != nil {
		return 0, fmt.Errorf("matchmaking_queue: size: %w", err)
	}
	return size, nil
}
