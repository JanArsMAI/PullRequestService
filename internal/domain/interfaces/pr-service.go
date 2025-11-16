package interfaces

import (
	"context"

	entityPr "github.com/JanArsMAI/PullRequestService/internal/domain/pullrequest"
	entityTeam "github.com/JanArsMAI/PullRequestService/internal/domain/team"
	entityUser "github.com/JanArsMAI/PullRequestService/internal/domain/user"
	"github.com/JanArsMAI/PullRequestService/internal/presentation/gin/dto"
)

type PrService interface {
	AddTeam(ctx context.Context, teamDto *dto.AddTeamRequest) error
	ReassignPullRequest(ctx context.Context, activePr entityPr.PullRequest, user entityUser.User) error
	GetTeam(ctx context.Context, teamName string) (*entityTeam.Team, error)
	SetUserActive(ctx context.Context, userID string, isActive bool) error
	GetUserWithTeam(ctx context.Context, userID string) (*entityUser.User, string, error)
	CreatePR(ctx context.Context, prDto dto.CreatePR) (*entityPr.PullRequest, error)
	GetUsersPr(ctx context.Context, userID string) ([]entityPr.PullRequest, error)
	Merge(ctx context.Context, userID string, prId string) (*entityPr.PullRequest, error)
	Reassign(ctx context.Context, prID, oldReviewerID string) (*entityPr.PullRequest, string, error)
	GetStatistics(ctx context.Context) (map[string]int, map[string]int, error)
	Deactivate(ctx context.Context, teamName string, userIDs []string) error
}
