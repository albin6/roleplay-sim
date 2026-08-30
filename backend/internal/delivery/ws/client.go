package ws

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"nhooyr.io/websocket"
)

const (
	writeWait      = 5 * time.Second
	sendBufferSize = 128
)

// Client represents an authenticated active WebSocket connection.
type Client struct {
	ID        string
	UserID    string
	Role      string
	conn      *websocket.Conn
	hub       *Hub
	send      chan []byte
	roomID    string
	seq       atomic.Int64
	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
	closed    bool
	mu        sync.RWMutex
}

// NewClient creates a new WebSocket client wrapper.
func NewClient(userID, role string, conn *websocket.Conn, hub *Hub) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	return &Client{
		ID:     "conn_" + uuid.New().String()[:8],
		UserID: userID,
		Role:   role,
		conn:   conn,
		hub:    hub,
		send:   make(chan []byte, sendBufferSize),
		ctx:    ctx,
		cancel: cancel,
	}
}

// SetRoomID sets the active room association for this client.
func (c *Client) SetRoomID(roomID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.roomID = roomID
}

// GetRoomID returns the current room ID.
func (c *Client) GetRoomID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.roomID
}

// Send serializes and enqueues an event envelope for delivery to the client.
func (c *Client) Send(event EventType, payload any) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.closed {
		return
	}

	seq := c.seq.Add(1)
	data, err := NewEnvelope(event, payload, seq)
	if err != nil {
		log.Error().Err(err).Str("event", string(event)).Msg("ws: failed to serialize envelope")
		return
	}

	select {
	case <-c.ctx.Done():
		return
	case c.send <- data:
	default:
		log.Warn().Str("user_id", c.UserID).Str("event", string(event)).Msg("ws: send buffer full, dropping message")
	}
}

// ReadPump listens for incoming messages from the WebSocket connection.
func (c *Client) ReadPump() {
	defer func() {
		c.Close()
		c.hub.Unregister(c)
	}()

	for {
		msgType, data, err := c.conn.Read(c.ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) && websocket.CloseStatus(err) != websocket.StatusNormalClosure {
				log.Debug().Err(err).Str("user_id", c.UserID).Msg("ws: read error")
			}
			break
		}

		if msgType == websocket.MessageBinary {
			// Binary frames: reserved for audio streaming
			c.hub.HandleBinaryAudio(c, data)
			continue
		}

		var env Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			log.Warn().Err(err).Str("user_id", c.UserID).Msg("ws: invalid json received")
			c.Send(EventError, ErrorPayload{
				Code:      "INVALID_FRAME",
				Message:   "Malformed JSON message framing",
				Retryable: false,
			})
			continue
		}

		c.hub.HandleMessage(c, &env)
	}
}

// WritePump drains the send channel and writes frames to the client.
func (c *Client) WritePump() {
	defer func() {
		c.Close()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg, ok := <-c.send:
			if !ok {
				_ = c.conn.Close(websocket.StatusNormalClosure, "server closing")
				return
			}

			writeCtx, writeCancel := context.WithTimeout(c.ctx, writeWait)
			err := c.conn.Write(writeCtx, websocket.MessageText, msg)
			writeCancel()

			if err != nil {
				if !errors.Is(err, context.Canceled) {
					log.Debug().Err(err).Str("user_id", c.UserID).Msg("ws: write error")
				}
				return
			}
		}
	}
}

// Close gracefully closes the client connection.
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.closed = true
		c.mu.Unlock()

		c.cancel()
		close(c.send)
		_ = c.conn.Close(websocket.StatusNormalClosure, "disconnect")
	})
}