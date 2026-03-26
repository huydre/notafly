package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hnam/notafly/internal/dto"
	"github.com/hnam/notafly/internal/model"
	"go.uber.org/zap"
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

	sessionID := uuid.New().String()
	audioDir, err := os.MkdirTemp("", "notafly-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "failed to create temp directory",
		})
		return
	}
	audioPath := filepath.Join(audioDir, "output.wav")

	session := &model.MeetingSession{
		ID:        sessionID,
		MeetLink:  req.MeetLink,
		AudioPath: audioPath,
		Duration:  req.Duration,
		Status:    model.StatusJoining,
		CreatedAt: time.Now(),
	}

	h.logger.Info("joining meeting",
		zap.String("session_id", session.ID),
		zap.String("meet_link", req.MeetLink),
	)

	browserCtx, cancelBrowser, err := h.meetSvc.JoinMeeting(c.Request.Context(), req.MeetLink)
	if err != nil {
		session.Status = model.StatusFailed
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "failed to join meeting",
			Details: err.Error(),
		})
		return
	}

	// Keep browser alive — caller is responsible for cleanup
	_ = browserCtx
	_ = cancelBrowser

	session.Status = model.StatusRecording
	h.logger.Info("meeting joined, ready for recording",
		zap.String("session_id", session.ID),
	)

	c.JSON(http.StatusOK, dto.JoinMeetResponse{
		SessionID: session.ID,
		Status:    string(session.Status),
		AudioPath: session.AudioPath,
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

	// TODO: Phase 4-5 — wire recorder + transcriber for full pipeline
	c.JSON(http.StatusNotImplemented, dto.ErrorResponse{
		Error: fmt.Sprintf("full pipeline not implemented yet (meet_link=%s)", req.MeetLink),
	})
}
