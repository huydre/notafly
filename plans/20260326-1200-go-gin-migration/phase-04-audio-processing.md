# Phase 4: Audio Recording & Processing

**Priority:** P0
**Status:** ⬜ Not started

---

## Context

Python uses `sounddevice` (PortAudio binding) to record system audio, `scipy` to write WAV. Falls back to `ffmpeg` for compression.

## Go Approach

**Decision: Use ffmpeg for both recording and processing.** No good native Go audio recording lib. Python's `sounddevice` also wraps C (PortAudio). Simpler to shell out to ffmpeg which is already a dependency.

```go
// internal/service/recorder.go
type RecorderService struct {
    config *config.Config
    logger *zap.Logger
}

func (s *RecorderService) Record(ctx context.Context, outputPath string, duration int) error {
    // Use ffmpeg to capture system audio
    // macOS: uses avfoundation
    // Linux: uses pulse/alsa
    args := []string{
        "-f", getAudioInput(),    // platform-specific
        "-i", getAudioDevice(),   // e.g., ":0" on macOS
        "-t", strconv.Itoa(duration),
        "-ar", strconv.Itoa(s.config.SampleRate),
        "-ac", "2",
        "-sample_fmt", "s16",
        outputPath,
    }

    cmd := exec.CommandContext(ctx, "ffmpeg", args...)
    return cmd.Run()
}

func (s *RecorderService) CompressIfNeeded(audioPath string) (string, error) {
    info, _ := os.Stat(audioPath)
    if info.Size() <= int64(s.config.MaxAudioSizeBytes) {
        return audioPath, nil
    }

    // Get duration, calculate target
    duration := s.getAudioDuration(audioPath)
    targetDuration := duration * float64(s.config.MaxAudioSizeBytes) / float64(info.Size())

    compressed := filepath.Join(os.TempDir(), fmt.Sprintf("compressed_%d.wav", time.Now().Unix()))
    cmd := exec.Command("ffmpeg", "-i", audioPath, "-ss", "0", "-t",
        fmt.Sprintf("%.2f", targetDuration), compressed)
    return compressed, cmd.Run()
}
```

## Python → Go Mapping

| Python | Go |
|--------|----|
| `sd.rec(frames, samplerate, channels, dtype)` | `exec.Command("ffmpeg", "-f", "avfoundation", ...)` |
| `sd.wait()` | `cmd.Run()` blocks until done |
| `scipy.io.wavfile.write()` | ffmpeg outputs WAV directly |
| `subprocess.run(['ffprobe', ...])` | `exec.Command("ffprobe", ...)` |
| `subprocess.run(['ffmpeg', ...])` | `exec.Command("ffmpeg", ...)` |

## Implementation Steps

- [ ] Create `internal/service/recorder.go`
- [ ] Implement `Record()` with platform detection (macOS avfoundation vs Linux pulse)
- [ ] Implement `CompressIfNeeded()` — same logic as Python's `resize_audio_if_needed`
- [ ] Implement `GetAudioDuration()` via ffprobe
- [ ] Add context cancellation support (stop recording early)
- [ ] Temp file cleanup on completion/error

## Success Criteria

- Records audio to WAV for specified duration
- Compresses audio if exceeds max size
- Works on macOS (avfoundation) and Linux (pulseaudio)
- Proper cleanup of temp files
