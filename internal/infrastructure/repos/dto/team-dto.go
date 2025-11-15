package dto

type TeamDto struct {
	Id   int    `db:"id"`
	Name string `db:"team_name"`
}
