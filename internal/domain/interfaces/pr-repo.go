package interfaces

import (
	"context"

	entityPr "github.com/JanArsMAI/PullRequestService/internal/domain/pullrequest"
	entityTeam "github.com/JanArsMAI/PullRequestService/internal/domain/team"
	entityUser "github.com/JanArsMAI/PullRequestService/internal/domain/user"
)

type PullRequestRepo interface {
	AddTeam(ctx context.Context, name string, users []entityUser.User) error
	GetTeam(ctx context.Context, id int) (*entityTeam.Team, error)
	GetTeamByName(ctx context.Context, name string) (*entityTeam.Team, error)
	AddPR(ctx context.Context, pr entityPr.PullRequest) error
	GetPr(ctx context.Context, prID string) (*entityPr.PullRequest, error)
	UpdatePr(ctx context.Context, prId string, newPr entityPr.PullRequest) error
	GetUsersPr(ctx context.Context, userId string, onlyActive bool) ([]entityPr.PullRequest, error)
	GetUserByID(ctx context.Context, userID string) (*entityUser.User, error)
	RemoveReviewerFromAllPR(ctx context.Context, reviewerID string) error
	UpdateUser(ctx context.Context, u entityUser.User) error
	GetUserWithTeam(ctx context.Context, userID string) (*entityUser.User, string, error)
	AddReviewerToPR(ctx context.Context, prId string, reviewerID string) error
	GetTeamPr(ctx context.Context, teamID int) ([]entityPr.PullRequest, error)
}
