package dto

import (
	"encoding/json"
	"testing"

	"github.com/hnam/notafly/internal/model"
)

func TestJoinMeetRequest_JSON(t *testing.T) {
	raw := `{"meet_link":"https://meet.google.com/abc","duration":120}`
	var req JoinMeetRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if req.MeetLink != "https://meet.google.com/abc" {
		t.Errorf("MeetLink = %q", req.MeetLink)
	}
	if req.Duration != 120 {
		t.Errorf("Duration = %d, want 120", req.Duration)
	}
}

func TestTranscribeRequest_JSON(t *testing.T) {
	raw := `{"audio_path":"/tmp/audio.wav","no_analysis":true}`
	var req TranscribeRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if req.AudioPath != "/tmp/audio.wav" {
		t.Errorf("AudioPath = %q", req.AudioPath)
	}
	if !req.NoAnalysis {
		t.Error("NoAnalysis should be true")
	}
}

func TestErrorResponse_OmitsEmptyDetails(t *testing.T) {
	resp := ErrorResponse{Error: "something failed"}
	data, _ := json.Marshal(resp)

	var decoded map[string]interface{}
	json.Unmarshal(data, &decoded)

	if _, ok := decoded["details"]; ok {
		t.Error("details should be omitted when empty")
	}
}

func TestTranscribeResponse_OmitsNilMinutes(t *testing.T) {
	resp := TranscribeResponse{Transcription: "hello world"}
	data, _ := json.Marshal(resp)

	var decoded map[string]interface{}
	json.Unmarshal(data, &decoded)

	if _, ok := decoded["minutes"]; ok {
		t.Error("minutes should be omitted when nil")
	}
}

func TestTranscribeResponse_IncludesMinutes(t *testing.T) {
	resp := TranscribeResponse{
		Transcription: "hello",
		Minutes:       &model.MeetingMinutes{Sentiment: "positive"},
	}
	data, _ := json.Marshal(resp)

	var decoded map[string]interface{}
	json.Unmarshal(data, &decoded)

	if _, ok := decoded["minutes"]; !ok {
		t.Error("minutes should be present when set")
	}
}
