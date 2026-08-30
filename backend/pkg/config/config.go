package config

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sethvargo/go-envconfig"
)

// Config holds all application configuration, populated from environment variables.
type Config struct {
	// Application
	AppEnv   string `env:"APP_ENV,default=development"`
	LogLevel string `env:"LOG_LEVEL,default=info"`
	Port     string `env:"PORT,default=8080"`

	// Database
	DatabaseURL string `env:"DATABASE_URL,default=postgresql://roleplay:roleplay_pass@127.0.0.1:5433/roleplay_db?sslmode=disable"`

	// Redis
	RedisURL string `env:"REDIS_URL,default=redis://localhost:6379"`

	// JWT
	JWTPrivateKeyPath     string        `env:"JWT_PRIVATE_KEY_PATH,default=./secrets/jwt_private.pem"`
	JWTPublicKeyPath      string        `env:"JWT_PUBLIC_KEY_PATH,default=./secrets/jwt_public.pem"`
	JWTAccessTokenExpiry  time.Duration `env:"JWT_ACCESS_TOKEN_EXPIRY,default=15m"`
	JWTRefreshTokenExpiry time.Duration `env:"JWT_REFRESH_TOKEN_EXPIRY,default=168h"`

	// AI Services
	DeepgramAPIKey string `env:"DEEPGRAM_API_KEY"`
	GeminiAPIKey   string `env:"GEMINI_API_KEY"`
	OpenAIAPIKey   string `env:"OPENAI_API_KEY"`

	// TURN
	TURNServerURL string `env:"TURN_SERVER_URL"`
	TURNSecret    string `env:"TURN_SECRET"`

	// Rate limiting
	RateLimitRequestsPerMin int `env:"RATE_LIMIT_REQUESTS_PER_MIN,default=60"`

	// Loaded at runtime (not from env)
	JWTPrivateKey *rsa.PrivateKey
	JWTPublicKey  *rsa.PublicKey
}

// Load reads configuration from environment variables (loading .env if found) and loads RSA keys.
func Load(ctx context.Context) (*Config, error) {
	// Attempt to load .env from current dir or parent dirs
	loadDotEnv()

	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return nil, fmt.Errorf("config: failed to process env: %w", err)
	}

	privateKey, err := loadOrCreatePrivateKey(cfg.JWTPrivateKeyPath, cfg.JWTPublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("config: failed to load or create JWT RSA keys: %w", err)
	}
	cfg.JWTPrivateKey = privateKey
	cfg.JWTPublicKey = &privateKey.PublicKey

	return &cfg, nil
}

func loadDotEnv() {
	candidates := []string{".env", "../.env", "backend/.env"}
	for _, path := range candidates {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				k := strings.TrimSpace(parts[0])
				v := strings.TrimSpace(parts[1])
				v = strings.Trim(v, `"'`)
				if os.Getenv(k) == "" {
					_ = os.Setenv(k, v)
				}
			}
		}
		break
	}
}

func loadOrCreatePrivateKey(privPath, pubPath string) (*rsa.PrivateKey, error) {
	if _, err := os.Stat(privPath); os.IsNotExist(err) {
		// Auto-generate 2048-bit RSA key pair for development
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("generate rsa key: %w", err)
		}

		if err := os.MkdirAll(filepath.Dir(privPath), 0700); err != nil {
			return nil, fmt.Errorf("create secrets dir: %w", err)
		}

		// Save private key
		privBytes := x509.MarshalPKCS1PrivateKey(key)
		privPem := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privBytes,
		})
		if err := os.WriteFile(privPath, privPem, 0600); err != nil {
			return nil, fmt.Errorf("save private key: %w", err)
		}

		// Save public key
		pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("marshal public key: %w", err)
		}
		pubPem := pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubBytes,
		})
		if err := os.WriteFile(pubPath, pubPem, 0644); err != nil {
			return nil, fmt.Errorf("save public key: %w", err)
		}

		return key, nil
	}

	data, err := os.ReadFile(privPath)
	if err != nil {
		return nil, fmt.Errorf("read key file %q: %w", privPath, err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from %q", privPath)
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		pkcs1Key, err2 := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("parse private key: %w (also tried PKCS1: %v)", err, err2)
		}
		return pkcs1Key, nil
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key at %q is not RSA", privPath)
	}
	return rsaKey, nil
}

// IsProduction returns true when running in production mode.
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}