package handler

import (
	"github.com/hnam/notafly/internal/config"
	"github.com/hnam/notafly/internal/service"
	"go.uber.org/zap"
)

type Handler struct {
	config  *config.Config
	logger  *zap.Logger
	meetSvc *service.MeetService
}

func New(cfg *config.Config, logger *zap.Logger, meetSvc *service.MeetService) *Handler {
	return &Handler{
		config:  cfg,
		logger:  logger,
		meetSvc: meetSvc,
	}
}
