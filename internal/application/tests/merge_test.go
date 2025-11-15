package application_test

import (
	"context"
	"testing"

	"github.com/JanArsMAI/PullRequestService/internal/application"
	mock_interfaces "github.com/JanArsMAI/PullRequestService/internal/domain/interfaces/mocks"
	entityPR "github.com/JanArsMAI/PullRequestService/internal/domain/pullrequest"
	entityUser "github.com/JanArsMAI/PullRequestService/internal/domain/user"
	"github.com/JanArsMAI/PullRequestService/internal/infrastructure/repos"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/golang/mock/gomock"
)

func TestPrService_Merge_PrNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	mockRepo.EXPECT().
		GetPr(gomock.Any(), "pr1").
		Return(nil, repos.ErrPrNotFound)

	pr, err := svc.Merge(context.Background(), "user1", "pr1")
	assert.Nil(t, pr)
	assert.ErrorIs(t, err, application.ErrPrNotFound)
}

func TestPrService_Merge_NotReviewer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	prObj := &entityPR.PullRequest{
		Id:        "pr1",
		Reviewers: []entityUser.User{{Id: "user2"}},
		Status:    "OPEN",
	}

	mockRepo.EXPECT().
		GetPr(gomock.Any(), "pr1").
		Return(prObj, nil)

	pr, err := svc.Merge(context.Background(), "user1", "pr1")
	assert.Nil(t, pr)
	assert.ErrorIs(t, err, application.ErrUnableToMerge)
}

func TestPrService_Merge_PRAlreadyClosed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	prObj := &entityPR.PullRequest{
		Id:        "pr1",
		Reviewers: []entityUser.User{{Id: "user1"}},
		Status:    "MERGED",
	}

	mockRepo.EXPECT().
		GetPr(gomock.Any(), "pr1").
		Return(prObj, nil)

	pr, err := svc.Merge(context.Background(), "user1", "pr1")
	assert.Nil(t, pr)
	assert.ErrorIs(t, err, application.ErrUnableToMerge)
}

func TestPrService_Merge_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	prObj := &entityPR.PullRequest{
		Id:        "pr1",
		Reviewers: []entityUser.User{{Id: "user1"}},
		Status:    "OPEN",
	}

	mockRepo.EXPECT().
		GetPr(gomock.Any(), "pr1").
		Return(prObj, nil)
	mockRepo.EXPECT().
		UpdatePr(gomock.Any(), "pr1", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, p entityPR.PullRequest) error {
			assert.Equal(t, "MERGED", p.Status)
			assert.NotNil(t, p.MergedAt)
			return nil
		})

	pr, err := svc.Merge(context.Background(), "user1", "pr1")
	assert.NoError(t, err)
	assert.Equal(t, "MERGED", pr.Status)
	assert.NotNil(t, pr.MergedAt)
}

func TestPrService_Merge_AdminCanMerge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	prObj := &entityPR.PullRequest{
		Id:        "pr1",
		Reviewers: []entityUser.User{},
		Status:    "OPEN",
	}

	mockRepo.EXPECT().
		GetPr(gomock.Any(), "pr1").
		Return(prObj, nil)
	mockRepo.EXPECT().
		UpdatePr(gomock.Any(), "pr1", gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, p entityPR.PullRequest) error {
			assert.Equal(t, "MERGED", p.Status)
			assert.NotNil(t, p.MergedAt)
			return nil
		})

	pr, err := svc.Merge(context.Background(), "admin", "pr1")
	assert.NoError(t, err)
	assert.Equal(t, "MERGED", pr.Status)
	assert.NotNil(t, pr.MergedAt)
}
