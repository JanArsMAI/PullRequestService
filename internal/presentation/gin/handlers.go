package rest

import (
	"errors"
	"net/http"

	"github.com/JanArsMAI/PullRequestService/internal/application"
	"github.com/JanArsMAI/PullRequestService/internal/domain/interfaces"
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
	svc    interfaces.PrService
	logger *zap.Logger
}

func NewHandlers(service interfaces.PrService, logger *zap.Logger) *Handlers {
	return &Handlers{
		svc:    service,
		logger: logger,
	}
}

const (
	CodeBadRequest   = "BAD_REQUEST"
	CodeNotFound     = "NOT_FOUND"
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
	CodePrExists     = "PR_EXISTS"
	CodeNoCandidate  = "NO_CANDIDATE"
	CodeNotAssigned  = "NOT_ASSIGNED"
)

// AddTeam godoc
// @Summary Создание новой команды
// @Description Создаёт новую команду. Если пользователь уже в другой команде, PR пользователя переназначается на участников старой команды.
// @Tags team
// @Accept json
// @Produce json
// @Param team body dto.AddTeamRequest true "Данные команды"
// @Success 201 {object} dto.TeamResponse "Команда создана"
// @Failure 400 {object} dto.ErrorResponse "Команда уже существует или некоректный запрос"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /team/add [post]
func (h *Handlers) AddTeam(ctx *gin.Context) {
	var body dto.AddTeamRequest
	err := ctx.BindJSON(&body)
	if err != nil {
		h.logger.Warn("invalid format of request to Add Team, error parsing JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeBadRequest,
				Message: "invalid format of request to Add Team, error parsing JSON",
			},
		})
		return
	}
	if body.TeamName == "" {
		h.logger.Warn("invalid format of request to Add Team, empty team name")
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeBadRequest,
				Message: "invalid format of request to Add Team, empty team name",
			},
		})
		return
	}
	if len(body.Members) == 0 {
		h.logger.Warn("invalid format of request to Add Team, empty team members")
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeBadRequest,
				Message: "invalid format of request to Add Team, empty team members",
			},
		})
		return
	}
	err = h.svc.AddTeam(ctx, &body)
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
	members := make([]dto.MemberDtoResponse, 0, len(body.Members))
	for _, user := range body.Members {
		members = append(members, dto.MemberDtoResponse{
			Id:       user.Id,
			Name:     user.Name,
			IsActive: user.IsActive,
		})
	}
	resp := dto.TeamResponse{
		Team: dto.TeamDtoResponse{
			TeamName: body.TeamName,
			Members:  members,
		},
	}
	h.logger.Info("successfully added team", zap.String("team_name", body.TeamName))
	ctx.JSON(http.StatusCreated, resp)
}

