.PHONY: migrate-create migrate-up migrate-down up down build logs restart clean dev test test-coverage swag-init

DOCKER_COMPOSE = docker compose --env-file .env -f docker/docker-compose.yaml


migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=your_migration_name"; \
		exit 1; \
	fi
	goose -dir internal/database/migrations create $(name) sql

migrate-up:
	@if [ -z "$$DATABASE_URL" ] && [ -f .env ]; then \
		set -a; . ./.env; set +a; \
	fi; \
	if [ -z "$$DATABASE_URL" ]; then \
		echo "DATABASE_URL is required"; \
		exit 1; \
	fi; \
	goose -dir internal/database/migrations postgres "$$DATABASE_URL" up

migrate-down:
	@if [ -z "$$DATABASE_URL" ] && [ -f .env ]; then \
		set -a; . ./.env; set +a; \
	fi; \
	if [ -z "$$DATABASE_URL" ]; then \
		echo "DATABASE_URL is required"; \
		exit 1; \
	fi; \
	goose -dir internal/database/migrations postgres "$$DATABASE_URL" down

up:
	$(DOCKER_COMPOSE) up -d --build

down:
	$(DOCKER_COMPOSE) down

build:
	$(DOCKER_COMPOSE) build

logs:
	$(DOCKER_COMPOSE) logs -f

restart:
	$(DOCKER_COMPOSE) down && $(DOCKER_COMPOSE) up -d --build

clean:
	$(DOCKER_COMPOSE) down -v

dev:
	# Replace with your hot-reload tool: air, realize, nodemon, etc.
	go run ./cmd/server

test:
	go test ./... -v

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

swag-init:
	swag init -g ./cmd/server/main.go -o ./cmd/server/docs --parseInternal --parseDependency