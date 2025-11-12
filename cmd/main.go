package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JanArsMAI/PullRequestService/internal/config"
	"github.com/JanArsMAI/PullRequestService/internal/di"
	zapLogger "github.com/JanArsMAI/PullRequestService/logger"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func listenRESTServer(r *gin.Engine, logger *zap.Logger, cfg config.ServerConfig) *http.Server {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: r,
	}

	go func() {
		logger.Info("Starting server", zap.Int("port", cfg.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start REST server: %v", zap.Error(err))
		}
	}()

	return server
}

func main() {
	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatal(".env loading error")
	}
	cfgPath := os.Getenv("CONFIG_PATH")
	cfg, err := config.MustLoad(cfgPath)
	if err != nil {
		log.Fatal(".yaml loading error")
	}
	logger := zapLogger.NewLogger(cfg.Logging.Level)
	r := gin.Default()
	close := di.ConfigureApp(logger, cfg.PR)
	defer close()
	serverREST := listenRESTServer(r, logger, cfg.Server)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := serverREST.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown: %v", zap.Error(err))
	}
	logger.Info("Server stopped")
}
