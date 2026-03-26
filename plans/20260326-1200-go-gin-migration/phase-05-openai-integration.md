# Phase 5: OpenAI Integration

**Priority:** P0
**Status:** ✅ Completed

---

## Context

Python uses `openai` SDK for:
1. **Whisper** — audio transcription (`client.audio.translations.create`)
2. **GPT-4** — 4 parallel analysis tasks (summary, key points, actions, sentiment)

## Go Implementation

```go
// internal/service/transcriber.go
type TranscriberService struct {
    client *openai.Client
    config *config.Config
    logger *zap.Logger
}

func NewTranscriberService(cfg *config.Config) *TranscriberService {
    client := openai.NewClient(cfg.OpenAIAPIKey)
    return &TranscriberService{client: client, config: cfg}
}

// Transcribe audio file via Whisper API
func (s *TranscriberService) Transcribe(ctx context.Context, audioPath string) (string, error) {
    req := openai.AudioRequest{
        Model:    s.config.WhisperModel,
        FilePath: audioPath,
    }
    resp, err := s.client.CreateTranslation(ctx, req)
    if err != nil {
        return "", fmt.Errorf("whisper transcription failed: %w", err)
    }
    return resp.Text, nil
}

// AnalyzeMeeting runs 4 GPT analyses CONCURRENTLY (improvement over Python)
func (s *TranscriberService) AnalyzeMeeting(ctx context.Context, transcription string) (*model.MeetingMinutes, error) {
    var (
        wg      sync.WaitGroup
        mu      sync.Mutex
        minutes model.MeetingMinutes
        errs    []error
    )

    analyses := []struct {
        prompt string
        target *string
    }{
        {summaryPrompt, &minutes.AbstractSummary},
        {keyPointsPrompt, &minutes.KeyPoints},
        {actionItemsPrompt, &minutes.ActionItems},
        {sentimentPrompt, &minutes.Sentiment},
    }

    for _, a := range analyses {
        wg.Add(1)
        go func(prompt string, target *string) {
            defer wg.Done()
            result, err := s.chatCompletion(ctx, prompt, transcription)
            mu.Lock()
            defer mu.Unlock()
            if err != nil {
                errs = append(errs, err)
                return
            }
            *target = result
        }(a.prompt, a.target)
    }

    wg.Wait()
    if len(errs) > 0 {
        return nil, fmt.Errorf("analysis errors: %v", errs)
    }
    return &minutes, nil
}
```

## Key Improvement: Concurrent Analysis

Python runs 4 GPT calls **sequentially** (~20-40s total). Go version uses goroutines to run all 4 **concurrently** (~5-10s total). 4x speedup on analysis phase.

## System Prompts (ported from Python)

Same prompts used in `speech_to_text.py` lines 59-123. Store as constants in the service file.

## Python → Go Mapping

| Python | Go |
|--------|----|
| `OpenAI(api_key=...)` | `openai.NewClient(apiKey)` |
| `client.audio.translations.create()` | `client.CreateTranslation(ctx, req)` |
| `client.chat.completions.create()` | `client.CreateChatCompletion(ctx, req)` |
| `response.choices[0].message.content` | `resp.Choices[0].Message.Content` |
| Sequential calls | `sync.WaitGroup` + goroutines |
| `json.dump(data, f)` | `json.NewEncoder(f).Encode(data)` |

## Implementation Steps

- [ ] Create `internal/service/transcriber.go`
- [ ] Implement `Transcribe()` — Whisper API call
- [ ] Implement `AnalyzeMeeting()` — concurrent GPT-4 calls
- [ ] Implement `chatCompletion()` — reusable helper
- [ ] Port all 4 system prompts as constants
- [ ] Add `SaveToJSON()` for file output
- [ ] Handle large file upload (multipart form)

## Success Criteria

- Audio file transcribed via Whisper API
- 4 analyses run concurrently
- Results match Python output format
- JSON output saved correctly
