package application

import "github.com/JanArsMAI/PullRequestService/internal/infrastructure/repos"

type PrService struct {
	repo *repos.PostgresRepo
}

func NewPrService(repo *repos.PostgresRepo) *PrService {
	return &PrService{
		repo: repo,
	}
}
