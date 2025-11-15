package application_test

import (
	"context"
	"testing"

	"github.com/JanArsMAI/PullRequestService/internal/application"
	mock_interfaces "github.com/JanArsMAI/PullRequestService/internal/domain/interfaces/mocks"
	entityUser "github.com/JanArsMAI/PullRequestService/internal/domain/user"
	"github.com/JanArsMAI/PullRequestService/internal/infrastructure/repos"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/golang/mock/gomock"
)

func TestPrService_SetUserActive_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	mockRepo.EXPECT().
		GetUserByID(gomock.Any(), "user1").
		Return(nil, repos.ErrNoUserWithId)

	err := svc.SetUserActive(context.Background(), "user1", true)
	assert.ErrorIs(t, err, application.ErrUserNotFound)
}
func TestPrService_SetUserActive_AlreadyActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	user := &entityUser.User{Id: "user1", IsActive: true}

	mockRepo.EXPECT().
		GetUserByID(gomock.Any(), "user1").
		Return(user, nil)

	err := svc.SetUserActive(context.Background(), "user1", true)
	assert.NoError(t, err)
}