// GetTeam godoc
// @Summary Получение информации о команде
// @Description Возвращает информацию о команде и её участниках. Доступ разрешён только для участников команды.
// @Tags team
// @Accept json
// @Produce json
// @Param Authorization header string true "токен пользователя(Вводить без Bearer)"
// @Param team_name query string true "Уникальное имя команды"
// @Success 200 {object} dto.TeamResponse "Обьект команды"
// @Failure 400 {object} dto.ErrorResponse "Пустое имя команды"
// @Failure 401 {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure 403 {object} dto.ErrorResponse "Доступ запрещён, пользователь не в команде"
// @Failure 404 {object} dto.ErrorResponse "Команда не найдена"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /team/get [get]
func (h *Handlers) GetTeam(ctx *gin.Context) {
	//сымитируем проверку токена пользователя: токен - это idпользователя, для упрощения
	//реализации было сделано так
	userId, ok := ctx.Get("User_Id")
	if !ok {
		h.logger.Warn("unauthorized user to get Team")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeUnauthorized,
				Message: "No token to get info of Team",
			},
		})
		return
	}
	par := ctx.Query("team_name")
	if par == "" {
		h.logger.Warn("GetTeam: empty team_name query parameter")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeBadRequest,
				Message: "team_name query parameter is required",
			},
		})
		return
	}
	team, err := h.svc.GetTeam(ctx, par)

	if err != nil {
		if errors.Is(err, application.ErrTeamNotFound) {

			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Error: dto.ErrorMessage{
					Code:    CodeNotFound,
					Message: "resource not found",
				},
			})
			h.logger.Warn("Not found team", zap.String("target_name", par))
			return
		}
		ctx.AbortWithStatus(http.StatusInternalServerError)
		h.logger.Error("error wthile getting team", zap.Error(err), zap.String("target_name", par))
		return
	}
	var isInTeam = false
	members := make([]dto.MemberDtoResponse, 0, len(team.Users))
	for _, user := range team.Users {
		members = append(members, dto.MemberDtoResponse{
			Id:       user.Id,
			Name:     user.Name,
			IsActive: user.IsActive,
		})
		if user.Id == userId.(string) {
			isInTeam = true
		}
	}
	//если пользователя нет в команде, которую запрашщивает, то запрещаем доступ
	if !isInTeam {
		h.logger.Warn("Forbidden access for user to get team", zap.String("team", team.Name), zap.String("user_id", userId.(string)))
		ctx.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeForbidden,
				Message: "User is not in found team",
			},
		})
		return
	}

	resp := dto.TeamResponse{
		Team: dto.TeamDtoResponse{
			TeamName: team.Name,
			Members:  members,
		},
	}
	ctx.JSON(http.StatusOK, resp)
	h.logger.Info("successfully got team", zap.String("team_name", team.Name))
}

// SetIsActive godoc
// @Summary Установить флаг активности пользователя
// @Description Позволяет изменить статус активности пользователя (активен/неактивен). Доступно только администраторам.
// @Tags Users
// @Accept json
// @Produce json
// @Param Authorization header string true "токен администратора(Вводить без Bearer)"
// @Param user body dto.SetUserActive true "Данные пользователя для изменения статуса активности"
// @Success 200 {object} dto.UserResponse "Обновлённый пользователь"
// @Failure 400 {object} dto.ErrorResponse "Некорректные данные запроса"
// @Failure 401 {object} dto.ErrorResponse "Нет/неверный админский токен"
// @Failure 404 {object} dto.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /users/setIsActive [post]
func (h *Handlers) SetIsActive(ctx *gin.Context) {
	var body dto.SetUserActive
	err := ctx.BindJSON(&body)
	if err != nil {
		h.logger.Warn("invalid format of request to Set user activity, error parsing JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeBadRequest,
				Message: "invalid format of request to Set user activity",
			},
		})
		return
	}
	if body.UserId == "" {
		h.logger.Warn("no user_id provided in SetUserActive")
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeBadRequest,
				Message: "No choosen user for changing info",
			},
		})
		return
	}
	err = h.svc.SetUserActive(ctx, body.UserId, body.IsActive)
	if err != nil {
		if errors.Is(err, application.ErrUserNotFound) {
			h.logger.Warn("user not found while SetUserActive", zap.String("user_id", body.UserId))
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: dto.ErrorMessage{
					Code:    CodeNotFound,
					Message: "user not found",
				},
			})
			return
		}

		h.logger.Error("failed to SetUserActive", zap.Error(err), zap.String("user_id", body.UserId))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	user, team, err := h.svc.GetUserWithTeam(ctx, body.UserId)
	if err != nil {
		h.logger.Error("failed to SetUserActive", zap.Error(err), zap.String("user_id", body.UserId))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	resp := dto.UserResponse{
		User: dto.UserWithTeam{
			Id:       user.Id,
			Name:     user.Name,
			IsActive: user.IsActive,
			Team:     team,
		},
	}
	ctx.JSON(http.StatusOK, resp)
	h.logger.Info("Successfully updated user", zap.String("user_id", user.Id))
}

