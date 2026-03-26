package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/hnam/notafly/internal/config"
	"go.uber.org/zap"
)

type RecorderService struct {
	config *config.Config
	logger *zap.Logger
}

func NewRecorderService(cfg *config.Config, logger *zap.Logger) *RecorderService {
	return &RecorderService{config: cfg, logger: logger}
}

// Record captures system audio to a WAV file for the given duration.
// Uses ffmpeg with platform-specific audio input (avfoundation on macOS, pulse on Linux).
func (s *RecorderService) Record(ctx context.Context, outputPath string, duration int) error {
	inputFormat, inputDevice := audioSource()

	args := []string{
		"-f", inputFormat,
		"-i", inputDevice,
		"-t", strconv.Itoa(duration),
		"-ar", strconv.Itoa(s.config.SampleRate),
		"-ac", "2",
		"-sample_fmt", "s16",
		"-y", // overwrite output
		outputPath,
	}

	s.logger.Info("starting audio recording",
		zap.String("output", outputPath),
		zap.Int("duration", duration),
		zap.String("format", inputFormat),
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stderr = os.Stderr // show ffmpeg progress

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg recording failed: %w", err)
	}

	s.logger.Info("recording finished", zap.String("output", outputPath))
	return nil
}

// CompressIfNeeded checks the file size and compresses if it exceeds MaxAudioSizeBytes.
// Returns the (possibly new) path to the audio file.
func (s *RecorderService) CompressIfNeeded(audioPath string) (string, error) {
	info, err := os.Stat(audioPath)
	if err != nil {
		return "", fmt.Errorf("stat audio file: %w", err)
	}

	maxSize := int64(s.config.MaxAudioSizeBytes)
	if info.Size() <= maxSize {
		s.logger.Debug("audio within size limit, no compression needed",
			zap.Int64("size", info.Size()),
			zap.Int64("max", maxSize),
		)
		return audioPath, nil
	}

	s.logger.Info("compressing audio",
		zap.Int64("original_size", info.Size()),
		zap.Int64("max_size", maxSize),
	)

	duration, err := s.GetAudioDuration(audioPath)
	if err != nil {
		return "", fmt.Errorf("get audio duration: %w", err)
	}

	targetDuration := duration * float64(maxSize) / float64(info.Size())

	compressed := filepath.Join(os.TempDir(),
		fmt.Sprintf("notafly_compressed_%d.wav", time.Now().UnixNano()))

	cmd := exec.Command("ffmpeg",
		"-i", audioPath,
		"-ss", "0",
		"-t", fmt.Sprintf("%.2f", targetDuration),
		"-y",
		compressed,
	)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg compression failed: %w", err)
	}

	s.logger.Info("compression done",
		zap.String("path", compressed),
		zap.Float64("target_duration", targetDuration),
	)
	return compressed, nil
}

// GetAudioDuration returns the duration of an audio file in seconds using ffprobe.
func (s *RecorderService) GetAudioDuration(audioPath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-i", audioPath,
		"-show_entries", "format=duration",
		"-v", "quiet",
		"-of", "csv=p=0",
	)

	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w", err)
	}

	durationStr := strings.TrimSpace(string(out))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration %q: %w", durationStr, err)
	}

	return duration, nil
}

// CheckFFmpeg verifies that ffmpeg and ffprobe are installed and accessible.
func CheckFFmpeg() error {
	for _, bin := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(bin); err != nil {
			return fmt.Errorf("%s not found in PATH: %w", bin, err)
		}
	}
	return nil
}

// audioSource returns the ffmpeg input format and device for the current platform.
func audioSource() (format string, device string) {
	switch runtime.GOOS {
	case "darwin":
		return "avfoundation", ":0" // default audio input on macOS
	case "linux":
		return "pulse", "default" // PulseAudio on Linux
	default:
		return "dshow", "audio=default" // DirectShow on Windows
	}
}
