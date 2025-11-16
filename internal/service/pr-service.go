package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	customerrors "AvitoTest/internal/custom-errors"
	"AvitoTest/internal/models"
	"AvitoTest/internal/repository"

	"github.com/google/uuid"
)

// prService реализация PRService
type prService struct {
	prRepo   repository.PullRequestRepository
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
	rnd      *rand.Rand
}

// NewPRService конструктор
func NewPRService(prRepo repository.PullRequestRepository, teamRepo repository.TeamRepository, userRepo repository.UserRepository) PRService {
	src := rand.NewSource(time.Now().UnixNano())
	return &prService{
		prRepo:   prRepo,
		teamRepo: teamRepo,
		userRepo: userRepo,
		rnd:      rand.New(src),
	}
}

// CreatePullRequest создаёт PR и назначает до 2 активных ревьюверов
func (s *prService) CreatePullRequest(ctx context.Context, pr models.PullRequest) (*models.PullRequest, error) {
	// Если pull_request_id не указан, генерируем новый
	if pr.PullRequestID == "" {
		pr.PullRequestID = uuid.New().String()
	}
	pr.Status = models.PRStatusOpen
	pr.CreatedAt = time.Now()

	team, err := s.teamRepo.GetTeamOfUser(ctx, pr.AuthorID)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			return nil, fmt.Errorf("%w: author_id=%d", customerrors.ErrNotFound, pr.AuthorID)
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	// кандидаты: активные и не автор
	candidates := []int{}
	for _, u := range team.Members {
		if u.ID != pr.AuthorID && u.IsActive {
			candidates = append(candidates, u.ID)
		}
	}

	pr.ReviewerIDs = chooseRandomUpToN(s.rnd, candidates, 2)

	if err := s.prRepo.CreatePullRequest(ctx, pr); err != nil {
		if errors.Is(err, customerrors.ErrAlreadyExists) {
			return nil, fmt.Errorf("%w: pull_request_id=%s", customerrors.ErrAlreadyExists, pr.PullRequestID)
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	created, err := s.prRepo.GetPullRequestByPublicID(ctx, pr.PullRequestID)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			return nil, fmt.Errorf("%w: созданный PR не найден", customerrors.ErrNotFound)
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	if len(pr.ReviewerIDs) > 0 {
		if err := s.prRepo.UpdateReviewers(ctx, created.ID, pr.ReviewerIDs); err != nil {
			return nil, fmt.Errorf("%w: не удалось сохранить ревьюверов: %v", customerrors.ErrDBQuery, err)
		}
		created.ReviewerIDs = pr.ReviewerIDs
	} else {
		created.ReviewerIDs = []int{}
	}

	return created, nil
}

// MergePullRequest идемпотентно помечает PR как MERGED
func (s *prService) MergePullRequest(ctx context.Context, prID string) (*models.PullRequest, error) {
	pr, err := s.prRepo.SetPullRequestMerged(ctx, prID)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			return nil, fmt.Errorf("%w: pr_id=%s", customerrors.ErrNotFound, prID)
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	return pr, nil
}

// ReassignReviewer заменяет одного ревьювера на другого из его команды
func (s *prService) ReassignReviewer(ctx context.Context, prID string, oldReviewerID int) (*models.PullRequest, error) {
	pr, err := s.prRepo.GetPullRequestByPublicID(ctx, prID)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			return nil, fmt.Errorf("%w: pr_id=%s", customerrors.ErrNotFound, prID)
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	if pr.Status == models.PRStatusMerged {
		return nil, fmt.Errorf("%w: нельзя менять ревьюверов после MERGED", customerrors.ErrForbidden)
	}

	found := false
	for _, r := range pr.ReviewerIDs {
		if r == oldReviewerID {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("%w: ревьювер %d не назначен на PR", customerrors.ErrNotFound, oldReviewerID)
	}

	// Получаем команду пользователя по его внутреннему ID
	team, err := s.teamRepo.GetTeamOfUser(ctx, oldReviewerID)
	if err != nil {
		if errors.Is(err, customerrors.ErrNotFound) {
			return nil, fmt.Errorf("%w: команда ревьювера не найдена", customerrors.ErrNotFound)
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	// Исключаем автора и других ревьюверов (но не самого заменяемого ревьювера)
	exclude := []int{pr.AuthorID}
	for _, reviewerID := range pr.ReviewerIDs {
		if reviewerID != oldReviewerID {
			exclude = append(exclude, reviewerID)
		}
	}

	candidates := []int{}
	for _, u := range team.Members {
		if u.IsActive && !contains(exclude, u.ID) {
			candidates = append(candidates, u.ID)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("%w: нет доступных кандидатов для замены ревьювера %d", customerrors.ErrForbidden, oldReviewerID)
	}

	newReviewerID := candidates[s.rnd.Intn(len(candidates))]

	for i, r := range pr.ReviewerIDs {
		if r == oldReviewerID {
			pr.ReviewerIDs[i] = newReviewerID
			break
		}
	}

	if err := s.prRepo.UpdateReviewers(ctx, pr.ID, pr.ReviewerIDs); err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	return pr, nil
}

// GetPRsForReviewer список PR где юзер ревьювер
func (s *prService) GetPRsForReviewer(ctx context.Context, userID int) ([]models.PullRequestShort, error) {
	prs, err := s.prRepo.GetPRsWhereUserReviewer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}
	return prs, nil
}

func chooseRandomUpToN(r *rand.Rand, candidates []int, n int) []int {
	l := len(candidates)
	if l == 0 || n <= 0 {
		return []int{}
	}
	if l <= n {
		res := make([]int, l)
		copy(res, candidates)
		r.Shuffle(len(res), func(i, j int) { res[i], res[j] = res[j], res[i] })
		return res
	}
	idxs := make([]int, l)
	copy(idxs, candidates)
	r.Shuffle(l, func(i, j int) { idxs[i], idxs[j] = idxs[j], idxs[i] })
	return idxs[:n]
}

func contains(arr []int, val int) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}
