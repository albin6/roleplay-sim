package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	deliveryhttp "github.com/roleplay-sim/backend/internal/delivery/http"
	"github.com/roleplay-sim/backend/internal/delivery/http/middleware"
	"github.com/roleplay-sim/backend/internal/repository/postgres"
	redisrepo "github.com/roleplay-sim/backend/internal/repository/redis"
	authusecase "github.com/roleplay-sim/backend/internal/usecase/auth"
	leaderboardusecase "github.com/roleplay-sim/backend/internal/usecase/leaderboard"
	matchmakingusecase "github.com/roleplay-sim/backend/internal/usecase/matchmaking"
	"github.com/roleplay-sim/backend/internal/delivery/ws"
	"github.com/roleplay-sim/backend/internal/usecase/evaluation"
	"github.com/roleplay-sim/backend/pkg/config"
	"github.com/roleplay-sim/backend/pkg/evaluator"
	"github.com/roleplay-sim/backend/pkg/jwt"
	"github.com/roleplay-sim/backend/pkg/stt"
)

func main() {
	ctx := context.Background()

	// Config
	cfg, err := config.Load(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	// Logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if cfg.LogLevel == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// PostgreSQL Pool
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse DATABASE_URL")
	}
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to PostgreSQL")
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("PostgreSQL ping failed")
	}
	log.Info().Msg("connected to PostgreSQL")

	// Redis Client
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse REDIS_URL")
	}
	redisClient := redis.NewClient(opt)
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("Redis ping failed")
	}
	log.Info().Msg("connected to Redis")

	// Repositories
	userRepo := postgres.NewUserRepo(pool)
	sessionRepo := postgres.NewSessionRepo(pool)
	scenarioRepo := postgres.NewScenarioRepo(pool)
	evalRepo := postgres.NewEvaluationRepo(pool)
	sessionStore := redisrepo.NewSessionStoreRepo(redisClient)
	queueRepo := redisrepo.NewMatchmakingQueueRepo(redisClient)
	roomStateRepo := redisrepo.NewRoomStateRepo(redisClient)
	webrtcCacheRepo := redisrepo.NewWebRTCCacheRepo(redisClient)
	audioBufferRepo := redisrepo.NewAudioBufferRepo(redisClient)

	// Services
	jwtSvc := jwt.NewService(cfg.JWTPrivateKey, cfg.JWTPublicKey, cfg.JWTAccessTokenExpiry)

	// STT Provider Selection (Deepgram -> Whisper -> Mock fallback)
	var sttProvider stt.STTProvider
	if cfg.DeepgramAPIKey != "" && cfg.DeepgramAPIKey != "your_deepgram_api_key_here" {
		log.Info().Msg("stt: using Deepgram Nova-2 STT provider")
		sttProvider = stt.NewDeepgramClient(cfg.DeepgramAPIKey)
	} else if cfg.OpenAIAPIKey != "" && cfg.OpenAIAPIKey != "your_openai_api_key_here" {
		log.Info().Msg("stt: using OpenAI Whisper STT provider")
		sttProvider = stt.NewWhisperClient(cfg.OpenAIAPIKey)
	} else {
		log.Info().Msg("stt: using development Mock STT provider")
		sttProvider = stt.NewMockSTT()
	}

	// Evaluator Service & Use Case
	evalService := evaluator.NewEvaluatorService(cfg.GeminiAPIKey)
	evalUC := evaluation.NewEvaluateSessionUseCase(
		audioBufferRepo,
		sttProvider,
		evalService,
		userRepo,
		sessionRepo,
		evalRepo,
	)

	// Use Cases
	registerUC := authusecase.NewRegisterUseCase(userRepo, sessionStore, jwtSvc, cfg.JWTRefreshTokenExpiry)
	loginUC := authusecase.NewLoginUseCase(userRepo, sessionStore, jwtSvc, cfg.JWTRefreshTokenExpiry)
	enqueueUC := matchmakingusecase.NewEnqueueUseCase(queueRepo, userRepo)
	dequeueUC := matchmakingusecase.NewDequeueUseCase(queueRepo)
	leaderboardUC := leaderboardusecase.NewFetchLeaderboardUseCase(userRepo)

	// WebSocket Hub & Real-time Signaling
	wsHub := ws.NewHub(
		userRepo,
		sessionRepo,
		scenarioRepo,
		queueRepo,
		roomStateRepo,
		webrtcCacheRepo,
		audioBufferRepo,
		evalUC,
	)
	go wsHub.Run(ctx)
	wsHandler := ws.NewHandler(wsHub, jwtSvc, sessionStore)

	// HTTP
	authMW := middleware.NewAuthMiddleware(jwtSvc, sessionStore)
	router := deliveryhttp.NewRouter(deliveryhttp.Dependencies{
		RegisterUC:     registerUC,
		LoginUC:        loginUC,
		EnqueueUC:      enqueueUC,
		DequeueUC:      dequeueUC,
		LeaderboardUC:  leaderboardUC,
		UserRepo:       userRepo,
		SessionRepo:    sessionRepo,
		AuthMW:         authMW,
		WSHandler:      wsHandler,
		AllowedOrigins: []string{"http://localhost:3000"},
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("HTTP server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	<-quit
	log.Info().Msg("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}
	log.Info().Msg("server stopped")
}
