package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JanArsMAI/PullRequestService/internal/application"
	mock_interfaces "github.com/JanArsMAI/PullRequestService/internal/domain/interfaces/mocks"
	entityPR "github.com/JanArsMAI/PullRequestService/internal/domain/pullrequest"
	"github.com/JanArsMAI/PullRequestService/internal/presentation/gin/dto"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/golang/mock/gomock"
)

func TestPrService_CreatePR_PRAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	prDto := dto.CreatePR{PrID: "pr1", PrAuthor: "user1", PrName: "MyPR"}

	mockRepo.EXPECT().GetPr(gomock.Any(), "pr1").Return(&entityPR.PullRequest{}, nil)

	pr, err := svc.CreatePR(context.Background(), prDto)
	assert.Nil(t, pr)
	assert.ErrorIs(t, err, application.ErrPrIsAlreadyCreated)
}

func TestPrService_CreatePR_ErrorCheckingPR(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	prDto := dto.CreatePR{PrID: "pr1", PrAuthor: "user1", PrName: "MyPR"}

	mockRepo.EXPECT().GetPr(gomock.Any(), "pr1").Return(nil, errors.New("db error"))

	pr, err := svc.CreatePR(context.Background(), prDto)
	assert.Nil(t, pr)
	assert.ErrorContains(t, err, "failed to check existing PR")
}
