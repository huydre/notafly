package service

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hnam/notafly/internal/config"
	"go.uber.org/zap"
)

func TestNewRecorderService(t *testing.T) {
	cfg := &config.Config{SampleRate: 44100, MaxAudioSizeBytes: 20 * 1024 * 1024}
	logger := zap.NewNop()
	svc := NewRecorderService(cfg, logger)

	if svc == nil {
		t.Fatal("expected non-nil RecorderService")
	}
	if svc.config.SampleRate != 44100 {
		t.Errorf("SampleRate = %d, want 44100", svc.config.SampleRate)
	}
}

func TestAudioSource(t *testing.T) {
	format, device := audioSource()

	switch runtime.GOOS {
	case "darwin":
		if format != "avfoundation" {
			t.Errorf("format = %q, want avfoundation", format)
		}
		if device != ":0" {
			t.Errorf("device = %q, want :0", device)
		}
	case "linux":
		if format != "pulse" {
			t.Errorf("format = %q, want pulse", format)
		}
		if device != "default" {
			t.Errorf("device = %q, want default", device)
		}
	}
}

func TestCheckFFmpeg(t *testing.T) {
	// Only run if ffmpeg is actually installed
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed, skipping")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not installed, skipping")
	}

	if err := CheckFFmpeg(); err != nil {
		t.Errorf("CheckFFmpeg() = %v, want nil", err)
	}
}

func TestGetAudioDuration(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}

	// Generate a 2-second silent WAV file
	tmpDir := t.TempDir()
	wavPath := filepath.Join(tmpDir, "test.wav")
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=stereo",
		"-t", "2", "-y", wavPath,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create test WAV: %v", err)
	}

	cfg := &config.Config{SampleRate: 44100, MaxAudioSizeBytes: 20 * 1024 * 1024}
	svc := NewRecorderService(cfg, zap.NewNop())

	dur, err := svc.GetAudioDuration(wavPath)
	if err != nil {
		t.Fatalf("GetAudioDuration() error: %v", err)
	}

	// Should be ~2 seconds (allow margin)
	if dur < 1.5 || dur > 2.5 {
		t.Errorf("duration = %.2f, want ~2.0", dur)
	}
}

func TestCompressIfNeeded_NoCompressionNeeded(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}

	tmpDir := t.TempDir()
	wavPath := filepath.Join(tmpDir, "small.wav")
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=stereo",
		"-t", "1", "-y", wavPath,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create test WAV: %v", err)
	}

	// Set max size very high — no compression should happen
	cfg := &config.Config{SampleRate: 44100, MaxAudioSizeBytes: 100 * 1024 * 1024}
	svc := NewRecorderService(cfg, zap.NewNop())

	result, err := svc.CompressIfNeeded(wavPath)
	if err != nil {
		t.Fatalf("CompressIfNeeded() error: %v", err)
	}

	// Should return original path
	if result != wavPath {
		t.Errorf("result = %q, want original %q", result, wavPath)
	}
}

func TestCompressIfNeeded_CompressesLargeFile(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}

	tmpDir := t.TempDir()
	wavPath := filepath.Join(tmpDir, "large.wav")
	// Generate a 3-second file
	cmd := exec.Command("ffmpeg",
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=stereo",
		"-t", "3", "-y", wavPath,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create test WAV: %v", err)
	}

	info, _ := os.Stat(wavPath)

	// Set max size to half the file size — should trigger compression
	cfg := &config.Config{
		SampleRate:        44100,
		MaxAudioSizeBytes: int(info.Size() / 2),
	}
	svc := NewRecorderService(cfg, zap.NewNop())

	result, err := svc.CompressIfNeeded(wavPath)
	if err != nil {
		t.Fatalf("CompressIfNeeded() error: %v", err)
	}

	// Should return a different (compressed) path
	if result == wavPath {
		t.Error("expected compressed path, got original")
	}

	// Compressed file should exist
	if _, err := os.Stat(result); os.IsNotExist(err) {
		t.Errorf("compressed file does not exist: %s", result)
	}

	// Cleanup compressed file
	os.Remove(result)
}

func TestCompressIfNeeded_NonexistentFile(t *testing.T) {
	cfg := &config.Config{MaxAudioSizeBytes: 1024}
	svc := NewRecorderService(cfg, zap.NewNop())

	_, err := svc.CompressIfNeeded("/nonexistent/file.wav")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
