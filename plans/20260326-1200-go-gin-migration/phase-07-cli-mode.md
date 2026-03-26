# Phase 7: CLI Mode

**Priority:** P2
**Status:** ✅ Completed

---

## Context

Preserve CLI functionality from Python version. Uses same service layer as Gin API.

## CLI Design (Cobra)

```
notafly join --meet-link <url> --duration <seconds> [--no-analysis]
notafly transcribe --audio <path> [--no-analysis]
notafly serve [--port 8080]
```

```go
// cmd/cli/main.go
var rootCmd = &cobra.Command{Use: "notafly"}

var joinCmd = &cobra.Command{
    Use:   "join",
    Short: "Join a Google Meet, record, and analyze",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Same service calls as Gin handler
        meetSvc.JoinAndRecord(ctx, meetLink, duration)
        if !noAnalysis {
            transcriberSvc.TranscribeAndAnalyze(ctx, audioPath)
        }
        return nil
    },
}
```

## Mapping from Python argparse

| Python flag | Go cobra flag |
|-------------|---------------|
| `--meet-link` | `--meet-link` / `-m` |
| `--duration` | `--duration` / `-d` |
| `--no-analysis` | `--no-analysis` / `-n` |
| N/A (new) | `serve --port` |

## Implementation Steps

- [ ] Create `cmd/cli/main.go` with cobra root command
- [ ] Add `join` subcommand
- [ ] Add `transcribe` subcommand
- [ ] Add `serve` subcommand (starts Gin server)
- [ ] Share service layer with API handlers

## Success Criteria

- `notafly join --meet-link <url> --duration 60` works like Python version
- `notafly serve` starts Gin HTTP server
