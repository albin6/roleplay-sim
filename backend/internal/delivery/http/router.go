package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/roleplay-sim/backend/internal/delivery/http/handler"
	"github.com/roleplay-sim/backend/internal/delivery/http/middleware"
	authusecase "github.com/roleplay-sim/backend/internal/usecase/auth"
	leaderboardusecase "github.com/roleplay-sim/backend/internal/usecase/leaderboard"
	matchmakingusecase "github.com/roleplay-sim/backend/internal/usecase/matchmaking"
	"github.com/roleplay-sim/backend/internal/domain/repository"
)

type Dependencies struct {
	RegisterUC    *authusecase.RegisterUseCase
	LoginUC       *authusecase.LoginUseCase
	EnqueueUC     *matchmakingusecase.EnqueueUseCase
	DequeueUC     *matchmakingusecase.DequeueUseCase
	LeaderboardUC *leaderboardusecase.FetchLeaderboardUseCase
	UserRepo      repository.UserRepository
	SessionRepo   repository.SessionRepository
	AuthMW        *middleware.AuthMiddleware
	WSHandler     http.Handler
	AllowedOrigins []string
}

func NewRouter(deps Dependencies) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Compress(5))
	r.Use(middleware.CORS(deps.AllowedOrigins))

	// Health
	r.Get("/health", handler.HealthHandler)

	// WebSocket upgrade endpoint (both /ws and /v1/ws supported)
	if deps.WSHandler != nil {
		r.Handle("/ws", deps.WSHandler)
		r.Handle("/v1/ws", deps.WSHandler)
	}

	// API v1
	r.Route("/v1", func(r chi.Router) {
		// Public auth routes
		authH := handler.NewAuthHandler(deps.RegisterUC, deps.LoginUC, deps.UserRepo, deps.AuthMW)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authH.Register)
			r.Post("/login", authH.Login)
			r.Post("/refresh", authH.Refresh)
			r.Post("/logout", authH.Logout)
		})

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(deps.AuthMW.Authenticate)

			// Users
			userH := handler.NewUserHandler(deps.UserRepo, deps.SessionRepo)
			r.Route("/users", func(r chi.Router) {
				r.Get("/me", userH.GetMe)
				r.Patch("/me", userH.UpdateMe)
				r.Get("/me/history", userH.GetHistory)
				r.Get("/me/history/{sessionID}", userH.GetHistoryDetail)
			})

			// Matchmaking
			mmH := handler.NewMatchmakingHandler(deps.EnqueueUC, deps.DequeueUC, deps.UserRepo)
			r.Route("/matchmaking", func(r chi.Router) {
				r.Post("/enqueue", mmH.Enqueue)
				r.Delete("/dequeue", mmH.Dequeue)
				r.Get("/status", mmH.Status)
			})
		})

		// Leaderboard (Public with optional user rank highlighting)
		lbH := handler.NewLeaderboardHandler(deps.LeaderboardUC)
		r.With(deps.AuthMW.OptionalAuthenticate).Get("/leaderboard", lbH.GetLeaderboard)
	})

	return r
}
