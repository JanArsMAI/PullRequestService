package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JanArsMAI/PullRequestService/internal/application"
	mock_interfaces "github.com/JanArsMAI/PullRequestService/internal/domain/interfaces/mocks"
	entityPR "github.com/JanArsMAI/PullRequestService/internal/domain/pullrequest"
	entityUser "github.com/JanArsMAI/PullRequestService/internal/domain/user"
	"github.com/JanArsMAI/PullRequestService/internal/infrastructure/repos"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/golang/mock/gomock"
)

func TestPrService_GetUsersPr_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	mockRepo.EXPECT().
		GetUserByID(gomock.Any(), "user1").
		Return(nil, repos.ErrNoUserWithId)

	prs, err := svc.GetUsersPr(context.Background(), "user1")
	assert.Nil(t, prs)
	assert.ErrorIs(t, err, application.ErrUserNotFound)
}

func TestPrService_GetUsersPr_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	user := &entityUser.User{Id: "user1", IsActive: true}
	mockRepo.EXPECT().
		GetUserByID(gomock.Any(), "user1").
		Return(user, nil)
	mockRepo.EXPECT().
		GetUsersPr(gomock.Any(), "user1", false).
		Return(nil, errors.New("db error"))

	prs, err := svc.GetUsersPr(context.Background(), "user1")
	assert.Nil(t, prs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestPrService_GetUsersPr_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	user := &entityUser.User{Id: "user1", IsActive: true}
	prList := []entityPR.PullRequest{
		{Id: "pr1", Author: *user},
		{Id: "pr2", Author: *user},
	}

	mockRepo.EXPECT().
		GetUserByID(gomock.Any(), "user1").
		Return(user, nil)
	mockRepo.EXPECT().
		GetUsersPr(gomock.Any(), "user1", false).
		Return(prList, nil)

	prs, err := svc.GetUsersPr(context.Background(), "user1")
	assert.NoError(t, err)
	assert.Equal(t, prList, prs)
}
