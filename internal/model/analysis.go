package model

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
