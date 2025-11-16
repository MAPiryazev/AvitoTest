package models

import "time"

// User
type User struct {
	ID       int    `json:"id"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamID   int    `json:"team_id"`
	IsActive bool   `json:"is_active"`
}

type UserWithTeam struct {
	ID       int    `json:"id"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamID   int    `json:"team_id"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

// Team
type Team struct {
	ID      int    `json:"id"`
	Name    string `json:"team_name"`
	Members []User `json:"members,omitempty"`
}

// Pull Request
type PullRequestStatus string

const (
	PRStatusOpen   PullRequestStatus = "OPEN"
	PRStatusMerged PullRequestStatus = "MERGED"
)

type PullRequest struct {
	ID            int               `json:"id"`
	PullRequestID string            `json:"pull_request_id"`
	Name          string            `json:"pull_request_name"`
	AuthorID      int               `json:"author_id"`
	Status        PullRequestStatus `json:"status"`
	CreatedAt     time.Time         `json:"created_at"`
	MergedAt      *time.Time        `json:"merged_at,omitempty"`

	ReviewerIDs []int `json:"reviewer_ids"`
}

// Упрощённая модель
type PullRequestShort struct {
	ID            int               `json:"id"`
	PullRequestID string            `json:"pull_request_id"`
	Name          string            `json:"pull_request_name"`
	AuthorID      int               `json:"author_id"`
	Status        PullRequestStatus `json:"status"`
}
