package handler

import (
	"github.com/hnam/notafly/internal/config"
	"go.uber.org/zap"
)

type Handler struct {
	config *config.Config
	logger *zap.Logger
}

func New(cfg *config.Config, logger *zap.Logger) *Handler {
	return &Handler{
		config: cfg,
		logger: logger,
	}
}
