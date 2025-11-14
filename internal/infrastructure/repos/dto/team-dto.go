package dto

import entity "github.com/JanArsMAI/PullRequestService/internal/domain/team"

type TeamDto struct {
	Id   int    `db:"id"`
	Name string `db:"team_name"`
}

func TeamFromEntity(team entity.Team) TeamDto {
	return TeamDto{
		Id:   team.Id,
		Name: team.Name,
	}
}

func (t TeamDto) TeamToEntity() entity.Team {
	return entity.Team{
		Id:   t.Id,
		Name: t.Name,
	}
}
