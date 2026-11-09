.DEFAULT_GOAL := help

# Run
.PHONY: up down logs

up: ## Build and start app + Postgres + Redis in Docker
	docker compose up --build -d

down: ## Stop Docker services and remove volumes (fresh DB next time)
	docker compose down -v

logs: ## Tail app container logs
	docker compose logs -f app

# Code generation
.PHONY: generate ent swagger mocks

generate: ent swagger mocks ## Regenerate everything (Ent, Swagger docs, mocks)

ent: ## Regenerate Ent code after changing ent/schema
	go generate ./ent

swagger: ## Regenerate Swagger docs (docs/) after changing handlers or DTOs
	swag init -g main.go

mocks: ## Regenerate mocks from .mockery.yaml
	go tool mockery

# Quality
.PHONY: build vet fmt fmt-check test test-integration verify tidy

build: ## Compile everything
	go build ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format all Go files
	gofmt -w .

fmt-check: ## Fail if any file is not gofmt-formatted
	@out="$$(gofmt -l .)"; \
	if [ -n "$$out" ]; then echo "Not formatted:"; echo "$$out"; exit 1; fi

test: ## Run unit tests (no Docker needed)
	go test ./...

test-integration: ## Run integration tests (testcontainers, needs Docker)
	go test -tags integration ./...

verify: build vet fmt-check test ## Definition of done: build, vet, gofmt check, tests

tidy: ## Tidy go.mod / go.sum
	go mod tidy

# Help
.PHONY: help

help: ## Show this help
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
