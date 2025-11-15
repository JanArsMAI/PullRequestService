package application

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/JanArsMAI/PullRequestService/internal/domain/interfaces"
	entity "github.com/JanArsMAI/PullRequestService/internal/domain/pullrequest"
	entityPR "github.com/JanArsMAI/PullRequestService/internal/domain/pullrequest"
	entityTeam "github.com/JanArsMAI/PullRequestService/internal/domain/team"
	entityUser "github.com/JanArsMAI/PullRequestService/internal/domain/user"
	"github.com/JanArsMAI/PullRequestService/internal/infrastructure/repos"
	"github.com/JanArsMAI/PullRequestService/internal/presentation/gin/dto"
)

var (
	ErrTeamWithNameAlreadyCreated = errors.New("error. Team with this name is created")
	ErrTeamNotFound               = errors.New("error. Team with this name is not found")
	ErrUserNotFound               = errors.New("error. User with this id is not found")
	ErrPrIsAlreadyCreated         = errors.New("error. Pull request with this id is already created")
	ErrAuthorOrTeamAreNotFound    = errors.New("error. Author or Team of Pull Request not exist")
	ErrPrNotFound                 = errors.New("PR with this ID is not found")
	ErrUnableToMerge              = errors.New("Unable to merge")
	ErrPrIsMerged                 = errors.New("PR is already merged")
	ErrNoCandidate                = errors.New("no candidate to reassign")
	ErrNotAssigned                = errors.New("no user with this id assigned to PR")
)

type PrService struct {
	repo interfaces.PullRequestRepo
}

func NewPrService(repo interfaces.PullRequestRepo) interfaces.PrService {
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
		user.Name = userDto.Name
		user.IsActive = userDto.IsActive

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
		if err == repos.ErrTeamNotFound {
			return ErrTeamNotFound
		}
		return fmt.Errorf("failed to get team for user %s: %w", user.Id, err)
	}
	filtered := make([]entityUser.User, 0, len(activePr.Reviewers))
	for _, r := range activePr.Reviewers {
		if r.Id != user.Id {
			filtered = append(filtered, r)
		}
	}
	activePr.Reviewers = filtered
	activeCount := 0
	for _, r := range activePr.Reviewers {
		for _, tUser := range team.Users {
			if tUser.Id == r.Id && tUser.IsActive {
				activeCount++
				break
			}
		}
	}
	if activeCount < 2 {
		activePr.NeedMoreReviewers = true
	}
	candidates := make([]entityUser.User, 0)
	for _, candidate := range team.Users {
		if candidate.Id != user.Id &&
			candidate.Id != activePr.Author.Id &&
			candidate.IsActive {
			candidates = append(candidates, candidate)
		}
	}
	if len(candidates) > 0 {
		rand.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })
		activePr.Reviewers = append(activePr.Reviewers, candidates[0])
		activeCount++
		if activeCount >= 2 {
			activePr.NeedMoreReviewers = false
		}
	}
	seen := make(map[string]struct{})
	unique := make([]entityUser.User, 0)
	for _, r := range activePr.Reviewers {
		if _, ok := seen[r.Id]; !ok {
			seen[r.Id] = struct{}{}
			unique = append(unique, r)
		}
	}
	activePr.Reviewers = unique

	return s.repo.UpdatePr(ctx, activePr.Id, activePr)
}

func (s *PrService) GetTeam(ctx context.Context, teamName string) (*entityTeam.Team, error) {
	team, err := s.repo.GetTeamByName(ctx, teamName)
	if err != nil {
		if errors.Is(err, repos.ErrTeamNotFound) {
			return nil, ErrTeamNotFound
		}
		return nil, err
	}
	return team, nil
}