// CreatePR godoc
// @Summary      Создать Pull Request
// @Description  Создаёт новый Pull Request от указанного автора.
// @Description  Ревьюеры выбираются автоматически на основе команды автора.
// @Tags         PullRequests
// @Accept       json
// @Produce      json
// @Param Authorization header string true "токен администратора(Вводить без Bearer)"
// @Param        body   body      dto.CreatePR  true  "Данные для создания Pull Request"
// @Success      201    {object}  dto.PullRequestResponse "PR успешно создан"
// @Failure      400    {object}  dto.ErrorResponse "Некорректный формат запроса"
// @Failure      404    {object}  dto.ErrorResponse "Автор или команда не найдены"
// @Failure      409    {object}  dto.ErrorResponse "PR с таким ID уже существует"
// @Failure      500    "Внутренняя ошибка сервера"
// @Router       /pullRequest/create [post]
func (h *Handlers) CreatePR(ctx *gin.Context) {
	var body dto.CreatePR
	err := ctx.ShouldBindJSON(&body)
	if err != nil {
		h.logger.Warn("invalid format of request to Create PR, error parsing JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeBadRequest,
				Message: "invalid format of request to Create PR, error parsing JSON",
			},
		})
		return
	}
	if body.PrAuthor == "" {
		h.logger.Warn("no author_id provided in CreatePR")
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeBadRequest,
				Message: "No choosen author for creating PR",
			},
		})
		return
	}
	if body.PrID == "" {
		h.logger.Warn("no pr_id provided in CreatePR")
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeBadRequest,
				Message: "No choosen PR id for creating PR",
			},
		})
		return
	}
	if body.PrName == "" {
		h.logger.Warn("no pr_name provided in CreatePR")
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeBadRequest,
				Message: "No choosen PR name for creating PR",
			},
		})
		return
	}
	pr, err := h.svc.CreatePR(ctx, body)
	if err != nil {
		if errors.Is(err, application.ErrPrIsAlreadyCreated) {
			h.logger.Warn("Pr with this ID is already created", zap.String("Pr_id", body.PrID))
			ctx.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: dto.ErrorMessage{
					Code:    CodePrExists,
					Message: "PR id already exists",
				},
			})
			return
		}
		if errors.Is(err, application.ErrAuthorOrTeamAreNotFound) {
			h.logger.Warn("Author or team of this PR not exist", zap.String("Pr_id", body.PrID),
				zap.String("author_id", body.PrAuthor))
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: dto.ErrorMessage{
					Code:    CodeNotFound,
					Message: "resource not found",
				},
			})
			return
		}
		h.logger.Error("failed to create PR", zap.Error(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	reviewers := make([]string, 0, len(pr.Reviewers))
	for _, reviewer := range pr.Reviewers {
		reviewers = append(reviewers, reviewer.Id)
	}
	resp := dto.PullRequestResponse{
		Pr: dto.PullRequest{
			Id:        pr.Id,
			Name:      pr.Name,
			AuthorId:  pr.Author.Id,
			Status:    pr.Status,
			Reviewers: reviewers,
		},
	}
	h.logger.Info("successfully created Pull Request", zap.String("author_id", pr.Author.Id), zap.String("Pr_id", pr.Id))
	ctx.JSON(http.StatusCreated, resp)
}

// GetUsersPr godoc
// @Summary Получить PR'ы, где пользователь назначен ревьювером
// @Description Получить PR'ы, где пользователь назначен ревьювером
// @Tags Users
// @Param Authorization header string true "токен пользователя (вводить без Bearer)"
// @Param user_id query string true "Id пользователя"
// @Success 200 {object} dto.UsersPrResponse "Список PR'ов пользователя"
// @Failure 400 {object} dto.ErrorResponse "Некорректный формат запроса"
// @Failure 401 {object} dto.ErrorResponse "Пользователь не авторизован"
// @Failure 404 {object} dto.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /users/getReview [get]
func (h *Handlers) GetUsersPr(ctx *gin.Context) {
	_, ok := ctx.Get("User_Id")
	if !ok {
		h.logger.Warn("unauthorized user to get PR")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeUnauthorized,
				Message: "No token to get info of PR",
			},
		})
		return
	}
	par := ctx.Query("user_id")
	if par == "" {
		h.logger.Warn("GetUsersPr: empty user_id parameter")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeBadRequest,
				Message: "No choosen user for getting PR",
			},
		})
		return
	}
	prs, err := h.svc.GetUsersPr(ctx, par)
	if err != nil {
		if errors.Is(err, application.ErrUserNotFound) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Error: dto.ErrorMessage{
					Code:    CodeNotFound,
					Message: "resource not found",
				},
			})
			h.logger.Warn("Not found user", zap.String("target_name", par))
			return
		}
		h.logger.Error("failed to get user PRs", zap.Error(err))
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	prResp := make([]dto.PullRequestOfUser, 0, len(prs))
	for _, pr := range prs {
		prResp = append(prResp, dto.PullRequestOfUser{
			Id:       pr.Id,
			Name:     pr.Name,
			AuthorId: pr.Author.Id,
			Status:   pr.Status,
		})
	}
	resp := dto.UsersPrResponse{
		UserId: par,
		Prs:    prResp,
	}
	ctx.JSON(http.StatusOK, resp)
	h.logger.Info("successfully got users pr", zap.String("user_id", par))
}

