package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/lib/pq"

	customerrors "AvitoTest/internal/custom-errors"
	"AvitoTest/internal/models"
)

type pgPullRequestRepository struct {
	db *sql.DB
}

func NewPullRequestRepositoryPG(db *sql.DB) *pgPullRequestRepository {
	return &pgPullRequestRepository{db: db}
}

// CreatePullRequest создает PR и автоматически назначает до 2 ревьюверов из команды автора
func (r *pgPullRequestRepository) CreatePullRequest(ctx context.Context, pr models.PullRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	defer tx.Rollback()

	// 1. Вставляем PR
	err = tx.QueryRowContext(ctx,
		`INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id)
		VALUES ($1, $2, $3) RETURNING id, created_at`,
		pr.PullRequestID, pr.Name, pr.AuthorID).Scan(&pr.ID, &pr.CreatedAt)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code.Name() == "unique_violation" {
			return customerrors.ErrAlreadyExists
		}
		return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	// 2. Получаем команду автора
	rows, err := tx.QueryContext(ctx,
		`SELECT u.id, u.user_id, u.username, u.team_id, u.is_active
		 FROM users u
		 JOIN teams t ON u.team_id = t.id
		 WHERE u.team_id = (SELECT team_id FROM users WHERE id=$1) AND u.is_active = TRUE AND u.id <> $1`,
		pr.AuthorID)
	if err != nil {
		return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	defer rows.Close()

	activeMembers := make([]int, 0)
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.UserID, &u.Username, &u.TeamID, &u.IsActive); err != nil {
			return fmt.Errorf("%w: %v", customerrors.ErrDBScan, err)
		}
		activeMembers = append(activeMembers, u.ID)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("%w: %v", customerrors.ErrDBScan, err)
	}

	// 3. Выбираем случайно до 2 ревьюверов
	rand.Seed(time.Now().UnixNano())
	n := len(activeMembers)
	reviewers := make([]int, 0)
	if n >= 2 {
		for i := 0; i < 2; i++ {
			idx := rand.Intn(len(activeMembers))
			reviewers = append(reviewers, activeMembers[idx])
			activeMembers = append(activeMembers[:idx], activeMembers[idx+1:]...)
		}
	} else if n == 1 {
		reviewers = append(reviewers, activeMembers[0])
	}

	// 4. Вставляем в pull_request_reviewers
	for _, rid := range reviewers {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)`,
			pr.ID, rid)
		if err != nil {
			return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	return nil
}

// GetPullRequestByPublicID возвращает PR по pull_request_id
func (r *pgPullRequestRepository) GetPullRequestByPublicID(ctx context.Context, prID string) (*models.PullRequest, error) {
	pr := models.PullRequest{}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		 FROM pull_requests WHERE pull_request_id=$1`,
		prID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("%w: pull_request_id=%s", customerrors.ErrNotFound, prID)
	}

	if err := rows.Scan(&pr.ID, &pr.PullRequestID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBScan, err)
	}

	// Получаем ревьюверов с их user_id
	reviewerRows, err := r.db.QueryContext(ctx,
		`SELECT u.id, u.user_id FROM pull_request_reviewers prr 
		 JOIN users u ON prr.reviewer_id = u.id 
		 WHERE prr.pull_request_id=$1`, pr.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	defer reviewerRows.Close()

	pr.ReviewerIDs = make([]int, 0)
	for reviewerRows.Next() {
		var rid int
		var userID string
		if err := reviewerRows.Scan(&rid, &userID); err != nil {
			return nil, fmt.Errorf("%w: %v", customerrors.ErrDBScan, err)
		}
		pr.ReviewerIDs = append(pr.ReviewerIDs, rid)
	}

	return &pr, nil
}

// SetPullRequestMerged помечает PR как MERGED, идемпотентно
func (r *pgPullRequestRepository) SetPullRequestMerged(ctx context.Context, prID string) (*models.PullRequest, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	defer tx.Rollback()

	pr, err := r.GetPullRequestByPublicID(ctx, prID)
	if err != nil {
		return nil, err
	}

	if pr.Status == models.PRStatusMerged {
		return pr, nil
	}

	mergedAt := time.Now()
	_, err = tx.ExecContext(ctx,
		`UPDATE pull_requests SET status='MERGED', merged_at=$1 WHERE id=$2`,
		mergedAt, pr.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	pr.Status = models.PRStatusMerged
	pr.MergedAt = &mergedAt

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	return pr, nil
}

// UpdateReviewers заменяет список ревьюверов PR, учитывая правила
func (r *pgPullRequestRepository) UpdateReviewers(ctx context.Context, prID int, newReviewerIDs []int) error {
	pr := models.PullRequest{}
	err := r.db.QueryRowContext(ctx, `SELECT status FROM pull_requests WHERE id=$1`, prID).Scan(&pr.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: prID=%d", customerrors.ErrNotFound, prID)
		}
		return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	if pr.Status == models.PRStatusMerged {
		return customerrors.ErrForbidden
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM pull_request_reviewers WHERE pull_request_id=$1`, prID)
	if err != nil {
		return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	for _, rid := range newReviewerIDs {
		_, err := tx.ExecContext(ctx, `INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)`, prID, rid)
		if err != nil {
			return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	return nil
}

// GetPRsWhereUserReviewer возвращает PR, где пользователь назначен ревьювером
func (r *pgPullRequestRepository) GetPRsWhereUserReviewer(ctx context.Context, userID int) ([]models.PullRequestShort, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT pr.id, pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		 FROM pull_requests pr
		 JOIN pull_request_reviewers prr ON pr.id = prr.pull_request_id
		 WHERE prr.reviewer_id=$1`, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	defer rows.Close()

	result := make([]models.PullRequestShort, 0)
	for rows.Next() {
		var pr models.PullRequestShort
		if err := rows.Scan(&pr.ID, &pr.PullRequestID, &pr.Name, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("%w: %v", customerrors.ErrDBScan, err)
		}
		result = append(result, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBScan, err)
	}

	return result, nil
}

// GetActiveTeamMembersExcept возвращает активных участников команды, исключая указанных
func (r *pgPullRequestRepository) GetActiveTeamMembersExcept(ctx context.Context, teamID int, exclude []int) ([]models.User, error) {
	query := `SELECT id, user_id, username, team_id, is_active FROM users WHERE team_id=$1 AND is_active=TRUE`
	args := []interface{}{teamID}

	if len(exclude) > 0 {
		query += " AND id NOT IN ("
		for i := range exclude {
			if i > 0 {
				query += ","
			}
			query += fmt.Sprintf("$%d", i+2)
			args = append(args, exclude[i])
		}
		query += ")"
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	defer rows.Close()

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
