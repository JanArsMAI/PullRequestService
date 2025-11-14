package entity

import (
	"time"

	entity "github.com/JanArsMAI/PullRequestService/internal/domain/user"
)

type PullRequest struct {
	Id                string
	Name              string
	Author            entity.User
	Reviewers         []entity.User
	Status            string
	NeedMoreReviewers bool
	CreatedAt         time.Time
	MergedAt          *time.Time
}
