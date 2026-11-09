---
description: Clean Architecture layer boundaries and dependency direction
paths:
    - 'internal/**/*.go'
    - 'app/**/*.go'
    - 'pkg/**/*.go'
---

# Architecture boundaries

- Dependencies point inward only: `domain ← application ← adapter / infrastructure`. A lower
  (inner) layer must never import an outer one.
- `internal/domain` imports nothing from `application`, `adapter`, `infrastructure`, or any
  framework/driver (gin, ent, redis, pgx, jwt…). It holds models, enums, value objects, domain
  errors and repository **interfaces**.
- `internal/application` depends on domain interfaces and its own ports (e.g. `PasswordService`,
  `SessionStorage`, `TokenService` in `application/auth`). It must not import
  `internal/infrastructure` or `internal/adapter`, and must not import `ent` or `gin`.
- `internal/adapter/http` talks to the application layer through `*Service` use-cases. Handlers
  never call repositories or `ent` directly.
- `internal/infrastructure` implements domain/application ports. It may import `ent`, redis, etc.,
  but contains no business rules.
- Wiring concrete types together happens only in `app/`. Don't construct infrastructure objects
  inside use-cases or handlers.
- A new port (interface) lives with its consumer: repository interfaces in `domain/repository`,
  service ports in the application package that uses them. After adding one, register it in
  `.mockery.yaml` (see `testing.md`).
- Cross-zone calls between application packages (e.g. `transfer` needing `account` data) should go
  through repository interfaces or `application/query`, not by importing another zone's `Service`.
