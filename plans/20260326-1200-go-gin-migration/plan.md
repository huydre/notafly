# Google Meet Bot → Go (Gin) Migration Plan

**Date:** 2026-03-26
**Status:** Planning
**Complexity:** Medium-High

---

## Current State

Python CLI tool moved to `old/` for reference. Go code lives at project root.
Original: automates Google Meet joining, audio recording, Whisper transcription, and GPT-4 analysis. CLI-only, no database, no HTTP endpoints.

| Component | Python | Go Equivalent |
|-----------|--------|---------------|
| CLI entry | `argparse` + `cli.py` | `cobra` or `flag` |
| Browser automation | `selenium` | `chromedp` (headless Chrome via DevTools Protocol) |
| Audio recording | `sounddevice` + `scipy` | `portaudio` bindings or exec `ffmpeg` |
| Audio processing | `ffmpeg` subprocess | `exec.Command("ffmpeg", ...)` |
| OpenAI API | `openai` Python SDK | `sashabaranov/go-openai` |
| Config | `python-dotenv` | `joho/godotenv` or `viper` |
| HTTP server (NEW) | N/A | `gin-gonic/gin` |

## Architecture Decision

Since user wants **Gin** (web framework), the migration transforms this from CLI → **Web API service** with optional CLI mode.

---

## Phases

| # | Phase | Status | Details |
|---|-------|--------|---------|
| 1 | [Project Setup & Config](phase-01-project-setup.md) | ✅ | Go module, project structure, config loading |
| 2 | [Core Models & DTOs](phase-02-models-dtos.md) | ✅ | Request/response structs, domain types |
| 3 | [Browser Automation](phase-03-browser-automation.md) | ✅ | chromedp-based Meet joining |
| 4 | [Audio Recording & Processing](phase-04-audio-processing.md) | ✅ | ffmpeg-based recording, WAV handling |
| 5 | [OpenAI Integration](phase-05-openai-integration.md) | ✅ | Whisper transcription + GPT analysis |
| 6 | [Gin HTTP API](phase-06-gin-api.md) | ✅ | REST endpoints, middleware, error handling |
| 7 | [CLI Mode](phase-07-cli-mode.md) | ✅ | Cobra CLI wrapping the service layer |
| 8 | [Testing & Deployment](phase-08-testing-deployment.md) | ⬜ | Unit/integration tests, Docker, Makefile |

---

## Key Decisions

1. **chromedp over selenium** — native Go, no WebDriver binary needed, lighter
2. **ffmpeg for audio** — no good native Go audio recording lib; shell out to ffmpeg (same as Python does for compression)
3. **3-layer architecture** — handler → service → external (OpenAI, Chrome, ffmpeg)
4. **Concurrent GPT calls** — use goroutines for parallel summary/keypoints/actions/sentiment (Python does them sequentially)

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Google Meet UI changes break selectors | High | Abstract selectors into config, add retry logic |
| Audio recording in Go is harder than Python | Medium | Use ffmpeg for both recording and processing |
| chromedp can't handle Meet's complex JS | Medium | Fallback: use rod (higher-level Chrome automation) |
| OpenAI Go SDK missing features | Low | Well-maintained, feature-complete |

## Unresolved Questions

1. **Should Gin API be the primary interface or keep CLI-first?** — Plan assumes both (API + CLI sharing service layer)
2. **Database needed?** — Current Python version has none. Consider adding SQLite/Postgres for meeting history in Go version?
3. **Authentication for API?** — Current version has no auth. Add JWT middleware for Gin endpoints?
4. **WebSocket for real-time status?** — Recording progress, transcription status could stream via WS
