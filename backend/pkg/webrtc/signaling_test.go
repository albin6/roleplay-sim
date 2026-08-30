package webrtc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSignal_Offer(t *testing.T) {
	raw := []byte(`{"type":"offer","sdp":"v=0\r\no=- 12345 2 IN IP4 127.0.0.1\r\n"}`)
	sig, err := ParseSignal(raw)
	require.NoError(t, err)
	assert.Equal(t, SignalTypeOffer, sig.Type)
	assert.Contains(t, sig.SDP, "v=0")
}

func TestParseSignal_Answer(t *testing.T) {
	raw := []byte(`{"type":"answer","sdp":"v=0\r\no=- 67890 2 IN IP4 127.0.0.1\r\n"}`)
	sig, err := ParseSignal(raw)
	require.NoError(t, err)
	assert.Equal(t, SignalTypeAnswer, sig.Type)
	assert.Contains(t, sig.SDP, "v=0")
}

func TestParseSignal_ICE(t *testing.T) {
	mid := "0"
	line := 0
	raw := []byte(`{"type":"ice","candidate":{"candidate":"candidate:1234 1 udp 2130706431 192.168.1.1 54321 typ host","sdpMid":"0","sdpMLineIndex":0}}`)
	sig, err := ParseSignal(raw)
	require.NoError(t, err)
	assert.Equal(t, SignalTypeICE, sig.Type)
	require.NotNil(t, sig.Candidate)
	assert.Equal(t, &mid, sig.Candidate.SDPMid)
	assert.Equal(t, &line, sig.Candidate.SDPMLineIndex)
}

func TestParseSignal_Invalid(t *testing.T) {
	// Missing SDP in offer
	_, err := ParseSignal([]byte(`{"type":"offer"}`))
	assert.ErrorIs(t, err, ErrMissingSDP)

	// Missing candidate in ICE
	_, err = ParseSignal([]byte(`{"type":"ice"}`))
	assert.ErrorIs(t, err, ErrMissingCandidate)

	// Invalid type
	_, err = ParseSignal([]byte(`{"type":"unknown"}`))
	assert.ErrorIs(t, err, ErrInvalidSignalType)
}