package dto

import "github.com/hnam/notafly/internal/model"

type JoinMeetResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	AudioPath string `json:"audio_path"`
}

type TranscribeResponse struct {
	Transcription string               `json:"transcription"`
	Minutes       *model.MeetingMinutes `json:"minutes,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
