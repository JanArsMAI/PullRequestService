package application

import (
	"context"
	"errors"
	"fmt"
	"sync"

	entityPR "github.com/JanArsMAI/PullRequestService/internal/domain/pullrequest"
	entityUser "github.com/JanArsMAI/PullRequestService/internal/domain/user"
	"github.com/JanArsMAI/PullRequestService/internal/infrastructure/repos"
	"github.com/JanArsMAI/PullRequestService/internal/presentation/gin/dto"
)

var (
	ErrTeamWithNameAlreadyCreated = errors.New("error. Team with this name is created")
)

type PrService struct {
	repo *repos.PostgresRepo
}

func NewPrService(repo *repos.PostgresRepo) *PrService {
	return &PrService{
		repo: repo,
	}
}

func (s *PrService) AddTeam(ctx context.Context, teamDto *dto.AddTeamRequest) error {
	team, _ := s.repo.GetTeamByName(ctx, teamDto.TeamName)
	if team != nil {
		return ErrTeamWithNameAlreadyCreated
	}
	users := make([]entityUser.User, 0, len(teamDto.Members))
	errChan := make(chan error, len(teamDto.Members))
	for _, userDto := range teamDto.Members {
		userDto := userDto
		user, err := s.repo.GetUserByID(ctx, userDto.Id)
		if err != nil {
			if err == repos.ErrNoUserWithId {
				users = append(users, entityUser.User{
					Id:       userDto.Id,
					Name:     userDto.Name,
					IsActive: userDto.IsActive,
				})
				continue
			}
			return fmt.Errorf("failed to get user: %w", err)
		}
		prs, err := s.repo.GetUsersPr(ctx, user.Id, true)
		if err != nil {
			return fmt.Errorf("failed to get user's PRs: %w", err)
		}
		var wg sync.WaitGroup
		for _, pr := range prs {
			wg.Add(1)
			pr := pr
			go func() {
				defer wg.Done()
				if err := s.ReassignPullRequest(ctx, pr, *user); err != nil {
					errChan <- fmt.Errorf("reassign PR %s for user %s: %w", pr.Id, user.Id, err)
				}
			}()
		}
		wg.Wait()

		if len(prs) > 0 {
			if err := s.repo.RemoveReviewerFromAllPR(ctx, user.Id); err != nil {
				return fmt.Errorf("failed to remove old reviewer %s from all PRs: %w", user.Id, err)
			}
		}
		users = append(users, *user)
	}
	close(errChan)
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	if err := s.repo.AddTeam(ctx, teamDto.TeamName, users); err != nil {
		return fmt.Errorf("failed to add team: %w", err)
	}

	return nil
}

func (s *PrService) ReassignPullRequest(ctx context.Context, activePr entityPR.PullRequest, user entityUser.User) error {
	team, err := s.repo.GetTeam(ctx, user.TeamID)
	if err != nil {
		return fmt.Errorf("failed to get team for user %s: %w", user.Id, err)
	}
	var newReviewer *entityUser.User
	for _, candidate := range team.Users {
		if candidate.Id != user.Id && candidate.Id != activePr.Author.Id && candidate.IsActive {
			newReviewer = &candidate
			break
		}
	}

	if newReviewer == nil {
		activePr.NeedMoreReviewers = true
	} else {
		for i, reviewer := range activePr.Reviewers {
			if reviewer.Id == user.Id {
				activePr.Reviewers[i] = *newReviewer
				break
			}
		}
	}
	if err := s.repo.UpdatePr(ctx, activePr.Id, activePr); err != nil {
		return fmt.Errorf("failed to update PR: %w", err)
	}

	return nil
}
