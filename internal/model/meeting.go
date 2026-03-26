package model

import "time"

type SessionStatus string

const (
	StatusPending      SessionStatus = "pending"
	StatusJoining      SessionStatus = "joining"
	StatusRecording    SessionStatus = "recording"
	StatusTranscribing SessionStatus = "transcribing"
	StatusCompleted    SessionStatus = "completed"
	StatusFailed       SessionStatus = "failed"
)

type MeetingSession struct {
	ID        string        `json:"id"`
	MeetLink  string        `json:"meet_link"`
	AudioPath string        `json:"audio_path"`
	Duration  int           `json:"duration"`
	Status    SessionStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
}
