export PATH := /opt/homebrew/bin:$(PATH)

.PHONY: dev-api dev-ui dev build

dev-api:
	cd backend && go run ./cmd/api

dev-ui:
	cd frontend && npm run dev

dev:
	make -j2 dev-api dev-ui

build:
	cd backend && go build -o ../bin/hube ./cmd/api
	cd frontend && npm run build
