package service

import (
	"context"
	"errors"
	"fmt"

	customerrors "AvitoTest/internal/custom-errors"
	"AvitoTest/internal/models"
	"AvitoTest/internal/repository"
)

// teamService — реализация TeamService
type teamService struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
}

// NewTeamService — конструктор
func NewTeamService(teamRepo repository.TeamRepository, userRepo repository.UserRepository) TeamService {
	return &teamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

// CreateTeam создаёт команду с участниками (создаёт или обновляет пользователей)
func (s *teamService) CreateTeam(ctx context.Context, team models.Team) error {
	if err := s.teamRepo.CreateTeam(ctx, team); err != nil {
		if errors.Is(err, customerrors.ErrAlreadyExists) {
			return fmt.Errorf("%w: команда с именем %s уже существует", customerrors.ErrAlreadyExists, team.Name)
		}
		if errors.Is(err, customerrors.ErrDBQuery) {
			return fmt.Errorf("%w: при создании команды %s", customerrors.ErrDBQuery, team.Name)
		}
		return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	return nil
}

// GetTeamByName возвращает команду с участниками по имени
func (s *teamService) GetTeamByName(ctx context.Context, name string) (*models.Team, error) {
	team, err := s.teamRepo.GetTeamByName(ctx, name)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			return nil, fmt.Errorf("%w: команда %s не найдена", customerrors.ErrNotFound, name)
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	return team, nil
}

// GetTeamOfUser возвращает команду пользователя по его внутреннему ID
func (s *teamService) GetTeamOfUser(ctx context.Context, userID int) (*models.Team, error) {
	team, err := s.teamRepo.GetTeamOfUser(ctx, userID)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			return nil, fmt.Errorf("%w: команда для user_id=%d не найдена", customerrors.ErrNotFound, userID)
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	return team, nil
}

// GetTeamWithMembers вернет нам еще название команды
func (s *teamService) GetTeamWithMembers(ctx context.Context, name string) ([]models.UserWithTeam, error) {
	team, err := s.teamRepo.GetTeamByName(ctx, name)
	if err != nil {
		return nil, err
	}

	members := make([]models.UserWithTeam, len(team.Members))
	for i, u := range team.Members {
		members[i] = models.UserWithTeam{
			ID:       u.ID,
			UserID:   u.UserID,
			Username: u.Username,
			TeamID:   u.TeamID,
			TeamName: team.Name,
			IsActive: u.IsActive,
		}
	}

	return members, nil
}
