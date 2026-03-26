# Phase 6: Gin HTTP API

**Priority:** P1
**Status:** ⬜ Not started

---

## Context

This is the **new** part — Python version has no HTTP API. Gin enables web-based control of the meeting bot.

## API Design

### Endpoints

```
GET  /health                    → 200 OK
POST /api/v1/meet/join          → Start meeting session (join + record)
POST /api/v1/transcribe         → Transcribe audio file
POST /api/v1/meet/full          → Full pipeline (join + record + transcribe + analyze)
GET  /api/v1/sessions/:id       → Get session status (future: with DB)
```

### Route Setup

```go
// cmd/server/main.go
func setupRouter(h *handler.Handler) *gin.Engine {
    r := gin.Default()

    // Middleware
    r.Use(middleware.CORS())
    r.Use(middleware.RequestLogger())
    r.Use(gin.Recovery())

    // Health
    r.GET("/health", h.Health)

    // API v1
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
```

### Handler Examples

```go
// internal/handler/meet.go
type Handler struct {
    meetSvc        *service.MeetService
    recorderSvc    *service.RecorderService
    transcriberSvc *service.TranscriberService
    logger         *zap.Logger
}

func (h *Handler) JoinMeet(c *gin.Context) {
    var req dto.JoinMeetRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, dto.ErrorResponse{Error: "invalid request", Details: err.Error()})
        return
    }

    session, err := h.meetSvc.JoinAndRecord(c.Request.Context(), req.MeetLink, req.Duration)
    if err != nil {
        c.JSON(500, dto.ErrorResponse{Error: "failed to join meeting", Details: err.Error()})
        return
    }

    c.JSON(200, dto.JoinMeetResponse{
        SessionID: session.ID,
        Status:    string(session.Status),
        AudioPath: session.AudioPath,
    })
}
```

## Middleware

| Middleware | Purpose |
|-----------|---------|
| `CORS()` | Allow cross-origin requests |
| `RequestLogger()` | Structured request logging (zap) |
| `Recovery()` | Panic recovery (built-in Gin) |
| `Auth()` (optional) | JWT auth for API protection |

## Implementation Steps

- [ ] Create `cmd/server/main.go` with Gin setup
- [ ] Create `internal/handler/meet.go`
- [ ] Create `internal/handler/transcribe.go`
- [ ] Create `internal/handler/health.go`
- [ ] Wire up dependency injection in main.go
- [ ] Add CORS middleware
- [ ] Add request logger middleware
- [ ] Add graceful shutdown (`signal.NotifyContext`)

## Success Criteria

- All endpoints respond correctly
- Request validation works (400 on bad input)
- Graceful shutdown on SIGINT/SIGTERM
- Structured JSON logging
