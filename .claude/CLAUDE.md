# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Backend for a personal finance app: Go 1.26, Gin (REST), Ent + PostgreSQL, Redis (sessions, rate
limiting), JWT in cookies, golang-migrate. Runs with Docker Compose only (`make up`). Listens on
`:3100`, API base path `/api`, Swagger UI at `/api/swagger/index.html`. The Swagger doc is the
HTTP contract for clients — see `rules/swagger-and-api-contract.md`.

## Commands

All common tasks are in the `Makefile` (`make help` lists them):

```bash
make up / down  # docker compose: app + Postgres + Redis (no .env needed)
make logs       # tail app container logs

make build      # go build ./...
make vet        # go vet ./...
make fmt        # gofmt -w .
make fmt-check  # fails if gofmt -l prints anything
make test       # go test ./... (unit tests only)
make test-integration  # go test -tags integration ./... (needs Docker)
make verify     # build + vet + fmt-check + test — run this before considering a task done

make generate   # ent + swagger + mocks
make ent        # go generate ./ent  (after changing ent/schema)
make swagger    # swag init -g main.go  (after changing handlers/DTOs)
make mocks      # go tool mockery  (after changing interfaces / .mockery.yaml)
make tidy       # go mod tidy

# single package / single test:
go test ./internal/application/transaction/...
go test ./internal/application/transaction -run TestService_DeleteTransaction
```

Definition of done for a task: `make verify` passes. If you touched handlers/DTOs, `docs/` must be
regenerated too (`make swagger`).

Integration tests (`*_integration_test.go`, build tag `integration`) use testcontainers and need Docker
running. `make test` skips them; run them with `make test-integration`.

## Architecture

Clean Architecture, dependencies point inward: `domain ← application ← adapter / infrastructure`.
Full directory layout is in `PROJECT_ARCHITECTURE.md` — don't duplicate it here.

- `internal/domain/` — models, enums, value objects, repository **interfaces**, domain errors.
  No framework or driver imports.
- `internal/application/<zone>/` — use-cases on a `Service` struct, one file per use-case.
  Depends on domain interfaces only. `application/error` holds the shared error type and
  `ValidationError`; `application/query` holds read-model query services.
- `internal/infrastructure/` — implementations of the ports (Ent/Postgres repositories, Redis,
  password hashing, JWT, notifier).
- `internal/adapter/http/<zone>/` — Gin handlers (one file per endpoint), plus shared `dto/`,
  `mapper/`, `request/` (session in context), `response/` (JSON + error → status mapping),
  `middleware/`, `cookie/`.
- `app/` — composition root; the only place that wires concrete types together. `main.go` stays thin.
- `ent/schema/` — source of truth for DB entity shapes. `migrations/` — versioned SQL, applied
  automatically on Postgres connect (`pkg/postgres`).
- `pkg/` — shared infrastructure libraries (`httpserver`, `logger`, `redis`, `postgres`,
  `transaction` DB-transaction manager).

Generated code — never edit by hand: `docs/` (swag), `ent/` except `ent/schema/` (ent),
`**/mock/` (mockery).

## Tooling notes

- Module name is `finance-app-go-demo`; imports are `finance-app-go-demo/internal/...`.
- Config comes from env (`config/`); demo defaults are set in `docker-compose.yaml`.
- `.claude/settings.local.json` is personal (globally git-ignored); shared permissions go in
  `.claude/settings.json`.
