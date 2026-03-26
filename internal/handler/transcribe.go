package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hnam/notafly/internal/dto"
	"go.uber.org/zap"
)

func (h *Handler) Transcribe(c *gin.Context) {
	var req dto.TranscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid request",
			Details: err.Error(),
		})
		return
	}

	h.logger.Info("transcribing audio", zap.String("path", req.AudioPath))

	if req.NoAnalysis {
		// Transcribe only — no GPT analysis
		text, err := h.transcriberSvc.Transcribe(c.Request.Context(), req.AudioPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error:   "transcription failed",
				Details: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, dto.TranscribeResponse{
			Transcription: text,
		})
		return
	}

	// Full pipeline: transcribe + analyze
	result, err := h.transcriberSvc.TranscribeAndAnalyze(c.Request.Context(), req.AudioPath, h.recorderSvc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "transcription and analysis failed",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.TranscribeResponse{
		Transcription: result.Text,
		Minutes:       &result.Minutes,
	})
}
