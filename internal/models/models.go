package models

type User struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    TeamName string `json:"team_name"`
    IsActive bool   `json:"is_active"`
}

type Team struct {
    TeamName string `json:"team_name"`
    Members  []User `json:"members"`
}

type PullRequest struct {
    PullRequestID    string   `json:"pull_request_id"`
    PullRequestName  string   `json:"pull_request_name"`
    AuthorID         string   `json:"author_id"`
    Status           string   `json:"status"`
    AssignedReviewers []string `json:"assigned_reviewers"`
    CreatedAt        *string  `json:"createdAt,omitempty"`
    MergedAt         *string  `json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
    PullRequestID   string `json:"pull_request_id"`
    PullRequestName string `json:"pull_request_name"`
    AuthorID        string `json:"author_id"`
    Status          string `json:"status"`
}