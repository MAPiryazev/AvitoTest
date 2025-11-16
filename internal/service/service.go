package service

import (
	"context"

	"AvitoTest/internal/models"
	"AvitoTest/internal/repository"
)

type PRService interface {
	CreatePullRequest(ctx context.Context, pr models.PullRequest) (*models.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string) (*models.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID string, oldReviewerID int) (*models.PullRequest, error)
	GetPRsForReviewer(ctx context.Context, userID int) ([]models.PullRequestShort, error)
}

type TeamService interface {
	CreateTeam(ctx context.Context, team models.Team) error
	GetTeamByName(ctx context.Context, name string) (*models.Team, error)
	GetTeamOfUser(ctx context.Context, userID int) (*models.Team, error)
	GetTeamWithMembers(ctx context.Context, name string) ([]models.UserWithTeam, error)
}

type UserService interface {
	SetUserActive(ctx context.Context, userID string, isActive bool) error
	GetUserByPublicID(ctx context.Context, userID string) (*models.User, error)
	GetUserWithTeam(ctx context.Context, userID string) (*models.UserWithTeam, error)
}

type Service struct {
	PR   PRService
	Team TeamService
	User UserService
}

func NewService(repo *repository.Repository) *Service {
	return &Service{
		PR:   NewPRService(repo.PRRepo, repo.TeamRepo, repo.UserRepo),
		Team: NewTeamService(repo.TeamRepo, repo.UserRepo),
		User: NewUserService(repo.UserRepo, repo.TeamRepo),
	}
}
