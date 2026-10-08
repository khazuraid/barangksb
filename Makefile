.PHONY: dev-backend dev-frontend build test up down logs create-user

dev-backend:
	cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

build:
	cd backend && go build -o bin/server ./cmd/server
	cd frontend && npm run build

test:
	cd backend && go test ./...

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f

create-user:
	cd backend && go run ./cmd/server -create-user admin@kantor.id
