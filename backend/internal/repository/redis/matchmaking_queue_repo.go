package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	matchmakingQueueEasy   = "matchmaking:queue:easy"
	matchmakingQueueMedium = "matchmaking:queue:medium"
	matchmakingQueueHard   = "matchmaking:queue:hard"
)

var allQueueKeys = []string{matchmakingQueueEasy, matchmakingQueueMedium, matchmakingQueueHard}

func getQueueKey(difficulty string) string {
	switch difficulty {
	case "easy":
		return matchmakingQueueEasy
	case "hard":
		return matchmakingQueueHard
	default:
		return matchmakingQueueMedium
	}
}

// MatchmakingQueueRepo manages the matchmaking Redis sorted sets per difficulty.
// Members are userIDs, scores are Elo ratings.
type MatchmakingQueueRepo struct {
	client *redis.Client
}

func NewMatchmakingQueueRepo(client *redis.Client) *MatchmakingQueueRepo {
	return &MatchmakingQueueRepo{client: client}
}

func (r *MatchmakingQueueRepo) Enqueue(ctx context.Context, userID string, eloRating float64, difficulty string) error {
	// First remove user from any existing queue so they cannot sit in multiple queues
	for _, key := range allQueueKeys {
		_ = r.client.ZRem(ctx, key, userID).Err()
	}

	key := getQueueKey(difficulty)
	err := r.client.ZAdd(ctx, key, redis.Z{
		Score:  eloRating,
		Member: userID,
	}).Err()
	if err != nil {
		return fmt.Errorf("matchmaking_queue: enqueue: %w", err)
	}
	return nil
}

// Dequeue atomically pops the two users with the closest Elo ratings for the given difficulty.
// Returns empty strings if fewer than 2 users are queued.
func (r *MatchmakingQueueRepo) Dequeue(ctx context.Context, difficulty string) (string, string, error) {
	key := getQueueKey(difficulty)

	// Use a Lua script for atomicity: pop 2 members from the specific difficulty queue
	script := redis.NewScript(`
		local count = redis.call('ZCARD', KEYS[1])
		if count < 2 then return {nil, nil} end
		local members = redis.call('ZPOPMIN', KEYS[1], 2)
		return {members[1], members[3]}
	`)
	result, err := script.Run(ctx, r.client, []string{key}).StringSlice()
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
	for _, key := range allQueueKeys {
		_ = r.client.ZRem(ctx, key, userID).Err()
	}
	return nil
}

func (r *MatchmakingQueueRepo) IsQueued(ctx context.Context, userID string) (bool, error) {
	for _, key := range allQueueKeys {
		_, err := r.client.ZRank(ctx, key, userID).Result()
		if err == nil {
			return true, nil
		}
		if err != redis.Nil {
			return false, fmt.Errorf("matchmaking_queue: is_queued: %w", err)
		}
	}
	return false, nil
}

func (r *MatchmakingQueueRepo) Size(ctx context.Context, difficulty string) (int64, error) {
	key := getQueueKey(difficulty)
	size, err := r.client.ZCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("matchmaking_queue: size: %w", err)
	}
	return size, nil
}
