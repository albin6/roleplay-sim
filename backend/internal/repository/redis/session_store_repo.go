package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// SessionStoreRepo manages JWT session tokens in Redis.
// Key schema:
//   session:<sessionID>        → userID  (TTL = refresh token expiry)
//   user:sessions:<userID>     → SET of sessionIDs
type SessionStoreRepo struct {
	client *redis.Client
}

func NewSessionStoreRepo(client *redis.Client) *SessionStoreRepo {
	return &SessionStoreRepo{client: client}
}

func (r *SessionStoreRepo) Create(ctx context.Context, sessionID, userID string, ttl time.Duration) error {
	key := sessionKey(sessionID)
	if err := r.client.Set(ctx, key, userID, ttl).Err(); err != nil {
		return fmt.Errorf("session_store: create: %w", err)
	}
	return nil
}

func (r *SessionStoreRepo) Get(ctx context.Context, sessionID string) (string, error) {
	userID, err := r.client.Get(ctx, sessionKey(sessionID)).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("session_store: session not found")
		}
		return "", fmt.Errorf("session_store: get: %w", err)
	}
	return userID, nil
}

func (r *SessionStoreRepo) Delete(ctx context.Context, sessionID string) error {
	if err := r.client.Del(ctx, sessionKey(sessionID)).Err(); err != nil {
		return fmt.Errorf("session_store: delete: %w", err)
	}
	return nil
}

func (r *SessionStoreRepo) DeleteAllForUser(ctx context.Context, userID string) error {
	membersKey := userSessionsKey(userID)
	sessionIDs, err := r.client.SMembers(ctx, membersKey).Result()
	if err != nil {
		return fmt.Errorf("session_store: list user sessions: %w", err)
	}
	if len(sessionIDs) == 0 {
		return nil
	}
	keys := make([]string, len(sessionIDs))
	for i, sid := range sessionIDs {
		keys[i] = sessionKey(sid)
	}
	keys = append(keys, membersKey)
	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("session_store: delete all for user: %w", err)
	}
	return nil
}

func (r *SessionStoreRepo) AddUserSession(ctx context.Context, userID, sessionID string, ttl time.Duration) error {
	key := userSessionsKey(userID)
	pipe := r.client.Pipeline()
	pipe.SAdd(ctx, key, sessionID)
	pipe.Expire(ctx, key, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("session_store: add user session: %w", err)
	}
	return nil
}

func sessionKey(sessionID string) string {
	return "session:" + sessionID
}

func userSessionsKey(userID string) string {
	return "user:sessions:" + userID
}