// Merge godoc
// @Summary Пометить PR как MERGED
// @Description Пометить PR как MERGED
// @Tags PullRequests
// @Param body body dto.MergeRequest true "Id PR"
//
//	@Param Authorization header string true "токен пользователя(Вводить без Bearer)"
//	@Success 200 {object} dto.MergeResponse "PR в состоянии MERGED"
//	@Failure 400 {object} dto.ErrorResponse "Некорректный формат запроса"
//
// @Failure 401 {object} dto.ErrorResponse "Некорректный формат запроса"
// @Failure 404 {object} dto.ErrorResponse "PR не найден"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /pullRequest/merge [post]
func (h *Handlers) Merge(ctx *gin.Context) {
	var body dto.MergeRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeBadRequest,
				Message: "Invalid merge request body",
			},
		})
		h.logger.Warn("invalid merge body", zap.Error(err))
		return
	}

	userId, ok := ctx.Get("User_Id")
	if !ok {
		h.logger.Warn("unauthorized user to merge PR")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeUnauthorized,
				Message: "No token to merge PR",
			},
		})
		return
	}

	pr, err := h.svc.Merge(ctx, userId.(string), body.Id)
	if err != nil {
		switch err {
		case application.ErrPrNotFound:
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Error: dto.ErrorMessage{
					Code:    CodeNotFound,
					Message: "resource not found",
				},
			})
			h.logger.Warn("PR not found", zap.String("pr_id", body.Id))
			return
		case application.ErrUnableToMerge:
			ctx.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{
				Error: dto.ErrorMessage{
					Code:    CodeForbidden,
					Message: "User is not a reviewer, unable to merge",
				},
			})
			h.logger.Warn("unable to merge, user is not a reviewer",
				zap.String("pr_id", body.Id),
				zap.String("user_id", userId.(string)),
			)
			return
		default:
			ctx.AbortWithStatus(http.StatusInternalServerError)
			h.logger.Error("error while merging PR",
				zap.Error(err),
				zap.String("pr_id", body.Id),
				zap.String("user_id", userId.(string)),
			)
			return
		}
	}

	reviewers := make([]string, len(pr.Reviewers))
	for i, r := range pr.Reviewers {
		reviewers[i] = r.Id
	}

	resp := dto.MergeResponse{
		Pr: dto.MergedPullRequestOfUser{
			Id:        pr.Id,
			Name:      pr.Name,
			AuthorId:  pr.Author.Id,
			Status:    pr.Status,
			Reviewers: reviewers,
			MergeAt:   pr.MergedAt,
		},
	}

	ctx.JSON(http.StatusOK, resp)
	h.logger.Info("successfully merged PR", zap.String("pr_id", body.Id))
}

