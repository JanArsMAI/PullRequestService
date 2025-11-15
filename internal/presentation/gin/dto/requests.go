package dto

type AddTeamRequest struct {
	TeamName string      `json:"team_name"`
	Members  []MemberDto `json:"members"`
}

type MemberDto struct {
	Id       string `json:"user_id"`
	Name     string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type CreatePR struct {
	PrID     string `json:"pull_request_id"`
	PrName   string `json:"pull_request_name"`
	PrAuthor string `json:"author_id"`
}

type SetUserActive struct {
	UserId   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type CreatePr struct {
	Id     string `json:"pull_request_id"`
	Name   string `json:"pull_request_name"`
	Author string `json:"author_id"`
}

type MergeRequest struct {
	Id string `json:"pull_request_id"`
}

type ReassignPullRequest struct {
	PrID        string `json:"pull_request_id"`
	OldReviewer string `json:"old_reviewer_id"`
}
