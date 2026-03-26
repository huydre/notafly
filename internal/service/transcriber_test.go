package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hnam/notafly/internal/config"
	"github.com/hnam/notafly/internal/model"
	"go.uber.org/zap"
)

func TestNewTranscriberService(t *testing.T) {
	cfg := &config.Config{
		OpenAIAPIKey: "sk-test",
		GPTModel:     "gpt-4",
		WhisperModel: "whisper-1",
	}
	svc := NewTranscriberService(cfg, zap.NewNop())

	if svc == nil {
		t.Fatal("expected non-nil TranscriberService")
	}
	if svc.client == nil {
		t.Fatal("expected non-nil OpenAI client")
	}
	if svc.config.GPTModel != "gpt-4" {
		t.Errorf("GPTModel = %q, want %q", svc.config.GPTModel, "gpt-4")
	}
}

func TestPrompts_AllPresent(t *testing.T) {
	prompts := Prompts()
	required := []string{"summary", "key_points", "action_items", "sentiment"}

	for _, key := range required {
		val, ok := prompts[key]
		if !ok {
			t.Errorf("missing prompt: %s", key)
			continue
		}
		if len(val) < 50 {
			t.Errorf("prompt %q too short (%d chars)", key, len(val))
		}
	}

	if len(prompts) != 4 {
		t.Errorf("expected 4 prompts, got %d", len(prompts))
	}
}

func TestPrompts_MatchPython(t *testing.T) {
	// Verify key phrases from the original Python prompts are preserved
	prompts := Prompts()

	checks := map[string]string{
		"summary":      "concise abstract paragraph",
		"key_points":   "distilling information into key points",
		"action_items": "extracting action items",
		"sentiment":    "language and emotion analysis",
	}

	for key, phrase := range checks {
		if p, ok := prompts[key]; ok {
			found := false
			if len(p) > 0 {
				for i := 0; i <= len(p)-len(phrase); i++ {
					if p[i:i+len(phrase)] == phrase {
						found = true
						break
					}
				}
			}
			if !found {
				t.Errorf("prompt %q missing phrase %q", key, phrase)
			}
		}
	}
}

func TestSaveToJSON(t *testing.T) {
	cfg := &config.Config{OpenAIAPIKey: "sk-test"}
	svc := NewTranscriberService(cfg, zap.NewNop())

	result := &model.TranscriptionResult{
		Text: "Hello everyone, welcome to the meeting.",
		Minutes: model.MeetingMinutes{
			AbstractSummary: "Team discussed quarterly goals.",
			KeyPoints:       "- Revenue\n- Hiring",
			ActionItems:     "- Bob: budget\n- Alice: roadmap",
			Sentiment:       "positive",
		},
	}

	err := svc.SaveToJSON(result)
	if err != nil {
		t.Fatalf("SaveToJSON() error: %v", err)
	}

	// Find the file just created
	files, _ := filepath.Glob(filepath.Join(os.TempDir(), "meeting_data_*.json"))
	if len(files) == 0 {
		t.Fatal("no meeting_data JSON file found")
	}

	// Read the latest one
	latestFile := files[len(files)-1]
	data, err := os.ReadFile(latestFile)
	if err != nil {
		t.Fatalf("read JSON: %v", err)
	}

	var decoded model.TranscriptionResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}

	if decoded.Text != result.Text {
		t.Errorf("Text = %q, want %q", decoded.Text, result.Text)
	}
	if decoded.Minutes.Sentiment != "positive" {
		t.Errorf("Sentiment = %q, want %q", decoded.Minutes.Sentiment, "positive")
	}

	// Cleanup
	os.Remove(latestFile)
}

func TestSaveToJSON_OutputFormat(t *testing.T) {
	cfg := &config.Config{OpenAIAPIKey: "sk-test"}
	svc := NewTranscriberService(cfg, zap.NewNop())

	result := &model.TranscriptionResult{
		Text: "test",
		Minutes: model.MeetingMinutes{
			AbstractSummary: "s",
			KeyPoints:       "k",
			ActionItems:     "a",
			Sentiment:       "n",
		},
	}

	svc.SaveToJSON(result)

	files, _ := filepath.Glob(filepath.Join(os.TempDir(), "meeting_data_*.json"))
	if len(files) == 0 {
		t.Fatal("no file")
	}

	data, _ := os.ReadFile(files[len(files)-1])

	// Verify all expected JSON keys
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	if _, ok := raw["text"]; !ok {
		t.Error("missing 'text' key")
	}
	if minutes, ok := raw["minutes"].(map[string]interface{}); ok {
		for _, key := range []string{"abstract_summary", "key_points", "action_items", "sentiment"} {
			if _, exists := minutes[key]; !exists {
				t.Errorf("missing minutes.%s key", key)
			}
		}
	} else {
		t.Error("missing or invalid 'minutes' key")
	}

	os.Remove(files[len(files)-1])
}
