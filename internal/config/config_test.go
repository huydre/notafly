package config

import (
	"os"
	"testing"
)

func TestLoad_WithRequiredEnvVars(t *testing.T) {
	os.Setenv("EMAIL_ID", "test@gmail.com")
	os.Setenv("EMAIL_PASSWORD", "pass123")
	os.Setenv("OPENAI_API_KEY", "sk-test")
	defer func() {
		os.Unsetenv("EMAIL_ID")
		os.Unsetenv("EMAIL_PASSWORD")
		os.Unsetenv("OPENAI_API_KEY")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.EmailID != "test@gmail.com" {
		t.Errorf("EmailID = %q, want %q", cfg.EmailID, "test@gmail.com")
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}
	if cfg.RecordingDuration != 60 {
		t.Errorf("RecordingDuration = %d, want 60", cfg.RecordingDuration)
	}
	if cfg.SampleRate != 44100 {
		t.Errorf("SampleRate = %d, want 44100", cfg.SampleRate)
	}
	if cfg.GPTModel != "gpt-4" {
		t.Errorf("GPTModel = %q, want %q", cfg.GPTModel, "gpt-4")
	}
}

func TestLoad_MissingRequired(t *testing.T) {
	os.Unsetenv("EMAIL_ID")
	os.Unsetenv("EMAIL_PASSWORD")
	os.Unsetenv("OPENAI_API_KEY")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing required vars, got nil")
	}
}

func TestLoad_CustomValues(t *testing.T) {
	os.Setenv("EMAIL_ID", "user@test.com")
	os.Setenv("EMAIL_PASSWORD", "secret")
	os.Setenv("OPENAI_API_KEY", "sk-custom")
	os.Setenv("PORT", "3000")
	os.Setenv("RECORDING_DURATION", "120")
	os.Setenv("GPT_MODEL", "gpt-4o")
	defer func() {
		os.Unsetenv("EMAIL_ID")
		os.Unsetenv("EMAIL_PASSWORD")
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("PORT")
		os.Unsetenv("RECORDING_DURATION")
		os.Unsetenv("GPT_MODEL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "3000" {
		t.Errorf("Port = %q, want %q", cfg.Port, "3000")
	}
	if cfg.RecordingDuration != 120 {
		t.Errorf("RecordingDuration = %d, want 120", cfg.RecordingDuration)
	}
	if cfg.GPTModel != "gpt-4o" {
		t.Errorf("GPTModel = %q, want %q", cfg.GPTModel, "gpt-4o")
	}
}
