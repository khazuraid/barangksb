.PHONY: run build test generate tidy

run:
	air 2>/dev/null || go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...

generate:
	cd internal/views && templ generate

tidy:
	go mod tidy
