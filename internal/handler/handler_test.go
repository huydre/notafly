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
	"github.com/hnam/notafly/internal/middleware"
	"github.com/hnam/notafly/internal/service"
	"go.uber.org/zap"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	logger := zap.NewNop()
	meetSvc := service.NewMeetService(cfg, logger)
	recorderSvc := service.NewRecorderService(cfg, logger)
	transcriberSvc := service.NewTranscriberService(cfg, logger)
	h := New(cfg, logger, meetSvc, recorderSvc, transcriberSvc)

	r := gin.New()
	r.Use(middleware.JSONRecovery(logger))
	r.Use(middleware.CORS())

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

// --- Health ---

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

// --- JoinMeet ---

func TestJoinMeet_InvalidURL(t *testing.T) {
	router := setupTestRouter()
	w := httptest.NewRecorder()
	body := `{"meet_link": "not-a-url", "duration": 60}`
	req, _ := http.NewRequest("POST", "/api/v1/meet/join", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
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

func TestJoinMeet_DurationTooSmall(t *testing.T) {
	router := setupTestRouter()
	w := httptest.NewRecorder()
	body := `{"meet_link": "https://meet.google.com/abc-defg-hij", "duration": 5}`
	req, _ := http.NewRequest("POST", "/api/v1/meet/join", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestJoinMeet_DurationTooLarge(t *testing.T) {
	router := setupTestRouter()
	w := httptest.NewRecorder()
	body := `{"meet_link": "https://meet.google.com/abc-defg-hij", "duration": 9999}`
	req, _ := http.NewRequest("POST", "/api/v1/meet/join", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestJoinMeet_InvalidJSON(t *testing.T) {
	router := setupTestRouter()
	w := httptest.NewRecorder()
	body := `{invalid`
	req, _ := http.NewRequest("POST", "/api/v1/meet/join", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// --- FullPipeline ---

func TestFullPipeline_InvalidRequest(t *testing.T) {
	router := setupTestRouter()
	w := httptest.NewRecorder()
	body := `{}`
	req, _ := http.NewRequest("POST", "/api/v1/meet/full", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// --- Transcribe ---

func TestTranscribe_MissingAudioPath(t *testing.T) {
	router := setupTestRouter()
	w := httptest.NewRecorder()
	body := `{}`
	req, _ := http.NewRequest("POST", "/api/v1/transcribe", strings.NewReader(body))
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

func TestTranscribe_InvalidJSON(t *testing.T) {
	router := setupTestRouter()
	w := httptest.NewRecorder()
	body := `not-json`
	req, _ := http.NewRequest("POST", "/api/v1/transcribe", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// --- CORS ---

func TestCORS_OptionsRequest(t *testing.T) {
	router := setupTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/v1/meet/join", nil)
	router.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Errorf("status = %d, want 204", w.Code)
	}
	if h := w.Header().Get("Access-Control-Allow-Origin"); h != "*" {
		t.Errorf("CORS origin = %q, want *", h)
	}
	if h := w.Header().Get("Access-Control-Allow-Methods"); h == "" {
		t.Error("CORS methods header missing")
	}
}

// --- 404 ---

func TestNotFound(t *testing.T) {
	router := setupTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/nonexistent", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// --- Method not allowed ---

func TestMethodNotAllowed(t *testing.T) {
	router := setupTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/meet/join", nil)
	router.ServeHTTP(w, req)

	// Gin returns 404 for wrong method on non-configured routes
	if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 404 or 405", w.Code)
	}
}
