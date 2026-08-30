package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/roleplay-sim/backend/internal/domain/entity"
)

const roomTTL = 2 * time.Hour

type RoomStateRepo struct {
	client *redis.Client
}

func NewRoomStateRepo(client *redis.Client) *RoomStateRepo {
	return &RoomStateRepo{client: client}
}

func (r *RoomStateRepo) Create(ctx context.Context, room *entity.Room) error {
	data, err := json.Marshal(room)
	if err != nil {
		return fmt.Errorf("room_state: marshal: %w", err)
	}
	key := roomKey(room.ID)
	pipe := r.client.Pipeline()
	pipe.Set(ctx, key, data, roomTTL)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("room_state: create: %w", err)
	}
	return nil
}

func (r *RoomStateRepo) Get(ctx context.Context, roomID string) (*entity.Room, error) {
	data, err := r.client.Get(ctx, roomKey(roomID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("room_state: not found")
		}
		return nil, fmt.Errorf("room_state: get: %w", err)
	}
	var room entity.Room
	if err := json.Unmarshal(data, &room); err != nil {
		return nil, fmt.Errorf("room_state: unmarshal: %w", err)
	}
	return &room, nil
}

func (r *RoomStateRepo) UpdateState(ctx context.Context, roomID string, state entity.RoomState) error {
	room, err := r.Get(ctx, roomID)
	if err != nil {
		return err
	}
	room.State = state
	return r.Create(ctx, room)
}

func (r *RoomStateRepo) SetPeerConnected(ctx context.Context, roomID, userID string, connected bool) error {
	room, err := r.Get(ctx, roomID)
	if err != nil {
		return err
	}
	if room.PeerA.UserID == userID {
		room.PeerA.Connected = connected
	} else if room.PeerB.UserID == userID {
		room.PeerB.Connected = connected
	}
	return r.Create(ctx, room)
}

func (r *RoomStateRepo) Delete(ctx context.Context, roomID string) error {
	return r.client.Del(ctx, roomKey(roomID)).Err()
}

func (r *RoomStateRepo) Exists(ctx context.Context, roomID string) (bool, error) {
	n, err := r.client.Exists(ctx, roomKey(roomID)).Result()
	if err != nil {
		return false, fmt.Errorf("room_state: exists: %w", err)
	}
	return n > 0, nil
}

func roomKey(roomID string) string {
	return "room:" + roomID
}