func (s *PrService) SetUserActive(ctx context.Context, userId string, isActive bool) error {
	user, err := s.repo.GetUserByID(ctx, userId)
	if err != nil {
		if errors.Is(err, repos.ErrNoUserWithId) {
			return ErrUserNotFound
		}
		return err
	}
	if user.IsActive == isActive {
		return nil
	}
	user.IsActive = isActive
	if err := s.repo.UpdateUser(ctx, *user); err != nil {
		return err
	}

	activePrs, err := s.repo.GetUsersPr(ctx, user.Id, true)
	if err != nil {
		return fmt.Errorf("failed to get user PRs for reassignment: %w", err)
	}
	team, _ := s.repo.GetTeam(ctx, user.TeamID)
	if !isActive {
		for _, pr := range activePrs {
			if err := s.ReassignPullRequest(ctx, pr, *user); err != nil {
				return fmt.Errorf("failed to reassign PR %s: %w", pr.Id, err)
			}
		}
		if err := s.repo.RemoveReviewerFromAllPR(ctx, user.Id); err != nil {
			return fmt.Errorf("failed to remove old reviewer %s from all PRs: %w", user.Id, err)
		}
	} else {
		activePrs, err = s.repo.GetTeamPr(ctx, user.TeamID)
		if err != nil {
			return fmt.Errorf("failed to get team for user %s: %w", user.Id, err)
		}
		for _, pr := range activePrs {
			if !pr.NeedMoreReviewers {
				continue
			}
			alreadyReviewer := false
			for _, r := range pr.Reviewers {
				if r.Id == user.Id {
					alreadyReviewer = true
					break
				}
			}
			if !alreadyReviewer && user.Id != pr.Author.Id {
				pr.Reviewers = append(pr.Reviewers, *user)
				activeReviewersCount := 0
				for _, r := range pr.Reviewers {
					for _, tUser := range team.Users {
						if tUser.Id == r.Id && tUser.IsActive {
							activeReviewersCount++
							break
						}
					}
				}
				if activeReviewersCount >= 2 {
					pr.NeedMoreReviewers = false
				}
				user.IsActive = isActive
				if err := s.repo.UpdateUser(ctx, *user); err != nil {
					return err
				}
				if err := s.repo.UpdatePr(ctx, pr.Id, pr); err != nil {
					return fmt.Errorf("failed to update PR %s after activating user: %w", pr.Id, err)
				}
				s.repo.AddReviewerToPR(ctx, pr.Id, userId)
			}
		}
	}
	return nil
}

func (s *PrService) GetUserWithTeam(ctx context.Context, userId string) (*entityUser.User, string, error) {
	user, team, err := s.repo.GetUserWithTeam(ctx, userId)
	if err != nil {
		if errors.Is(err, repos.ErrNoUserWithId) {
			return nil, "", ErrUserNotFound
		}
		return nil, "", err
	}
	return user, team, nil
}

