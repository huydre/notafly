FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build both binaries
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o notafly ./cmd/cli
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o notafly-server ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache \
    chromium \
    ffmpeg \
    ca-certificates \
    tzdata

# chromedp needs to know where Chromium is
ENV CHROME_BIN=/usr/bin/chromium-browser

COPY --from=builder /app/notafly /usr/local/bin/
COPY --from=builder /app/notafly-server /usr/local/bin/

EXPOSE 8080

CMD ["notafly", "serve"]
