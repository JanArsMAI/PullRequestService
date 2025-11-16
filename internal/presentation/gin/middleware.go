package rest

import (
	"net/http"

	"github.com/JanArsMAI/PullRequestService/internal/presentation/gin/dto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	AdminToken = "admin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// middleware для проверки, что токен - админский
// здесь допустим небольшое упрощение и будем сравнивать с константным значением токена
func (h *Handlers) AdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//здесь должна быть имитация похода по gRPC на другой микросервис авторизации для проверки токена
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			h.logger.Warn("AdminMiddleware: missing Authorization header")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: dto.ErrorMessage{
					Code:    CodeNotFound,
					Message: "resource not found",
				},
			})
			return
		}
		if authHeader != AdminToken {
			h.logger.Warn("AdminMiddleware: invalid admin token", zap.String("header", authHeader))
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: dto.ErrorMessage{
					Code:    CodeNotFound,
					Message: "resource not found",
				},
			})
			return
		}
		ctx.Next()
	}
}

func (h *Handlers) UserMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//аналогично, идём на gRPC, для проверки токена, но тут мы этого не делаем
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			h.logger.Warn("UserMiddleware: missing Authorization header")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{
				Error: dto.ErrorMessage{
					Code:    CodeNotFound,
					Message: "resource not found",
				},
			})
			return
		}
		userID := authHeader
		ctx.Set("User_Id", userID)
		ctx.Next()
	}
}
