package jwt

import (
	"crypto/rsa"
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims contains the standard JWT fields plus our custom claims.
type Claims struct {
	gojwt.RegisteredClaims
	UserID    string `json:"uid"`
	SessionID string `json:"sid"` // Redis session key for revocation check
	Role      string `json:"role"`
}

// Service handles JWT generation and validation using RS256.
type Service struct {
	privateKey  *rsa.PrivateKey
	publicKey   *rsa.PublicKey
	accessExpiry time.Duration
}

// NewService creates a JWT service with the provided RSA key pair.
func NewService(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, accessExpiry time.Duration) *Service {
	return &Service{
		privateKey:   privateKey,
		publicKey:    publicKey,
		accessExpiry: accessExpiry,
	}
}

// GenerateAccessToken creates a signed RS256 JWT access token.
func (s *Service) GenerateAccessToken(userID, sessionID, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: gojwt.RegisteredClaims{
			ID:        uuid.New().String(),
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(now.Add(s.accessExpiry)),
			Issuer:    "roleplay-sim",
		},
		UserID:    userID,
		SessionID: sessionID,
		Role:      role,
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("jwt: sign token: %w", err)
	}
	return signed, nil
}

// ValidateAccessToken parses and validates an RS256 JWT, returning the claims.
func (s *Service) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := gojwt.ParseWithClaims(tokenStr, &Claims{}, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("jwt: unexpected signing method: %v", t.Header["alg"])
		}
		return s.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("jwt: parse: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("jwt: invalid token")
	}
	return claims, nil
}
