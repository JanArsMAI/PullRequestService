package rest

import (
	_ "github.com/JanArsMAI/PullRequestService/docs"
	"github.com/JanArsMAI/PullRequestService/internal/application"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func InitRoutes(r *gin.Engine, svc *application.PrService, logger *zap.Logger) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	h := NewHandlers(svc, logger)
	apiTeam := r.Group("team")
	{
		apiTeam.POST("/add", h.AddTeam)
		apiTeam.GET("/get_team/:team_name", h.UserMiddleware(), h.GetTeam)
	}

	apiUsers := r.Group("users")
	{
		apiUsers.POST("/setIsActive", h.AdminMiddleware(), h.SetIsActive)
		apiUsers.GET("/getReview/:user_id", h.UserMiddleware(), h.GetUsersPr)
	}

	apiPullRequests := r.Group("pullRequest")
	{
		apiPullRequests.POST("/create", h.AdminMiddleware(), h.CreatePR)
		apiPullRequests.POST("/merge", h.UserMiddleware(), h.Merge)
		apiPullRequests.POST("/reassign", h.AdminMiddleware(), h.Reasign)
	}
	r.Use(CORSMiddleware())
}