func (s *PrService) CreatePR(ctx context.Context, prDto dto.CreatePR) (*entityPR.PullRequest, error) {
	potentialPr, err := s.repo.GetPr(ctx, prDto.PrID)
	if err == nil && potentialPr != nil {
		return nil, ErrPrIsAlreadyCreated
	}
	if err != nil && !errors.Is(err, repos.ErrPrNotFound) {
		return nil, fmt.Errorf("failed to check existing PR: %w", err)
	}
	author, teamName, err := s.GetUserWithTeam(ctx, prDto.PrAuthor)
	if err != nil {
		return nil, ErrAuthorOrTeamAreNotFound
	}
	team, err := s.repo.GetTeamByName(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	pr := &entityPR.PullRequest{
		Id:                prDto.PrID,
		Name:              prDto.PrName,
		Author:            *author,
		Reviewers:         []entityUser.User{},
		NeedMoreReviewers: false,
		Status:            "OPEN",
		CreatedAt:         time.Now(),
	}
	activeUsers := make([]entityUser.User, 0)
	for _, u := range team.Users {
		if u.IsActive && u.Id != author.Id {
			activeUsers = append(activeUsers, u)
		}
	}
	rand.Shuffle(len(activeUsers), func(i, j int) {
		activeUsers[i], activeUsers[j] = activeUsers[j], activeUsers[i]
	})

	if len(activeUsers) == 0 {
		pr.NeedMoreReviewers = true
	} else if len(activeUsers) < 2 {
		pr.Reviewers = append(pr.Reviewers, activeUsers...)
		pr.NeedMoreReviewers = true
	} else {
		pr.Reviewers = append(pr.Reviewers, activeUsers[0], activeUsers[1])
	}

	if err := s.repo.AddPR(ctx, *pr); err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	return pr, nil
}

func (s *PrService) GetUsersPr(ctx context.Context, userId string) ([]entity.PullRequest, error) {
	user, err := s.repo.GetUserByID(ctx, userId)
	if err != nil {
		if errors.Is(err, repos.ErrNoUserWithId) {
			return nil, ErrUserNotFound
		}
	}
	prs, err := s.repo.GetUsersPr(ctx, user.Id, false)
	if err != nil {
		return nil, err
	}
	return prs, nil
}

func (s *PrService) Merge(ctx context.Context, userId string, prId string) (*entityPR.PullRequest, error) {
	pr, err := s.repo.GetPr(ctx, prId)
	if err != nil {
		if errors.Is(err, repos.ErrPrNotFound) {
			return nil, ErrPrNotFound
		}
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}
	isReviewer := false
	for _, r := range pr.Reviewers {
		if r.Id == userId {
			isReviewer = true
			break
		}
	}
	if userId == "admin" {
		isReviewer = true
	}
	if !isReviewer {
		return nil, ErrUnableToMerge
	}
	if pr.Status != "OPEN" {
		return nil, ErrUnableToMerge
	}
	pr.Status = "MERGED"
	now := time.Now().UTC()
	pr.MergedAt = &now
	if err := s.repo.UpdatePr(ctx, pr.Id, *pr); err != nil {
		return nil, fmt.Errorf("failed to update PR: %w", err)
	}
	return pr, nil
}

func (s *PrService) Reassign(ctx context.Context, prID, oldReviewerID string) (*entityPR.PullRequest, string, error) {
	pr, err := s.repo.GetPr(ctx, prID)
	if err != nil {
		if errors.Is(err, repos.ErrPrNotFound) {
			return nil, "", ErrPrNotFound
		}
		return nil, "", fmt.Errorf("cannot get PR: %w", err)
	}
	if pr.Status == "MERGED" {
		return nil, "", ErrPrIsMerged
	}
	found := false
	for _, r := range pr.Reviewers {
		if r.Id == oldReviewerID {
			found = true
			break
		}
	}
	if !found {
		return nil, "", ErrNotAssigned
	}

	user, err := s.repo.GetUserByID(ctx, pr.Author.Id)
	if err != nil {
		return nil, "", err
	}

	team, err := s.repo.GetTeam(ctx, user.TeamID)
	if err != nil {
		return nil, "", fmt.Errorf("cannot get team: %w", err)
	}
	candidates := make([]entityUser.User, 0)
	for _, u := range team.Users {
		if u.IsActive && u.Id != oldReviewerID {
			alreadyAssigned := false
			for _, r := range pr.Reviewers {
				if r.Id == u.Id {
					alreadyAssigned = true
					break
				}
			}
			if !alreadyAssigned {
				candidates = append(candidates, u)
			}
		}
	}

	if len(candidates) == 0 {
		return nil, "", ErrNoCandidate
	}
	newReviewer := candidates[rand.Intn(len(candidates))]

	updatedReviewers := make([]entityUser.User, 0, len(pr.Reviewers))
	for _, r := range pr.Reviewers {
		if r.Id != oldReviewerID {
			updatedReviewers = append(updatedReviewers, r)
		}
	}
	updatedReviewers = append(updatedReviewers, newReviewer)
	pr.Reviewers = updatedReviewers
	activeCount := 0
	for _, r := range pr.Reviewers {
		for _, u := range team.Users {
			if u.Id == r.Id && u.IsActive {
				activeCount++
			}
		}
	}
	pr.NeedMoreReviewers = activeCount < 2

	if err := s.repo.UpdatePr(ctx, prID, *pr); err != nil {
		return nil, "", fmt.Errorf("failed to update PR reviewers: %w", err)
	}

	return pr, newReviewer.Id, nil
}
