package repository

import (
	"context"

	"AvitoTest/internal/models"
)

type UserRepository interface {
	CreateOrUpdateUser(ctx context.Context, u models.User) error
	SetUserActive(ctx context.Context, userID string, isActive bool) error
	GetUserByPublicID(ctx context.Context, userID string) (*models.User, error)
	GetUsersByTeamID(ctx context.Context, teamID int) ([]models.User, error)
}

type TeamRepository interface {
	CreateTeam(ctx context.Context, team models.Team) error
	GetTeamByName(ctx context.Context, name string) (*models.Team, error)
	GetTeamOfUser(ctx context.Context, userID int) (*models.Team, error)
}

type PullRequestRepository interface {
	CreatePullRequest(ctx context.Context, pr models.PullRequest) error
	GetPullRequestByPublicID(ctx context.Context, prID string) (*models.PullRequest, error)
	SetPullRequestMerged(ctx context.Context, prID string) (*models.PullRequest, error)
	UpdateReviewers(ctx context.Context, prID int, newReviewerIDs []int) error
	GetPRsWhereUserReviewer(ctx context.Context, userID int) ([]models.PullRequestShort, error)
	GetActiveTeamMembersExcept(ctx context.Context, teamID int, exclude []int) ([]models.User, error)
}

type Repository struct {
	UserRepo UserRepository
	TeamRepo TeamRepository
	PRRepo   PullRequestRepository
}
