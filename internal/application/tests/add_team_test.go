package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JanArsMAI/PullRequestService/internal/application"
	mock_interfaces "github.com/JanArsMAI/PullRequestService/internal/domain/interfaces/mocks"
	entityTeam "github.com/JanArsMAI/PullRequestService/internal/domain/team"
	entityUser "github.com/JanArsMAI/PullRequestService/internal/domain/user"
	"github.com/JanArsMAI/PullRequestService/internal/infrastructure/repos"
	"github.com/JanArsMAI/PullRequestService/internal/presentation/gin/dto"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/golang/mock/gomock"
)

func TestPrService_AddTeam_TeamAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	teamDto := &dto.AddTeamRequest{
		TeamName: "team1",
		Members:  []dto.MemberDto{},
	}
	mockRepo.EXPECT().
		GetTeamByName(gomock.Any(), "team1").
		Return(&entityTeam.Team{}, nil)

	err := svc.AddTeam(context.Background(), teamDto)
	assert.ErrorIs(t, err, application.ErrTeamWithNameAlreadyCreated)
}

func TestPrService_AddTeam_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	teamDto := &dto.AddTeamRequest{
		TeamName: "team1",
		Members: []dto.MemberDto{
			{Id: "user1", Name: "Alice", IsActive: true},
		},
	}
	mockRepo.EXPECT().
		GetTeamByName(gomock.Any(), "team1").
		Return(nil, nil)
	mockRepo.EXPECT().
		GetUserByID(gomock.Any(), "user1").
		Return(nil, repos.ErrNoUserWithId)
	mockRepo.EXPECT().
		AddTeam(gomock.Any(), "team1", gomock.AssignableToTypeOf([]entityUser.User{})).
		Return(nil)

	err := svc.AddTeam(context.Background(), teamDto)
	assert.NoError(t, err)
}

func TestPrService_AddTeam_ErrorGettingUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	teamDto := &dto.AddTeamRequest{
		TeamName: "team1",
		Members: []dto.MemberDto{
			{Id: "user1", Name: "Alice", IsActive: true},
		},
	}
	mockRepo.EXPECT().
		GetTeamByName(gomock.Any(), "team1").
		Return(nil, nil)
	mockRepo.EXPECT().
		GetUserByID(gomock.Any(), "user1").
		Return(nil, errors.New("db error"))

	err := svc.AddTeam(context.Background(), teamDto)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user")
}

func TestPrService_AddTeam_ErrorGettingUsersPr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	teamDto := &dto.AddTeamRequest{
		TeamName: "team1",
		Members: []dto.MemberDto{
			{Id: "user1", Name: "Alice", IsActive: true},
		},
	}
	mockRepo.EXPECT().
		GetTeamByName(gomock.Any(), "team1").
		Return(nil, nil)
	mockRepo.EXPECT().
		GetUserByID(gomock.Any(), "user1").
		Return(&entityUser.User{Id: "user1", Name: "Alice"}, nil)
	mockRepo.EXPECT().
		GetUsersPr(gomock.Any(), "user1", true).
		Return(nil, errors.New("db error"))

	err := svc.AddTeam(context.Background(), teamDto)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user's PRs")
}
