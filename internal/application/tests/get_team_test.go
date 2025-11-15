package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/JanArsMAI/PullRequestService/internal/application"
	mock_interfaces "github.com/JanArsMAI/PullRequestService/internal/domain/interfaces/mocks"
	entityTeam "github.com/JanArsMAI/PullRequestService/internal/domain/team"
	"github.com/JanArsMAI/PullRequestService/internal/infrastructure/repos"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/golang/mock/gomock"
)

func TestPrService_GetTeam_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	teamName := "team1"
	expectedTeam := &entityTeam.Team{
		Id:   1,
		Name: teamName,
	}

	mockRepo.EXPECT().
		GetTeamByName(gomock.Any(), teamName).
		Return(expectedTeam, nil)

	team, err := svc.GetTeam(context.Background(), teamName)
	assert.NoError(t, err)
	assert.Equal(t, expectedTeam, team)
}

func TestPrService_GetTeam_TeamNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	teamName := "team1"

	mockRepo.EXPECT().
		GetTeamByName(gomock.Any(), teamName).
		Return(nil, repos.ErrTeamNotFound)

	team, err := svc.GetTeam(context.Background(), teamName)
	assert.Nil(t, team)
	assert.ErrorIs(t, err, application.ErrTeamNotFound)
}

func TestPrService_GetTeam_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_interfaces.NewMockPullRequestRepo(ctrl)
	svc := application.NewPrService(mockRepo)

	teamName := "team1"

	mockRepo.EXPECT().
		GetTeamByName(gomock.Any(), teamName).
		Return(nil, errors.New("db error"))

	team, err := svc.GetTeam(context.Background(), teamName)
	assert.Nil(t, team)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
