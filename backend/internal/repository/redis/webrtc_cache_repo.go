package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// WebRTCCacheRepo caches SDP/ICE signals for late-joining peers.
// Key: webrtc:<roomID>:<forUserID>  → Redis List of signal JSON bytes
type WebRTCCacheRepo struct {
	client *redis.Client
}

func NewWebRTCCacheRepo(client *redis.Client) *WebRTCCacheRepo {
	return &WebRTCCacheRepo{client: client}
}

func (r *WebRTCCacheRepo) StoreSignal(ctx context.Context, roomID, fromUserID string, signal []byte, ttl time.Duration) error {
	key := webrtcKey(roomID, fromUserID)
	pipe := r.client.Pipeline()
	pipe.RPush(ctx, key, signal)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("webrtc_cache: store: %w", err)
	}
	return nil
}

func (r *WebRTCCacheRepo) GetSignals(ctx context.Context, roomID, forUserID string) ([][]byte, error) {
	key := webrtcKey(roomID, forUserID)
	result, err := r.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("webrtc_cache: get: %w", err)
	}
	signals := make([][]byte, len(result))
	for i, s := range result {
		signals[i] = []byte(s)
	}
	return signals, nil
}

func (r *WebRTCCacheRepo) ClearSignals(ctx context.Context, roomID, forUserID string) error {
	return r.client.Del(ctx, webrtcKey(roomID, forUserID)).Err()
}

func webrtcKey(roomID, userID string) string {
	return fmt.Sprintf("webrtc:%s:%s", roomID, userID)
}
