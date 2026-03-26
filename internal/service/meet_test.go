package service

import (
	"testing"

	"github.com/hnam/notafly/internal/config"
	"go.uber.org/zap"
)

func TestNewMeetService(t *testing.T) {
	cfg := &config.Config{
		EmailID:       "test@gmail.com",
		EmailPassword: "pass",
	}
	logger := zap.NewNop()

	svc := NewMeetService(cfg, logger)

	if svc == nil {
		t.Fatal("expected non-nil MeetService")
	}
	if svc.config.EmailID != "test@gmail.com" {
		t.Errorf("EmailID = %q, want %q", svc.config.EmailID, "test@gmail.com")
	}
}

func TestSelectors_NotEmpty(t *testing.T) {
	// Verify all selectors are set — catches accidental empty strings
	checks := []struct {
		name  string
		value string
	}{
		{"EmailInput", selectors.EmailInput},
		{"EmailNext", selectors.EmailNext},
		{"PasswordInput", selectors.PasswordInput},
		{"PasswordNext", selectors.PasswordNext},
		{"MicToggle", selectors.MicToggle},
		{"CameraToggle", selectors.CameraToggle},
		{"JoinButton", selectors.JoinButton},
	}
	for _, c := range checks {
		if c.value == "" {
			t.Errorf("selector %s is empty", c.name)
		}
	}
}

func TestLoginURL(t *testing.T) {
	if loginURL == "" {
		t.Fatal("loginURL is empty")
	}
	if loginURL[:8] != "https://" {
		t.Errorf("loginURL should start with https://, got %q", loginURL[:8])
	}
}
