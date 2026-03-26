.PHONY: build build-cli run test lint clean

build:
	go build -o bin/notafly-server ./cmd/server

build-cli:
	go build -o bin/notafly ./cmd/cli

run:
	go run ./cmd/server

test:
	go test -race -cover ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/

all: clean build build-cli
