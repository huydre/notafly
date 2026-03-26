package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hnam/notafly/internal/dto"
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

	// TODO: Phase 5 — wire TranscriberService
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
		Error: "not implemented yet",
	})
}
