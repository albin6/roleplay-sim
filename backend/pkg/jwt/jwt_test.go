package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWT_GenerateAndValidate(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	svc := NewService(privateKey, &privateKey.PublicKey, 15*time.Minute)

	tokenStr, err := svc.GenerateAccessToken("user-123", "session-456", "player")
	require.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	claims, err := svc.ValidateAccessToken(tokenStr)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "session-456", claims.SessionID)
	assert.Equal(t, "player", claims.Role)
	assert.Equal(t, "roleplay-sim", claims.Issuer)
}

func TestJWT_ExpiredToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	svc := NewService(privateKey, &privateKey.PublicKey, -1*time.Minute)

	tokenStr, err := svc.GenerateAccessToken("user-123", "session-456", "player")
	require.NoError(t, err)

	_, err = svc.ValidateAccessToken(tokenStr)
	assert.Error(t, err, "Expired token should fail validation")
}

func TestJWT_InvalidSignature(t *testing.T) {
	key1, _ := rsa.GenerateKey(rand.Reader, 2048)
	key2, _ := rsa.GenerateKey(rand.Reader, 2048)

	svc1 := NewService(key1, &key1.PublicKey, 15*time.Minute)
	svc2 := NewService(key2, &key2.PublicKey, 15*time.Minute)

	tokenStr, _ := svc1.GenerateAccessToken("user-123", "session-456", "player")

	_, err := svc2.ValidateAccessToken(tokenStr)
	assert.Error(t, err, "Token signed by different key should fail")
}