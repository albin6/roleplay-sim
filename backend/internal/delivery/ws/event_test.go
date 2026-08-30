package ws

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEnvelope(t *testing.T) {
	payload := RoomReadyPayload{
		RoomID:          "room_123",
		PeerDisplayName: "Bob Lead",
		PeerAvatarURL:   "https://cdn.example.com/bob.png",
		PeerEloRating:   1450.0,
		Seat:            "A",
	}

	bytes, err := NewEnvelope(EventRoomReady, payload, 1)
	require.NoError(t, err)

	var env Envelope
	err = json.Unmarshal(bytes, &env)
	require.NoError(t, err)
	assert.Equal(t, EventRoomReady, env.Event)
	assert.Equal(t, int64(1), env.Seq)
	assert.NotEmpty(t, env.Timestamp)

	var resPayload RoomReadyPayload
	err = json.Unmarshal(env.Payload, &resPayload)
	require.NoError(t, err)
	assert.Equal(t, "room_123", resPayload.RoomID)
	assert.Equal(t, "Bob Lead", resPayload.PeerDisplayName)
	assert.Equal(t, 1450.0, resPayload.PeerEloRating)
	assert.Equal(t, "A", resPayload.Seat)
}

func TestNewEnvelope_NilPayload(t *testing.T) {
	bytes, err := NewEnvelope(EventJoinQueue, nil, 5)
	require.NoError(t, err)

	var env Envelope
	err = json.Unmarshal(bytes, &env)
	require.NoError(t, err)
	assert.Equal(t, EventJoinQueue, env.Event)
	assert.Equal(t, int64(5), env.Seq)
	assert.Equal(t, string(env.Payload), "{}")
}