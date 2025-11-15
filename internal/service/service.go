package service

import (
	"context"

	"AvitoTest/internal/models"
	"AvitoTest/internal/repository"
)

type PRService interface {
	// Создать PR и автоматически назначить до 2 активных ревьюверов из команды автора
	CreatePullRequest(ctx context.Context, pr models.PullRequest) (*models.PullRequest, error)
	// Пометить PR как MERGED идемпотентно
	MergePullRequest(ctx context.Context, prID string) (*models.PullRequest, error)
	// Переназначить одного ревьювера на другого из его команды
	// исключая автора и других ревьюверов
	ReassignReviewer(ctx context.Context, prID string, oldReviewerID int) (*models.PullRequest, error)
	// Получить список PR, где пользователь назначен ревьювером
	GetPRsForReviewer(ctx context.Context, userID int) ([]models.PullRequestShort, error)
}

type TeamService interface {
	// Создать команду с пользователями (создает или обновляет пользователей)
	CreateTeam(ctx context.Context, team models.Team) error
	// Получить команду с участниками по имени
	GetTeamByName(ctx context.Context, name string) (*models.Team, error)
	// Получить команду пользователя по его внутреннему ID
	GetTeamOfUser(ctx context.Context, userID int) (*models.Team, error)
	// Получить команду с названием
	GetTeamWithMembers(ctx context.Context, name string) ([]models.UserWithTeam, error)
}

type UserService interface {
	// Установить флаг активности пользователя
	SetUserActive(ctx context.Context, userID string, isActive bool) error
	// Получить пользователя по user_id
	GetUserByPublicID(ctx context.Context, userID string) (*models.User, error)
	// Получить пользователя с именем команды
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
