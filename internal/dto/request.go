package dto

type JoinMeetRequest struct {
	MeetLink string `json:"meet_link" binding:"required,url"`
	Duration int    `json:"duration" binding:"required,min=10,max=7200"`
}

type TranscribeRequest struct {
	AudioPath  string `json:"audio_path" binding:"required"`
	NoAnalysis bool   `json:"no_analysis"`
}
