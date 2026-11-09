---
description: Ent schema, generated code, SQL migrations, repository implementations
paths:
    - 'ent/**'
    - 'migrations/**'
    - 'internal/infrastructure/postgres/**'
    - 'internal/domain/repository/**'
---

# Persistence

- `ent/schema/*.go` is the source of truth for entity shapes. After editing a schema run
  `make ent`. Never hand-edit anything in `ent/` outside `ent/schema/` — it is fully
  regenerated.
- Schema changes need a matching SQL migration; Ent does not auto-migrate here. Migrations are
  applied on startup by `pkg/postgres` (`golang-migrate`, `file://migrations`).
  - File pair per change: `YYYYMMDDHHMMSS_name.up.sql` and `.down.sql`. The down migration must
    actually reverse the up.
  - Never edit a migration that has already been applied/committed — add a new one.
  - Keep SQL types consistent with the schema (e.g. `uuid` ids with `gen_random_uuid()`,
    `timestamptz`, `numeric(19,4)` for money, `varchar(n)` matching the `*MaxLen` limits in
    `domain/model/limits.go`).
- Repository **interfaces** live in `internal/domain/repository`; implementations live in
  `internal/infrastructure/postgres`, one file per aggregate. Implementations convert `ent`
  entities to domain models at the boundary — `ent` types must not leak out of
  `infrastructure/postgres`.
- "Not found" from a repository is `nil, nil` (use-cases translate it), not a sentinel error.
- Reads that don't fit an aggregate (lists, filters, joins) go through query services
  (`application/query` interface + `infrastructure/postgres/*_query.go`).
- Multi-write operations run inside the transaction manager (`pkg/transaction`); repository
  methods must pick up the tx from `ctx` rather than opening their own.
- Queries on user-owned data always filter by `user_id`.
- New repository interface or method → update the mock via `make mocks` (see `testing.md`).
