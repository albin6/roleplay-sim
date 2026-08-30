package leaderboard

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/roleplay-sim/backend/internal/domain/entity"
	"github.com/roleplay-sim/backend/internal/domain/repository"
)

type FetchResult struct {
	Users      []*entity.User
	Total      int64
	MyRank     int64
	MyElo      float64
}

type FetchLeaderboardUseCase struct {
	userRepo repository.UserRepository
}

func NewFetchLeaderboardUseCase(userRepo repository.UserRepository) *FetchLeaderboardUseCase {
	return &FetchLeaderboardUseCase{userRepo: userRepo}
}

func (uc *FetchLeaderboardUseCase) Execute(ctx context.Context, requesterID uuid.UUID, limit, offset int) (*FetchResult, error) {
	users, total, err := uc.userRepo.GetLeaderboard(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("leaderboard: %w", err)
	}

	myRank, err := uc.userRepo.GetRank(ctx, requesterID)
	if err != nil {
		myRank = 0
	}

	myUser, err := uc.userRepo.GetByID(ctx, requesterID)
	var myElo float64
	if err == nil {
		myElo = myUser.EloRating
	}

	return &FetchResult{
		Users:  users,
		Total:  total,
		MyRank: myRank,
		MyElo:  myElo,
	}, nil
}
