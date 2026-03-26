# Phase 2: Core Models & DTOs

**Priority:** P0
**Status:** ⬜ Not started

---

## Context

Python version uses dicts and loosely typed returns. Go version needs explicit structs.

## Domain Models

```go
// internal/model/meeting.go
type MeetingSession struct {
    ID            string        `json:"id"`
    MeetLink      string        `json:"meet_link"`
    AudioPath     string        `json:"audio_path"`
    Duration      int           `json:"duration"`
    Status        SessionStatus `json:"status"`
    CreatedAt     time.Time     `json:"created_at"`
}

type SessionStatus string
const (
    StatusPending      SessionStatus = "pending"
    StatusJoining      SessionStatus = "joining"
    StatusRecording    SessionStatus = "recording"
    StatusTranscribing SessionStatus = "transcribing"
    StatusCompleted    SessionStatus = "completed"
    StatusFailed       SessionStatus = "failed"
)

// internal/model/analysis.go
type MeetingMinutes struct {
    AbstractSummary string `json:"abstract_summary"`
    KeyPoints       string `json:"key_points"`
    ActionItems     string `json:"action_items"`
    Sentiment       string `json:"sentiment"`
}

type TranscriptionResult struct {
    Text    string         `json:"text"`
    Minutes MeetingMinutes `json:"minutes"`
}
```

## Request/Response DTOs

```go
// internal/dto/request.go
type JoinMeetRequest struct {
    MeetLink string `json:"meet_link" binding:"required,url"`
    Duration int    `json:"duration" binding:"required,min=10,max=7200"`
}

type TranscribeRequest struct {
    AudioPath  string `json:"audio_path" binding:"required"`
    NoAnalysis bool   `json:"no_analysis"`
}

// internal/dto/response.go
type JoinMeetResponse struct {
    SessionID string `json:"session_id"`
    Status    string `json:"status"`
    AudioPath string `json:"audio_path"`
}

type TranscribeResponse struct {
    Transcription string                `json:"transcription"`
    Minutes       *model.MeetingMinutes `json:"minutes,omitempty"`
}

type ErrorResponse struct {
    Error   string `json:"error"`
    Details string `json:"details,omitempty"`
}
```

## Python → Go Type Mapping

| Python | Go |
|--------|----|
| `str` | `string` |
| `int` (RECORDING_DURATION) | `int` |
| `dict` (meeting_minutes return) | `MeetingMinutes` struct |
| `os.getenv()` | `Config` struct field |
| `None` | pointer type or zero value |

## Implementation Steps

- [ ] Create `internal/model/meeting.go`
- [ ] Create `internal/model/analysis.go`
- [ ] Create `internal/dto/request.go`
- [ ] Create `internal/dto/response.go`
- [ ] Add JSON tags and Gin binding validators

## Success Criteria

- All structs compile
- Binding tags cover required fields
- JSON serialization matches expected API format
