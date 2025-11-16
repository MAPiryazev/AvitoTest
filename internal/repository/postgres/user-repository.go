package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	customerrors "AvitoTest/internal/custom-errors"
	"AvitoTest/internal/models"
)

// pgUserRepository реализует репозиторий для работы с пользователями
type pgUserRepository struct {
	db *sql.DB
}

func NewUserRepositoryPG(db *sql.DB) *pgUserRepository {
	return &pgUserRepository{db: db}
}

// CreateOrUpdateUser создает или обновляет пользователя, проверяя существование команды
func (r *pgUserRepository) CreateOrUpdateUser(ctx context.Context, u models.User) error {
	// Проверяем, что команда существует
	var teamExists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM teams WHERE id=$1)`, u.TeamID).Scan(&teamExists)
	if err != nil {
		return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	if !teamExists {
		return fmt.Errorf("%w: team_id=%d does not exist", customerrors.ErrNotFound, u.TeamID)
	}

	query := `
	INSERT INTO users (user_id, username, team_id, is_active)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (user_id)
	DO UPDATE SET username = EXCLUDED.username, team_id = EXCLUDED.team_id, is_active = EXCLUDED.is_active
	`

	_, err = r.db.ExecContext(ctx, query, u.UserID, u.Username, u.TeamID, u.IsActive)
	if err != nil {
		return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	return nil
}

// SetUserActive устанавливает флаг активности пользователя
func (r *pgUserRepository) SetUserActive(ctx context.Context, userID string, isActive bool) error {
	query := `
	UPDATE users
	SET is_active = $1
	WHERE user_id = $2
	`

	res, err := r.db.ExecContext(ctx, query, isActive, userID)
	if err != nil {
		return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %v", customerrors.ErrDBScan, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: user_id=%s", customerrors.ErrNotFound, userID)
	}

	return nil
}

// GetUserByPublicID возвращает пользователя по user_id
func (r *pgUserRepository) GetUserByPublicID(ctx context.Context, userID string) (*models.User, error) {
	query := `
	SELECT id, user_id, username, team_id, is_active
	FROM users
	WHERE user_id = $1
	`

	var u models.User
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&u.ID, &u.UserID, &u.Username, &u.TeamID, &u.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: user_id=%s", customerrors.ErrNotFound, userID)
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBScan, err)
	}

	return &u, nil
}

// GetUsersByTeamID возвращает всех пользователей команды
func (r *pgUserRepository) GetUsersByTeamID(ctx context.Context, teamID int) ([]models.User, error) {
	query := `
	SELECT id, user_id, username, team_id, is_active
	FROM users
	WHERE team_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	defer safeClose(rows)

	users := make([]models.User, 0)
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.UserID, &u.Username, &u.TeamID, &u.IsActive); err != nil {
			return nil, fmt.Errorf("%w: %v", customerrors.ErrDBScan, err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBScan, err)
	}

	return users, nil
}
