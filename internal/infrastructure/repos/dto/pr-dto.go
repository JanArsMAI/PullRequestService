package dto

import "time"

type PullRequestDto struct {
	ID                string     `db:"pull_request_id"`
	Name              string     `db:"pull_request_name"`
	AuthorID          string     `db:"author_id"`
	Status            string     `db:"status"`
	NeedMoreReviewers bool       `db:"need_more_reviewers"`
	CreatedAt         time.Time  `db:"created_at"`
	MergedAt          *time.Time `db:"merged_at"`
}

type PullRequestReviewerDto struct {
	PullRequestID string    `db:"pull_request_id"`
	ReviewerID    string    `db:"reviewer_id"`
	AssignedAt    time.Time `db:"assigned_at"`
}
