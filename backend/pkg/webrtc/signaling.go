package webrtc

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrInvalidSignalType = errors.New("invalid signal type: must be offer, answer, or ice")
	ErrMissingSDP        = errors.New("missing sdp in offer or answer signal")
	ErrMissingCandidate  = errors.New("missing candidate in ice signal")
)

// SignalType defines the WebRTC signal kind.
type SignalType string

const (
	SignalTypeOffer  SignalType = "offer"
	SignalTypeAnswer SignalType = "answer"
	SignalTypeICE    SignalType = "ice"
)

// ICECandidate holds an interactive connectivity establishment candidate.
type ICECandidate struct {
	Candidate        string  `json:"candidate"`
	SDPMid           *string `json:"sdpMid,omitempty"`
	SDPMLineIndex    *int    `json:"sdpMLineIndex,omitempty"`
	UsernameFragment *string `json:"usernameFragment,omitempty"`
}

// SignalPayload is the payload relayed between peers during WebRTC handshake.
type SignalPayload struct {
	Type      SignalType    `json:"type"`
	SDP       string        `json:"sdp,omitempty"`
	Candidate *ICECandidate `json:"candidate,omitempty"`
}

// Validate ensures the signal payload matches WebRTC conventions.
func (s *SignalPayload) Validate() error {
	switch s.Type {
	case SignalTypeOffer, SignalTypeAnswer:
		if s.SDP == "" {
			return ErrMissingSDP
		}
	case SignalTypeICE:
		if s.Candidate == nil || s.Candidate.Candidate == "" {
			return ErrMissingCandidate
		}
	default:
		return fmt.Errorf("%w: %s", ErrInvalidSignalType, s.Type)
	}
	return nil
}

// ParseSignal deserializes a raw JSON message into a SignalPayload and validates it.
func ParseSignal(data []byte) (*SignalPayload, error) {
	var payload SignalPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("webrtc: parse signal: %w", err)
	}
	if err := payload.Validate(); err != nil {
		return nil, err
	}
	return &payload, nil
}