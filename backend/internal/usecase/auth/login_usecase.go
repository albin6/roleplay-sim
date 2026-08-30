package auth

import (
	"context"
	"fmt"

	domainerrors "github.com/roleplay-sim/backend/internal/domain/errors"
	"github.com/roleplay-sim/backend/internal/domain/repository"
	"github.com/roleplay-sim/backend/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUseCase struct {
	userRepo     repository.UserRepository
	sessionStore repository.SessionStoreRepository
	jwtSvc       *jwt.Service
	refreshTTL   time.Duration
}

func NewLoginUseCase(
	userRepo repository.UserRepository,
	sessionStore repository.SessionStoreRepository,
	jwtSvc *jwt.Service,
	refreshTTL time.Duration,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo:     userRepo,
		sessionStore: sessionStore,
		jwtSvc:       jwtSvc,
		refreshTTL:   refreshTTL,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, in LoginInput) (*AuthResult, error) {
	user, err := uc.userRepo.GetByEmail(ctx, in.Email)
	if err != nil {
		return nil, domainerrors.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		return nil, domainerrors.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, fmt.Errorf("%w: account is disabled", domainerrors.ErrInvalidCredentials)
	}

	// Reuse register's session creation logic
	reg := &RegisterUseCase{
		userRepo:     uc.userRepo,
		sessionStore: uc.sessionStore,
		jwtSvc:       uc.jwtSvc,
		refreshTTL:   uc.refreshTTL,
	}
	return reg.createSession(ctx, user)
}
