package entity

import entity "github.com/JanArsMAI/PullRequestService/internal/domain/user"

type Team struct {
	Id    int
	Name  string
	Users []entity.User
}
