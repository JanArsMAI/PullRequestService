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
	api := r.Group("team")
	{
		api.POST("/add", h.AddTeam)
	}
	r.Use(CORSMiddleware())
}
