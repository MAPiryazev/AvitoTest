package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	customerrors "AvitoTest/internal/custom-errors"
	"AvitoTest/internal/models"
)

// pgTeamRepository реализует репозиторий для работы с командами
type pgTeamRepository struct {
	db *sql.DB
}

func NewTeamRepositoryPG(db *sql.DB) *pgTeamRepository {
	return &pgTeamRepository{db: db}
}

// CreateTeam создает команду с пользователями
func (r *pgTeamRepository) CreateTeam(ctx context.Context, team models.Team) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && rollbackErr != sql.ErrTxDone {
			fmt.Printf("tx rollback error: %v\n", rollbackErr)
		}
	}()

	var teamID int
	err = tx.QueryRowContext(ctx,
		`INSERT INTO teams (team_name) VALUES ($1) RETURNING id`,
		team.Name,
	).Scan(&teamID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code.Name() == "unique_violation" {
			return customerrors.ErrAlreadyExists
		}
		return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	// вставка участников
	for _, user := range team.Members {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO users (user_id, username, team_id, is_active)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT(user_id) DO UPDATE
			SET username = EXCLUDED.username,
				is_active = EXCLUDED.is_active,
				team_id = EXCLUDED.team_id
		`, user.UserID, user.Username, teamID, user.IsActive)
		if err != nil {
			return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	return nil
}

// GetTeamByName возвращает команду по имени вместе с членами
func (r *pgTeamRepository) GetTeamByName(ctx context.Context, name string) (*models.Team, error) {
	team := models.Team{}

	// Получаем команду
	row := r.db.QueryRowContext(ctx, `SELECT id, team_name FROM teams WHERE team_name=$1`, name)
	if err := row.Scan(&team.ID, &team.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, customerrors.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBScan, err)
	}

	rows, err := r.db.QueryContext(ctx, `SELECT id, user_id, username, team_id, is_active FROM users WHERE team_id=$1`, team.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	defer safeClose(rows)

	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.UserID, &u.Username, &u.TeamID, &u.IsActive); err != nil {
			return nil, fmt.Errorf("%w: %v", customerrors.ErrDBScan, err)
		}
		team.Members = append(team.Members, u)
	}

	return &team, nil
}

// GetTeamOfUser возвращает команду, в которой состоит пользователь
func (r *pgTeamRepository) GetTeamOfUser(ctx context.Context, userID int) (*models.Team, error) {
	team := models.Team{}

	row := r.db.QueryRowContext(ctx,
		`SELECT t.id, t.team_name
		 FROM teams t
		 JOIN users u ON t.id = u.team_id
		 WHERE u.id = $1`,
		userID,
	)
	if err := row.Scan(&team.ID, &team.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, customerrors.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBScan, err)
	}

	rows, err := r.db.QueryContext(ctx, `SELECT id, user_id, username, team_id, is_active FROM users WHERE team_id=$1`, team.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	defer safeClose(rows)

	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.UserID, &u.Username, &u.TeamID, &u.IsActive); err != nil {
			return nil, fmt.Errorf("%w: %v", customerrors.ErrDBScan, err)
		}
		team.Members = append(team.Members, u)
	}

	return &team, nil
}
