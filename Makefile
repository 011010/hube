export PATH := /opt/homebrew/bin:$(PATH)

.PHONY: dev-api dev-ui dev build install-cli docker-build docker-up docker-down

dev-api:
	cd backend && go run ./cmd/api

dev-ui:
	cd frontend && pnpm run dev

dev:
	make -j2 dev-api dev-ui

build:
	cd backend && go build -o ../bin/hube-api ./cmd/api
	cd frontend && pnpm run build

install-cli:
	cd backend && go install ./cmd/cli

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down
