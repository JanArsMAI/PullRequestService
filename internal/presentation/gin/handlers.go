package rest

import (
	"github.com/JanArsMAI/PullRequestService/internal/application"
	"go.uber.org/zap"
)

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
