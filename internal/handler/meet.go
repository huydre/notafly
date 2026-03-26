package handler

import (
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

	session := h.newSession(req.MeetLink, req.Duration)

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

	// Keep browser alive — caller responsible for cleanup
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

	session := h.newSession(req.MeetLink, req.Duration)

	// Step 1: Join meeting
	h.logger.Info("full pipeline: joining", zap.String("session_id", session.ID))
	session.Status = model.StatusJoining

	_, cancelBrowser, err := h.meetSvc.JoinMeeting(c.Request.Context(), req.MeetLink)
	if err != nil {
		session.Status = model.StatusFailed
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "failed to join meeting",
			Details: err.Error(),
		})
		return
	}
	defer cancelBrowser()

	// Step 2: Record audio
	h.logger.Info("full pipeline: recording", zap.String("session_id", session.ID))
	session.Status = model.StatusRecording

	if err := h.recorderSvc.Record(c.Request.Context(), session.AudioPath, req.Duration); err != nil {
		session.Status = model.StatusFailed
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "recording failed",
			Details: err.Error(),
		})
		return
	}

	// Step 3: Transcribe + analyze
	h.logger.Info("full pipeline: transcribing", zap.String("session_id", session.ID))
	session.Status = model.StatusTranscribing

	result, err := h.transcriberSvc.TranscribeAndAnalyze(c.Request.Context(), session.AudioPath, h.recorderSvc)
	if err != nil {
		session.Status = model.StatusFailed
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "transcription failed",
			Details: err.Error(),
		})
		return
	}

	session.Status = model.StatusCompleted
	h.logger.Info("full pipeline: complete", zap.String("session_id", session.ID))

	c.JSON(http.StatusOK, gin.H{
		"session_id":    session.ID,
		"status":        string(session.Status),
		"audio_path":    session.AudioPath,
		"transcription": result.Text,
		"minutes":       result.Minutes,
	})
}

// newSession creates a MeetingSession with temp directory for audio.
func (h *Handler) newSession(meetLink string, duration int) *model.MeetingSession {
	audioDir, _ := os.MkdirTemp("", "notafly-*")
	audioPath := filepath.Join(audioDir, "output.wav")

	return &model.MeetingSession{
		ID:        uuid.New().String(),
		MeetLink:  meetLink,
		AudioPath: audioPath,
		Duration:  duration,
		Status:    model.StatusPending,
		CreatedAt: time.Now(),
	}
}
