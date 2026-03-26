package handler

import (
	"github.com/hnam/notafly/internal/config"
	"github.com/hnam/notafly/internal/service"
	"go.uber.org/zap"
)

type Handler struct {
	config        *config.Config
	logger        *zap.Logger
	meetSvc       *service.MeetService
	recorderSvc   *service.RecorderService
	transcriberSvc *service.TranscriberService
}

func New(cfg *config.Config, logger *zap.Logger, meetSvc *service.MeetService, recorderSvc *service.RecorderService, transcriberSvc *service.TranscriberService) *Handler {
	return &Handler{
		config:        cfg,
		logger:        logger,
		meetSvc:       meetSvc,
		recorderSvc:   recorderSvc,
		transcriberSvc: transcriberSvc,
	}
}
