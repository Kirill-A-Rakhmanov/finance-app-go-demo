# Finance App Go Demo

Personal finance API written in Go. Manage accounts, transactions, and transfers with JWT auth and Clean Architecture.

## Quick start

Requires only Docker with Compose. No `.env`, Go or `make` needed:

```bash
docker compose up --build
```

* API: `http://localhost:3100/api`
* Swagger UI: `http://localhost:3100/api/swagger/index.html`
* Health: `http://localhost:3100/health`

All settings have demo defaults in [`docker-compose.yaml`](./docker-compose.yaml) (demo secrets only, not for real use). `docker compose down -v` stops everything and wipes the data.

Try it:

```bash
# 1. register (no SMTP in the demo: the confirmation link is written to the app log)
curl -X POST localhost:3100/api/auth/register -H 'Content-Type: application/json' \
  -d '{"firstName":"Demo","lastName":"User","email":"demo@example.com","password":"Password123!"}'
docker compose logs app | grep "registration confirmation"   # copy the token from the link

# 2. confirm and log in
curl -X POST localhost:3100/api/auth/verify-email -H 'Content-Type: application/json' -d '{"token":"<token>"}'
curl -c cookies.txt -X POST localhost:3100/api/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"Password123!"}'
```

---

## Stack

* **Go** 1.26
* **Clean Architecture** — domain → application → infrastructure / adapter
* **Gin** — HTTP API
* **Ent** — schema and generated data access
* **golang-migrate** — SQL migrations (applied on app start)
* **PostgreSQL** — primary storage
* **Redis** — auth sessions, password recovery codes, rate limiting
* **JWT** — access / refresh tokens
* **Swagger** — OpenAPI docs (swaggo)

---

## Domain

| Zone | What you can do |
|------|-----------------|
| **Auth** | Register with email confirmation (can be closed with the `registration` feature flag), login, logout, refresh tokens, verify session, password recovery and reset |
| **Account** | Create and update accounts, list your accounts, get balance |
| **Transaction** | Record income and expense, update, delete, get by id |
| **Transfer** | Move money between accounts, delete a transfer |
| **User** | Current user profile and total balance |
| **Category** | Hierarchical categories, default set on register, full CRUD via HTTP |
| **Counterparty** | Create, update, list and delete counterparties |
| **Report** | Spending by category and counterparty, income vs expense, balance over time, category dynamics by month |
| **Feature flags** | Global flags with per-user overrides; the frontend reads the effective values |

Money rules: an amount must be greater than zero, not exceed 1 000 000 000 000 and have at most 4 decimal places. Operations that lower the balance of an account without overdraft (expenses, transfers out, deleting an income or an incoming transfer) take a lock on the account for the rest of the database transaction (`Manager.Lock(ctx, key)`, a PostgreSQL transaction-level advisory lock; the caller builds the key, here `model.AccountBalanceLockKey`), so concurrent requests cannot overdraw it. A transaction that takes several locks must take them in one agreed order to avoid deadlocks.

Updating a transaction recomputes the effect on the balance: the old confirmed amount is taken out and the new one put in, and the account is checked and locked only when the update lowers its balance. Changing the account moves the transaction (both accounts must be the user's and in the same currency), and then the account it leaves and the account it joins are each checked for the part that lowers their balance.

---

## Architecture

The project follows Clean Architecture:

* **`internal/domain`** — models, repository interfaces, value objects (no external deps)
* **`internal/application`** — use-cases per domain zone
* **`internal/infrastructure`** — PostgreSQL, Redis, password hashing, JWT
* **`internal/adapter`** — HTTP handlers, DTOs, middleware

Full layout and layer notes: [PROJECT_ARCHITECTURE.md](./PROJECT_ARCHITECTURE.md).

---

## Database Migrations

SQL migrations live in `migrations/` and are managed with **golang-migrate**.

They run automatically when the app connects to PostgreSQL. To apply manually:

```bash
migrate -path ./migrations -database "postgres://USER:PASSWORD@HOST:PORT/DB?sslmode=disable" up
```

---

## Registration

Registration is a two-step flow, so an address is only used by the person who owns the mailbox:

1. `POST /api/auth/register` stores the registration in Redis (name, email and the password hash, under a hashed single-use token) and emails a confirmation link built from `FRONTEND_VERIFY_EMAIL_URL`. Nothing is written to PostgreSQL and nobody is signed in.
2. `POST /api/auth/verify-email` with the token creates the user and the default categories. The user then logs in; there is no auto-login.

An unconfirmed registration expires from Redis after `AUTH_REGISTRATION_TTL` (24 hours), so there is nothing to clean up. The response of step 1 is the same whether the email is free, already registered or in its cooldown, so the endpoint does not reveal which emails have accounts: an existing account's owner gets a "someone tried to register" email instead. Only one registration per email is accepted every 10 minutes; a later one replaces the earlier pending registration and its link.

---

## Feature Flags

Flags live in two tables: `feature_flags` (global value) and `feature_flag_overrides` (per-user value). The effective value is the override if there is one, otherwise the global value; an unknown flag is disabled.

* Flags known to the code are listed in `model.DefaultFeatureFlags`. Missing ones are created on startup with their default value; existing rows are never modified, so values changed by hand survive restarts.
* Change a value with plain SQL, e.g. close registration: `UPDATE feature_flags SET enabled = false WHERE key = 'registration';`
* Endpoints: `GET /api/feature-flag/list` (effective values for the current user) and `GET /api/feature-flag/global` (public, global values).
* Backend checks use `FeatureFlagRepository.IsEnabled` / `IsEnabledForUser`.

---

## API

Interactive Swagger UI:

`http://localhost:3100/api/swagger/index.html`

The UI is served only when `APP_SWAGGER_ENABLED=true`; set it to `false` on a public server (the variable is required, there is no default).

---

## Testing

Covered with table-driven tests, using [testify](https://github.com/stretchr/testify) (`assert`/`require`) for assertions and [mockery](https://github.com/vektra/mockery) for generated mocks of repository and service interfaces:

* **`internal/domain`** — value objects with real validation logic (`Email`, `Currency`) and custom error types
* **`internal/application`** — every use-case per domain zone, mocking repositories/services one layer down
* **`internal/adapter/http`** — handlers exercised end-to-end through a real `net/http/httptest` router with a real service backed by mocked repositories, checking status-code mapping, request binding, and response shape

Integration tests (`*_integration_test.go`, build tag `integration`) run the Postgres repositories and query services, the Redis storages and the rate limiter against a real PostgreSQL/Redis started with testcontainers. They need Docker and are skipped by `make test`.

```bash
make test              # unit tests (same as go test ./...)
make test-integration  # integration tests (go test -tags integration ./...)
make verify            # build + vet + gofmt check + unit tests
go test ./... -v       # verbose, per-scenario output
go test ./... -cover   # coverage per package
```

Mocks live in a `mock` subpackage next to the interface they mock (e.g. `internal/domain/repository/mock`), configured in [`.mockery.yaml`](./.mockery.yaml). Regenerate after adding or changing an interface:

```bash
make mocks             # go tool mockery
```

---

## License

MIT.
