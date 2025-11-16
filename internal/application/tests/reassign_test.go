package application_test

import (
	"context"
	"testing"

	"github.com/JanArsMAI/PullRequestService/internal/application"
	mock_interfaces "github.com/JanArsMAI/PullRequestService/internal/domain/interfaces/mocks"
	entityPR "github.com/JanArsMAI/PullRequestService/internal/domain/pullrequest"
	entityTeam "github.com/JanArsMAI/PullRequestService/internal/domain/team"
	entityUser "github.com/JanArsMAI/PullRequestService/internal/domain/user"
	"github.com/JanArsMAI/PullRequestService/internal/infrastructure/repos"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/golang/mock/gomock"
)

func TestPrService_Reassign_PrNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	mockRepo.EXPECT().GetPr(gomock.Any(), "pr1").Return(nil, repos.ErrPrNotFound)

	pr, newID, err := svc.Reassign(context.Background(), "pr1", "user1")
	assert.Nil(t, pr)
	assert.Empty(t, newID)
	assert.ErrorIs(t, err, application.ErrPrNotFound)
}

func TestPrService_Reassign_PrMerged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	prObj := &entityPR.PullRequest{
		Id:        "pr1",
		Status:    "MERGED",
		Reviewers: []entityUser.User{{Id: "user1"}},
		Author:    entityUser.User{Id: "author1"},
	}

	mockRepo.EXPECT().GetPr(gomock.Any(), "pr1").Return(prObj, nil)

	pr, newID, err := svc.Reassign(context.Background(), "pr1", "user1")
	assert.Nil(t, pr)
	assert.Empty(t, newID)
	assert.ErrorIs(t, err, application.ErrPrIsMerged)
}

func TestPrService_Reassign_NotAssigned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	prObj := &entityPR.PullRequest{
		Id:        "pr1",
		Status:    "OPEN",
		Reviewers: []entityUser.User{{Id: "user2"}},
		Author:    entityUser.User{Id: "author1"},
	}

	mockRepo.EXPECT().GetPr(gomock.Any(), "pr1").Return(prObj, nil)

	pr, newID, err := svc.Reassign(context.Background(), "pr1", "user1")
	assert.Nil(t, pr)
	assert.Empty(t, newID)
	assert.ErrorIs(t, err, application.ErrNotAssigned)
}
func TestPrService_Reassign_NoCandidates(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	prObj := &entityPR.PullRequest{
		Id:        "pr1",
		Status:    "OPEN",
		Reviewers: []entityUser.User{{Id: "user1"}},
		Author:    entityUser.User{Id: "author1", TeamID: 1},
	}

	team := &entityTeam.Team{
		Id:    1,
		Users: []entityUser.User{{Id: "user1", IsActive: true}},
	}

	mockRepo.EXPECT().GetPr(gomock.Any(), "pr1").Return(prObj, nil)
	mockRepo.EXPECT().GetUserByID(gomock.Any(), "author1").Return(&prObj.Author, nil)
	mockRepo.EXPECT().GetTeam(gomock.Any(), 1).Return(team, nil)

	pr, newID, err := svc.Reassign(context.Background(), "pr1", "user1")
	assert.Nil(t, pr)
	assert.Empty(t, newID)
	assert.ErrorIs(t, err, application.ErrNoCandidate)
}

func TestPrService_Reassign_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	author := entityUser.User{Id: "author1", TeamID: 1}
	oldReviewer := entityUser.User{Id: "user1"}
	newReviewer := entityUser.User{Id: "user2", IsActive: true}

	prObj := &entityPR.PullRequest{
		Id:        "pr1",
		Status:    "OPEN",
		Reviewers: []entityUser.User{oldReviewer},
		Author:    author,
	}

	team := &entityTeam.Team{
		Id:    1,
		Users: []entityUser.User{oldReviewer, newReviewer},
	}

	mockRepo.EXPECT().GetPr(gomock.Any(), "pr1").Return(prObj, nil)
	mockRepo.EXPECT().GetUserByID(gomock.Any(), "author1").Return(&author, nil)
	mockRepo.EXPECT().GetTeam(gomock.Any(), 1).Return(team, nil)
	mockRepo.EXPECT().UpdatePr(gomock.Any(), "pr1", gomock.Any()).DoAndReturn(
		func(_ context.Context, _ string, p entityPR.PullRequest) error {
			assert.Len(t, p.Reviewers, 1)
			assert.Equal(t, newReviewer.Id, p.Reviewers[0].Id)
			return nil
		},
	)

	pr, newID, err := svc.Reassign(context.Background(), "pr1", oldReviewer.Id)
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, newReviewer.Id, newID)
}
