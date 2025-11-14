package dto

type UserDto struct {
	Id       string `db:"user_id"`
	Name     string `db:"username"`
	IsActive bool   `db:"is_active"`
	TeamID   int    `db:"team_id"`
}
