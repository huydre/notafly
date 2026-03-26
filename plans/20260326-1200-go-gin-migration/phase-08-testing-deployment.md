# Phase 8: Testing & Deployment

**Priority:** P1
**Status:** ✅ Completed

---

## Context

Python version has **zero tests**. Go version should have proper test coverage from the start.

## Testing Strategy

### Unit Tests (service layer)

```go
// internal/service/transcriber_test.go
func TestAnalyzeMeeting_ConcurrentCalls(t *testing.T) {
    // Mock OpenAI client
    // Verify all 4 analyses run
    // Verify result struct populated
}
```

### Integration Tests (handler layer)

```go
func TestJoinMeet_InvalidRequest(t *testing.T) {
    router := setupTestRouter()
    w := httptest.NewRecorder()
    body := `{"meet_link": "not-a-url"}`
    req, _ := http.NewRequest("POST", "/api/v1/meet/join", strings.NewReader(body))
    router.ServeHTTP(w, req)
    assert.Equal(t, 400, w.Code)
}
```

### Test Structure

```
internal/
├── service/
│   ├── transcriber.go
│   └── transcriber_test.go      # Unit tests
├── handler/
│   ├── meet.go
│   └── meet_test.go             # Integration tests (httptest)
```

## Deployment

### Dockerfile

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o notafly ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache chromium ffmpeg
COPY --from=builder /app/notafly /usr/local/bin/
EXPOSE 8080
CMD ["notafly"]
```

### Makefile

```makefile
.PHONY: build run test lint

build:
	go build -o bin/notafly ./cmd/server

run:
	go run ./cmd/server

test:
	go test -race -cover ./...

lint:
	golangci-lint run

docker:
	docker build -t notafly .
```

## Implementation Steps

- [ ] Add unit tests for TranscriberService (mock OpenAI)
- [ ] Add unit tests for RecorderService (mock exec)
- [ ] Add handler integration tests with httptest
- [ ] Add Dockerfile (multi-stage: builder + alpine with chromium + ffmpeg)
- [ ] Add Makefile
- [ ] Add GitHub Actions CI (test + lint + build)
- [ ] Add `.golangci.yml` for linter config

## Success Criteria

- `go test ./...` passes with >70% coverage
- `docker build` produces working image
- CI pipeline green on push
