# Notafly

Google Meet bot — join meetings, record audio, transcribe with Whisper, and analyze with GPT.

Rewritten in Go with Gin HTTP API + Cobra CLI.

## Quick Start

```bash
# Copy and configure environment
cp .env.example .env
# Edit .env with your credentials

# Run as CLI
go run ./cmd/cli join -m "https://meet.google.com/xxx-xxxx-xxx" -d 60

# Run as HTTP server
go run ./cmd/cli serve
```

## CLI Commands

```
notafly join -m <url> -d <seconds> [-n]    # Join + record + analyze
notafly transcribe -a <file> [-n]           # Transcribe audio file
notafly serve [-p 8080]                     # Start HTTP API server
```

## API Endpoints

```
GET  /health                  → Health check
POST /api/v1/meet/join        → Join meeting + record
POST /api/v1/meet/full        → Full pipeline (join + record + transcribe + analyze)
POST /api/v1/transcribe       → Transcribe audio file
```

## Requirements

- Go 1.26+
- Chrome/Chromium (for browser automation)
- ffmpeg + ffprobe (for audio recording/processing)
- OpenAI API key

## Build

```bash
make build          # Build CLI binary
make build-server   # Build standalone server
make test           # Run tests
make docker         # Build Docker image
```

## Docker

```bash
docker build -t notafly .
docker run -p 8080:8080 --env-file .env notafly
```

## License

MIT
