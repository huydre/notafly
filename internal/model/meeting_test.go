package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSessionStatus_Values(t *testing.T) {
	statuses := []SessionStatus{
		StatusPending, StatusJoining, StatusRecording,
		StatusTranscribing, StatusCompleted, StatusFailed,
	}
	expected := []string{
		"pending", "joining", "recording",
		"transcribing", "completed", "failed",
	}
	for i, s := range statuses {
		if string(s) != expected[i] {
			t.Errorf("status[%d] = %q, want %q", i, s, expected[i])
		}
	}
}

func TestMeetingSession_JSON(t *testing.T) {
	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)
	s := MeetingSession{
		ID:        "abc-123",
		MeetLink:  "https://meet.google.com/xxx",
		AudioPath: "/tmp/output.wav",
		Duration:  60,
		Status:    StatusPending,
		CreatedAt: now,
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded MeetingSession
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.ID != s.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, s.ID)
	}
	if decoded.Status != StatusPending {
		t.Errorf("Status = %q, want %q", decoded.Status, StatusPending)
	}
	if decoded.Duration != 60 {
		t.Errorf("Duration = %d, want 60", decoded.Duration)
	}
}
