package auth

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/roleplay-sim/backend/internal/domain/entity"
	domainerrors "github.com/roleplay-sim/backend/internal/domain/errors"
	"github.com/roleplay-sim/backend/internal/domain/repository"
	"github.com/roleplay-sim/backend/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type AuthResult struct {
	User        *entity.User
	AccessToken string
	SessionID   string
	ExpiresIn   int // seconds
}

type RegisterUseCase struct {
	userRepo     repository.UserRepository
	sessionStore repository.SessionStoreRepository
	jwtSvc       *jwt.Service
	refreshTTL   time.Duration
}

func NewRegisterUseCase(
	userRepo repository.UserRepository,
	sessionStore repository.SessionStoreRepository,
	jwtSvc *jwt.Service,
	refreshTTL time.Duration,
) *RegisterUseCase {
	return &RegisterUseCase{
		userRepo:     userRepo,
		sessionStore: sessionStore,
		jwtSvc:       jwtSvc,
		refreshTTL:   refreshTTL,
	}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	// Validate input
	if err := validateRegisterInput(in); err != nil {
		return nil, err
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("register: hash password: %w", err)
	}

	// Create user entity
	user := &entity.User{
		Username:     strings.ToLower(strings.TrimSpace(in.Username)),
		Email:        strings.ToLower(strings.TrimSpace(in.Email)),
		PasswordHash: string(hash),
		DisplayName:  strings.TrimSpace(in.DisplayName),
		EloRating:    1200.0,
		Role:         entity.UserRolePlayer,
		IsActive:     true,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err // already wrapped (ErrUserAlreadyExists)
	}

	return uc.createSession(ctx, user)
}

func (uc *RegisterUseCase) createSession(ctx context.Context, user *entity.User) (*AuthResult, error) {
	sessionID := uuid.New().String()

	// Store in Redis
	if err := uc.sessionStore.Create(ctx, sessionID, user.ID.String(), uc.refreshTTL); err != nil {
		return nil, fmt.Errorf("register: create session: %w", err)
	}
	if err := uc.sessionStore.AddUserSession(ctx, user.ID.String(), sessionID, uc.refreshTTL); err != nil {
		return nil, fmt.Errorf("register: track session: %w", err)
	}

	// Generate JWT
	token, err := uc.jwtSvc.GenerateAccessToken(user.ID.String(), sessionID, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("register: generate token: %w", err)
	}

	return &AuthResult{
		User:        user,
		AccessToken: token,
		SessionID:   sessionID,
		ExpiresIn:   900, // 15 minutes
	}, nil
}

func validateRegisterInput(in RegisterInput) error {
	if len(in.Username) < 3 || len(in.Username) > 50 {
		return fmt.Errorf("%w: username must be 3-50 characters", domainerrors.ErrInvalidCredentials)
	}
	for _, r := range in.Username {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return fmt.Errorf("%w: username may only contain letters, digits, and underscores", domainerrors.ErrInvalidCredentials)
		}
	}
	if len(in.Password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", domainerrors.ErrInvalidCredentials)
	}
	if !strings.Contains(in.Email, "@") {
		return fmt.Errorf("%w: invalid email", domainerrors.ErrInvalidCredentials)
	}
	if len(in.DisplayName) < 2 || len(in.DisplayName) > 100 {
		return fmt.Errorf("%w: display name must be 2-100 characters", domainerrors.ErrInvalidCredentials)
	}
	return nil
}
