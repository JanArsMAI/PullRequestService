package dto

type AddTeamRequest struct {
	TeamName string      `json:"team_name"`
	Members  []MemberDto `json:"members"`
}

type MemberDto struct {
	Id       string `json:"user_id"`
	Name     string `json:"username"`
	IsActive bool   `json:"is_active"`
}
