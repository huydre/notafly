.PHONY: build build-server run run-server test lint clean

build:
	go build -o bin/notafly ./cmd/cli

build-server:
	go build -o bin/notafly-server ./cmd/server

run:
	go run ./cmd/cli serve

run-server:
	go run ./cmd/server

test:
	go test -race -cover ./...

lint:
	golangci-lint run

clean:
	rm -rf bin/

all: clean build build-server
