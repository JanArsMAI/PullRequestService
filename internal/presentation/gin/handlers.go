package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/JanArsMAI/PullRequestService/internal/application"
	"github.com/JanArsMAI/PullRequestService/internal/presentation/gin/dto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @title Pull Request Service
// @version 1.0
// @description API для управления Pull Request
// @host localhost:8080
// @BasePath /
// @schemes http https

type Handlers struct {
	svc    *application.PrService
	logger *zap.Logger
}

func NewHandlers(service *application.PrService, logger *zap.Logger) *Handlers {
	return &Handlers{
		svc:    service,
		logger: logger,
	}
}

// AddTeam godoc
// @Summary Создаёт новую команду
// @Description Создаёт новую команду. Если пользователь уже в команде, то происходит переназначение его PR на пользователей его старой команды.
// @Tags team
// @Accept json
// @Produce json
// @Param team body dto.AddTeamRequest true "Team data"
// @Success 201 {object} dto.TeamResponse "Team successfully created"
// @Failure 400 {object} dto.ErrorResponse "Bad request, invalid data"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /team/add [post]
func (h *Handlers) AddTeam(ctx *gin.Context) {
	var body dto.AddTeamRequest
	err := ctx.BindJSON(&body)
	if err != nil {
		h.logger.Warn("invalid format of request to Add Team, error parsing JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    "BAD_REQUEST",
				Message: "invalid format of request to Add Team, error parsing JSON",
			},
		})
		return
	}
	if body.TeamName == "" {
		h.logger.Warn("invalid format of request to Add Team, empty team name")
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    "BAD_REQUEST",
				Message: "invalid format of request to Add Team, empty team name",
			},
		})
		return
	}
	if len(body.Members) == 0 {
		h.logger.Warn("invalid format of request to Add Team, empty team members")
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    "BAD_REQUEST",
				Message: "invalid format of request to Add Team, empty team members",
			},
		})
		return
	}
	err = h.svc.AddTeam(context.Background(), &body)
	if err != nil {
		if errors.Is(err, application.ErrTeamWithNameAlreadyCreated) {
			h.logger.Warn("invalid name to Add Team, error parsing JSON", zap.Error(err), zap.String("name", body.TeamName))
			ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: dto.ErrorMessage{
					Code:    "TEAM_EXISTS",
					Message: "team_name already exists",
				},
			})
			return
		}
		h.logger.Error("error to Add team", zap.Error(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	resp := dto.TeamResponse{
		Team: dto.TeamDtoResponse{
			TeamName: body.TeamName,
			Members:  body.Members,
		},
	}
	ctx.JSON(http.StatusCreated, resp)
}
