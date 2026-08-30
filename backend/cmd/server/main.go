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
	"github.com/roleplay-sim/backend/pkg/config"
	"github.com/roleplay-sim/backend/pkg/jwt"
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
	if !cfg.IsProduction() {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	log.Info().Str("env", cfg.AppEnv).Str("port", cfg.Port).Msg("starting server")

	// PostgreSQL
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("postgres ping failed")
	}
	log.Info().Msg("postgres connected")

	// Redis
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse redis URL")
	}
	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatal().Err(err).Msg("redis ping failed")
	}
	log.Info().Msg("redis connected")

	// Repositories
	userRepo := postgres.NewUserRepo(pool)
	sessionRepo := postgres.NewSessionRepo(pool)
	scenarioRepo := postgres.NewScenarioRepo(pool)
	sessionStore := redisrepo.NewSessionStoreRepo(redisClient)
	queueRepo := redisrepo.NewMatchmakingQueueRepo(redisClient)
	roomStateRepo := redisrepo.NewRoomStateRepo(redisClient)
	webrtcCacheRepo := redisrepo.NewWebRTCCacheRepo(redisClient)

	// Services
	jwtSvc := jwt.NewService(cfg.JWTPrivateKey, cfg.JWTPublicKey, cfg.JWTAccessTokenExpiry)

	// Use Cases
	registerUC := authusecase.NewRegisterUseCase(userRepo, sessionStore, jwtSvc, cfg.JWTRefreshTokenExpiry)
	loginUC := authusecase.NewLoginUseCase(userRepo, sessionStore, jwtSvc, cfg.JWTRefreshTokenExpiry)
	enqueueUC := matchmakingusecase.NewEnqueueUseCase(queueRepo, userRepo)
	dequeueUC := matchmakingusecase.NewDequeueUseCase(queueRepo)
	leaderboardUC := leaderboardusecase.NewFetchLeaderboardUseCase(userRepo)

	// WebSocket Hub & Real-time Signaling
	wsHub := ws.NewHub(userRepo, sessionRepo, scenarioRepo, queueRepo, roomStateRepo, webrtcCacheRepo)
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
