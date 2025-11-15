package adapter

import (
	"AvitoTest/internal/api"
	"AvitoTest/internal/models"
	"strconv"
)

// UserModelToAPI конвертирует обычного User.
// Здесь TeamName пустой, потому что обычный User его не содержит.
func UserModelToAPI(u models.User) api.User {
	return api.User{
		UserId:   u.UserID,
		Username: u.Username,
		TeamName: "",
		IsActive: u.IsActive,
	}
}

// UserWithTeamModelToAPI — корректная конвертация с названием команды.
func UserWithTeamModelToAPI(u models.UserWithTeam) api.User {
	return api.User{
		UserId:   u.UserID,
		Username: u.Username,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}
}

// Пользователи → TeamMember (version 1 — только User)
func UsersModelToAPI(users []models.User) []api.TeamMember {
	out := make([]api.TeamMember, len(users))
	for i, u := range users {
		out[i] = api.TeamMember{
			UserId:   u.UserID,
			Username: u.Username,
			IsActive: u.IsActive,
		}
	}
	return out
}

// Пользователи → TeamMember (version 2 — UserWithTeam)
func UsersWithTeamModelToAPI(users []models.UserWithTeam) []api.TeamMember {
	out := make([]api.TeamMember, len(users))
	for i, u := range users {
		out[i] = api.TeamMember{
			UserId:   u.UserID,
			Username: u.Username,
			IsActive: u.IsActive,
		}
	}
	return out
}

// TeamModelToAPI — если у тебя обычная Team и участники — обычные User
func TeamModelToAPI(team models.Team, members []models.User) api.Team {
	return api.Team{
		TeamName: team.Name,
		Members:  UsersModelToAPI(members),
	}
}

// TeamWithMembersToAPI — если метод возвращает []UserWithTeam
func TeamWithMembersToAPI(teamName string, members []models.UserWithTeam) api.Team {
	return api.Team{
		TeamName: teamName,
		Members:  UsersWithTeamModelToAPI(members),
	}
}

// PullRequestModelToAPIWithUserIDs конвертирует PR с получением user_id для ревьюверов и автора
func PullRequestModelToAPIWithUserIDs(pr models.PullRequest, authorUserID string, reviewerUserIDs []string) api.PullRequest {
	return api.PullRequest{
		PullRequestId:     pr.PullRequestID,
		PullRequestName:   pr.Name,
		AuthorId:          authorUserID,
		AssignedReviewers: reviewerUserIDs,
		Status:            api.PullRequestStatus(pr.Status),
		CreatedAt:         &pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

// PullRequestModelToAPI - устаревший метод, используйте PullRequestModelToAPIWithUserIDs
// Оставлен для обратной совместимости, но возвращает внутренние ID вместо user_id
func PullRequestModelToAPI(pr models.PullRequest) api.PullRequest {
	assigned := make([]string, len(pr.ReviewerIDs))
	for i, id := range pr.ReviewerIDs {
		assigned[i] = strconv.Itoa(id)
	}

	return api.PullRequest{
		PullRequestId:     pr.PullRequestID,
		PullRequestName:   pr.Name,
		AuthorId:          strconv.Itoa(pr.AuthorID),
		AssignedReviewers: assigned,
		Status:            api.PullRequestStatus(pr.Status),
		CreatedAt:         &pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

func PullRequestShortModelToAPI(pr models.PullRequestShort) api.PullRequestShort {
	return api.PullRequestShort{
		PullRequestId:   pr.PullRequestID,
		PullRequestName: pr.Name,
		AuthorId:        strconv.Itoa(pr.AuthorID),
		Status:          api.PullRequestShortStatus(pr.Status),
	}
}
