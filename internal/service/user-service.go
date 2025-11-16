package service

import (
	"context"
	"errors"
	"fmt"

	customerrors "AvitoTest/internal/custom-errors"
	"AvitoTest/internal/models"
	"AvitoTest/internal/repository"
)

// userService реализация UserService
type userService struct {
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
}

// NewUserService конструктор
func NewUserService(userRepo repository.UserRepository, teamRepo repository.TeamRepository) UserService {
	return &userService{
		userRepo: userRepo,
		teamRepo: teamRepo,
	}
}

// SetUserActive устанавливает флаг активности пользователя
func (s *userService) SetUserActive(ctx context.Context, userID string, isActive bool) error {
	err := s.userRepo.SetUserActive(ctx, userID, isActive)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			return fmt.Errorf("%w: user_id=%s", customerrors.ErrNotFound, userID)
		}
		if errors.Is(err, customerrors.ErrDBQuery) || errors.Is(err, customerrors.ErrDBScan) {
			return fmt.Errorf("%w: при обновлении активности пользователя %s", customerrors.ErrDBQuery, userID)
		}
		return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	return nil
}

// GetUserByPublicID возвращает пользователя по user_id
func (s *userService) GetUserByPublicID(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.userRepo.GetUserByPublicID(ctx, userID)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			return nil, fmt.Errorf("%w: user_id=%s", customerrors.ErrNotFound, userID)
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	return user, nil
}

func (s *userService) GetUserWithTeam(ctx context.Context, userID string) (*models.UserWithTeam, error) {
	u, err := s.userRepo.GetUserByPublicID(ctx, userID)
	if err != nil {
		return nil, err
	}

	team, err := s.teamRepo.GetTeamOfUser(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	return &models.UserWithTeam{
		ID:       u.ID,
		UserID:   u.UserID,
		Username: u.Username,
		TeamID:   u.TeamID,
		TeamName: team.Name,
		IsActive: u.IsActive,
	}, nil
}
