# Phase 1: Project Setup & Config

**Priority:** P0 — Must do first
**Status:** ✅ Completed

---

## Context

- Python code moved to `old/` for reference
- Go code lives at project root (`/Users/hnam/Desktop/notafly/`)

## Target Project Structure

```
notafly/                          # project root
├── old/                          # Python code (reference only)
│   ├── src/google_meet_bot/
│   ├── pyproject.toml
│   ├── requirements.txt
│   └── ...
├── plans/                        # Migration plan docs
├── cmd/
│   ├── server/
│   │   └── main.go              # Gin HTTP server entry
│   └── cli/
│       └── main.go              # CLI entry (cobra)
├── internal/
│   ├── config/
│   │   └── config.go            # Env loading, validation
│   ├── handler/
│   │   ├── meet.go              # POST /api/meet/join
│   │   ├── transcribe.go        # POST /api/transcribe
│   │   └── health.go            # GET /health
│   ├── service/
│   │   ├── meet.go              # Browser automation logic
│   │   ├── recorder.go          # Audio recording logic
│   │   └── transcriber.go       # OpenAI Whisper + GPT logic
│   ├── model/
│   │   ├── meeting.go           # Domain types
│   │   └── analysis.go          # Summary, KeyPoints, etc.
│   ├── middleware/
│   │   ├── auth.go              # JWT (optional)
│   │   ├── cors.go              # CORS
│   │   └── logger.go            # Request logging
│   └── dto/
│       ├── request.go           # API request structs
│       └── response.go          # API response structs
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
└── README.md
```

## Config Struct (maps from current .env.example)

```go
type Config struct {
    // Google credentials
    EmailID       string `env:"EMAIL_ID,required"`
    EmailPassword string `env:"EMAIL_PASSWORD,required"`

    // Meeting
    MeetLink          string `env:"MEET_LINK"`
    RecordingDuration int    `env:"RECORDING_DURATION" envDefault:"60"`

    // Audio
    SampleRate       int `env:"SAMPLE_RATE" envDefault:"44100"`
    MaxAudioSizeBytes int `env:"MAX_AUDIO_SIZE_BYTES" envDefault:"20971520"`

    // OpenAI
    OpenAIAPIKey string `env:"OPENAI_API_KEY,required"`
    GPTModel     string `env:"GPT_MODEL" envDefault:"gpt-4"`
    WhisperModel string `env:"WHISPER_MODEL" envDefault:"whisper-1"`

    // Server (NEW)
    Port string `env:"PORT" envDefault:"8080"`
}
```

## Dependencies (go.mod)

```
github.com/gin-gonic/gin          # HTTP framework
github.com/chromedp/chromedp      # Chrome DevTools Protocol
github.com/sashabaranov/go-openai # OpenAI API client
github.com/joho/godotenv          # .env file loading
github.com/spf13/cobra            # CLI framework
github.com/caarlos0/env/v11       # Struct-based env parsing
go.uber.org/zap                   # Structured logging
```

## Implementation Steps

- [ ] `go mod init github.com/hnam/notafly`
- [ ] Create directory structure
- [ ] Implement `internal/config/config.go` with env parsing + validation
- [ ] Create `cmd/server/main.go` with basic Gin setup
- [ ] Create `cmd/cli/main.go` stub
- [ ] Add `.env.example` for Go version
- [ ] Add `Makefile` with `build`, `run`, `test`, `lint` targets

## Success Criteria

- `go build ./...` passes
- Config loads from `.env` correctly
- `GET /health` returns 200
