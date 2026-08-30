package redis

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type AudioBufferRepo struct {
	client *redis.Client
}

func NewAudioBufferRepo(client *redis.Client) *AudioBufferRepo {
	return &AudioBufferRepo{client: client}
}

func audioStreamKey(roomID, userID string) string {
	return fmt.Sprintf("audio:stream:%s:%s", roomID, userID)
}

func (r *AudioBufferRepo) AppendChunk(ctx context.Context, roomID, userID string, chunkIndex int, data []byte) error {
	key := audioStreamKey(roomID, userID)
	b64Data := base64.StdEncoding.EncodeToString(data)

	err := r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: key,
		Values: map[string]any{
			"index":     chunkIndex,
			"data":      b64Data,
			"timestamp": time.Now().UnixMilli(),
		},
	}).Err()
	if err != nil {
		return fmt.Errorf("audio_buffer: append chunk: %w", err)
	}

	// Maintain 2-hour expiration on stream buffer
	r.client.Expire(ctx, key, 2*time.Hour)
	return nil
}

func (r *AudioBufferRepo) GetChunks(ctx context.Context, roomID, userID string) ([][]byte, error) {
	key := audioStreamKey(roomID, userID)
	messages, err := r.client.XRange(ctx, key, "-", "+").Result()
	if err != nil {
		return nil, fmt.Errorf("audio_buffer: get chunks: %w", err)
	}

	chunks := make([][]byte, 0, len(messages))
	for _, msg := range messages {
		rawB64, ok := msg.Values["data"].(string)
		if !ok {
			continue
		}
		data, err := base64.StdEncoding.DecodeString(rawB64)
		if err == nil {
			chunks = append(chunks, data)
		}
	}

	return chunks, nil
}

func (r *AudioBufferRepo) ClearBuffer(ctx context.Context, roomID, userID string) error {
	key := audioStreamKey(roomID, userID)
	return r.client.Del(ctx, key).Err()
}