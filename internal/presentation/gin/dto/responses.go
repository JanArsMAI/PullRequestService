package dto

type TeamResponse struct {
	Team TeamDtoResponse `json:"team"`
}

type TeamDtoResponse struct {
	TeamName string      `json:"team_name"`
	Members  []MemberDto `json:"members"`
}

type ErrorResponse struct {
	Error ErrorMessage `json:"error"`
}

type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
