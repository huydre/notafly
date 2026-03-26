package model

import (
	"encoding/json"
	"testing"
)

func TestMeetingMinutes_JSON(t *testing.T) {
	m := MeetingMinutes{
		AbstractSummary: "Team discussed Q2 goals",
		KeyPoints:       "- Revenue target\n- Hiring plan",
		ActionItems:     "- Bob: draft budget\n- Alice: review roadmap",
		Sentiment:       "positive",
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded MeetingMinutes
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.AbstractSummary != m.AbstractSummary {
		t.Errorf("AbstractSummary mismatch")
	}
	if decoded.Sentiment != "positive" {
		t.Errorf("Sentiment = %q, want %q", decoded.Sentiment, "positive")
	}
}

func TestTranscriptionResult_JSON(t *testing.T) {
	r := TranscriptionResult{
		Text: "Hello everyone, welcome to the meeting.",
		Minutes: MeetingMinutes{
			AbstractSummary: "Summary",
			KeyPoints:       "Points",
			ActionItems:     "Actions",
			Sentiment:       "neutral",
		},
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded TranscriptionResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.Text != r.Text {
		t.Errorf("Text mismatch")
	}
	if decoded.Minutes.Sentiment != "neutral" {
		t.Errorf("Minutes.Sentiment = %q, want %q", decoded.Minutes.Sentiment, "neutral")
	}
}

func TestMeetingMinutes_EmptyFields(t *testing.T) {
	m := MeetingMinutes{}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Empty strings should still serialize
	var decoded map[string]string
	json.Unmarshal(data, &decoded)

	keys := []string{"abstract_summary", "key_points", "action_items", "sentiment"}
	for _, k := range keys {
		if _, ok := decoded[k]; !ok {
			t.Errorf("missing key %q in JSON output", k)
		}
	}
}
