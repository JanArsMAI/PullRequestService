package di

import (
	"github.com/JanArsMAI/PullRequestService/internal/application"
	"github.com/JanArsMAI/PullRequestService/internal/config"
	"github.com/JanArsMAI/PullRequestService/internal/infrastructure/db"
	"github.com/JanArsMAI/PullRequestService/internal/infrastructure/repos"
	rest "github.com/JanArsMAI/PullRequestService/internal/presentation/gin"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ConfigureApp(r *gin.Engine, logger *zap.Logger, cfg config.PRConfig) func() {
	logger.Info("Starting configuring app...")

	db, err := db.NewPostgresConnection(db.ReadConfig())
	if err != nil {
		logger.Fatal("failed to connect to db", zap.Error(err))
	}
	repo := repos.NewPostgresRepo(db)
	svc := application.NewPrService(repo)
	rest.InitRoutes(r, svc, logger)
	return func() {
		logger.Sync()
	}
}
