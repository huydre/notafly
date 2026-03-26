package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hnam/notafly/internal/dto"
)

func (h *Handler) JoinMeet(c *gin.Context) {
	var req dto.JoinMeetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid request",
			Details: err.Error(),
		})
		return
	}

	// TODO: Phase 3 — wire MeetService
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
		Error: "not implemented yet",
	})
}

func (h *Handler) FullPipeline(c *gin.Context) {
	var req dto.JoinMeetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid request",
			Details: err.Error(),
		})
		return
	}

	// TODO: Phase 3-5 — wire full pipeline
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
		Error: "not implemented yet",
	})
}
