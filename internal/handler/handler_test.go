package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hnam/notafly/internal/config"
	"github.com/hnam/notafly/internal/dto"
	"go.uber.org/zap"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	logger := zap.NewNop()
	h := New(cfg, logger)

	r := gin.New()
	r.GET("/health", h.Health)

	v1 := r.Group("/api/v1")
	{
		meet := v1.Group("/meet")
		{
			meet.POST("/join", h.JoinMeet)
			meet.POST("/full", h.FullPipeline)
		}
		v1.POST("/transcribe", h.Transcribe)
	}

	return r
}

func TestHealth(t *testing.T) {
	router := setupTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Errorf("status = %q, want %q", body["status"], "ok")
	}
	if body["service"] != "notafly" {
		t.Errorf("service = %q, want %q", body["service"], "notafly")
	}
}

func TestJoinMeet_InvalidRequest(t *testing.T) {
	router := setupTestRouter()
	w := httptest.NewRecorder()
	body := `{"meet_link": "not-a-url"}`
	req, _ := http.NewRequest("POST", "/api/v1/meet/join", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var errResp dto.ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &errResp)
	if errResp.Error != "invalid request" {
		t.Errorf("error = %q, want %q", errResp.Error, "invalid request")
	}
}

func TestJoinMeet_MissingFields(t *testing.T) {
	router := setupTestRouter()
	w := httptest.NewRecorder()
	body := `{}`
	req, _ := http.NewRequest("POST", "/api/v1/meet/join", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestTranscribe_InvalidRequest(t *testing.T) {
	router := setupTestRouter()
	w := httptest.NewRecorder()
	body := `{}`
	req, _ := http.NewRequest("POST", "/api/v1/transcribe", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestJoinMeet_ValidRequest_ReturnsNotImplemented(t *testing.T) {
	router := setupTestRouter()
	w := httptest.NewRecorder()
	body := `{"meet_link": "https://meet.google.com/abc-defg-hij", "duration": 60}`
	req, _ := http.NewRequest("POST", "/api/v1/meet/join", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Should return 501 until service layer is wired (Phase 3)
	if w.Code != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotImplemented)
	}
}
