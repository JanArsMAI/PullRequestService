package rest

import (
	"github.com/JanArsMAI/PullRequestService/internal/application"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func InitRoutes(r *gin.Engine, svc *application.PrService, logger *zap.Logger) {
	h := NewHandlers(svc, logger)
	api := r.Group("pull_request_service")

}