// Reasign godoc
// @Summary Переназначить конкретного ревьювера на другого из его команды
// @Description Переназначить конкретного ревьювера на другого из его команды
// @Tags PullRequests
// @Param body body dto.ReassignPullRequest true "data"
//
//	@Param Authorization header string true "токен администратора(Вводить без Bearer)"
//	@Success 200 {object} dto.MergeResponse "Переназначение выполнено"
//	@Failure 400 {object} dto.ErrorResponse "Некорректный формат запроса"
//
// @Failure 401 {object} dto.ErrorResponse "Некорректный формат запроса"
// @Failure 404 {object} dto.ErrorResponse "PR не найден"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /pullRequest/reassign [post]
func (h *Handlers) Reasign(ctx *gin.Context) {
	var body dto.ReassignPullRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeBadRequest,
				Message: "Invalid reassign PR body",
			},
		})
		h.logger.Warn("invalid reassign PR body", zap.Error(err))
		return
	}

	if body.OldReviewer == "" {
		h.logger.Warn("no old_reviewer_id provided in Reasign")
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeBadRequest,
				Message: "No old_reviewer_id for reassigning PR",
			},
		})
		return
	}
	if body.PrID == "" {
		h.logger.Warn("no pull_request_id provided in Reasign")
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorMessage{
				Code:    CodeBadRequest,
				Message: "No pull_request_id for reassigning PR",
			},
		})
		return
	}

	pr, replacedBy, err := h.svc.Reassign(ctx, body.PrID, body.OldReviewer)
	if err != nil {
		switch err {
		case application.ErrPrIsMerged:
			ctx.AbortWithStatusJSON(http.StatusConflict, dto.ErrorResponse{
				Error: dto.ErrorMessage{
					Code:    "PR_MERGED",
					Message: "cannot reassign on merged PR",
				},
			})
			h.logger.Warn(" pull_request_id is already merged", zap.String("pr_id", body.PrID))
			return
		case application.ErrPrNotFound:
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Error: dto.ErrorMessage{
					Code:    CodeNotFound,
					Message: "resource not found",
				},
			})
			h.logger.Warn(" pull_request_id is not found", zap.String("pr_id", body.PrID))
			return
		case application.ErrTeamNotFound:
			ctx.AbortWithStatus(http.StatusNotFound)
			h.logger.Error(" error while reassigning PR", zap.String("pr_id", body.PrID), zap.Error(err))
			return
		case application.ErrNoCandidate:
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Error: dto.ErrorMessage{
					Code:    CodeNoCandidate,
					Message: "no available reviewer",
				},
			})
			h.logger.Warn("no available reviewer to reassign PR", zap.String("pr_id", body.PrID))
			return
		case application.ErrNotAssigned:
			ctx.AbortWithStatusJSON(http.StatusNotFound, dto.ErrorResponse{
				Error: dto.ErrorMessage{
					Code:    CodeNotAssigned,
					Message: "reviewer is not assigned to this PR",
				},
			})
			h.logger.Warn(" no candidate to reassign PR", zap.String("pr_id", body.PrID))
			return
		default:
			ctx.AbortWithStatus(http.StatusInternalServerError)
			h.logger.Error(" error while reassigning PR", zap.String("pr_id", body.PrID), zap.Error(err))
			return
		}
	}
	assignedReviewers := make([]string, 0, len(pr.Reviewers))
	for _, r := range pr.Reviewers {
		assignedReviewers = append(assignedReviewers, r.Id)
	}

	resp := dto.ReassignResponse{
		Pr: dto.ReassignPR{
			PullRequestID:     pr.Id,
			PullRequestName:   pr.Name,
			AuthorID:          pr.Author.Id,
			Status:            pr.Status,
			AssignedReviewers: assignedReviewers,
		},
		ReplacedBy: replacedBy,
	}

	ctx.JSON(http.StatusOK, resp)
	h.logger.Info("successfully reassigned PR", zap.String("pr_id", body.PrID), zap.String("replaced_by", replacedBy))
}
