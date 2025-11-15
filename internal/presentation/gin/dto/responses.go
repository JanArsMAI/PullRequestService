package dto

import "time"

type TeamResponse struct {
	Team TeamDtoResponse `json:"team"`
}

type TeamDtoResponse struct {
	TeamName string              `json:"team_name"`
	Members  []MemberDtoResponse `json:"members"`
}

type ErrorResponse struct {
	Error ErrorMessage `json:"error"`
}

type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type MemberDtoResponse struct {
	Id       string `json:"user_id"`
	Name     string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type UserResponse struct {
	User UserWithTeam `json:"user"`
}

type UserWithTeam struct {
	Id       string `json:"user_id"`
	Name     string `json:"username"`
	IsActive bool   `json:"is_active"`
	Team     string `json:"team_name"`
}

type PullRequestResponse struct {
	Pr PullRequest `json:"pr"`
}

type PullRequest struct {
	Id        string   `json:"pull_request_id"`
	Name      string   `json:"pull_request_name"`
	AuthorId  string   `json:"author_id"`
	Status    string   `json:"status"`
	Reviewers []string `json:"assigned_reviewers"`
}

type UsersPrResponse struct {
	UserId string              `json:"user_id"`
	Prs    []PullRequestOfUser `json:"pull_requests"`
}

type PullRequestOfUser struct {
	Id       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorId string `json:"author_id"`
	Status   string `json:"status"`
}

//TODO: после заверщения и тестирования бахнуть рефакторинг responses, чтобы было omitempty и универсальность

type MergeResponse struct {
	Pr MergedPullRequestOfUser `json:"pr"`
}

type MergedPullRequestOfUser struct {
	Id        string     `json:"pull_request_id"`
	Name      string     `json:"pull_request_name"`
	AuthorId  string     `json:"author_id"`
	Status    string     `json:"status"`
	Reviewers []string   `json:"assigned_reviewers"`
	MergeAt   *time.Time `json:"merged_at,omitempty"`
}

type ReassignResponse struct {
	Pr         ReassignPR `json:"pr"`
	ReplacedBy string     `json:"replaced_by"`
}

type ReassignPR struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
}
