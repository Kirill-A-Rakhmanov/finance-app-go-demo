---
description: Conventions for application-layer use-cases, validation and errors
paths:
    - 'internal/application/**/*.go'
---

# Use-cases (`internal/application`)

- One package per zone (`account`, `auth`, `transaction`, `transfer`, `user`,
  `transaction_category`). Package directories use snake_case.
- One file per use-case, named after the action in snake_case: `delete_transaction.go`, with its
  test next to it as `delete_transaction_test.go`.
- Each use-case file contains `XInput`, `XResult` (when it returns more than an entity) and a
  method `func (s *Service) X(ctx context.Context, input XInput) (*XResult, error)`.
  `ctx` is always the first parameter.
- `Service` and its constructor `NewService(...)` live in `service.go`. Dependencies are
  interfaces, injected via the constructor, stored in unexported fields.
- Every `XInput` has `func (i *XInput) Validate() error` built on `apperror.NewValidationError()`
  (`v.AddIf(cond, "field", "message")`, `v.MaxLength`, …) and returns `v.ErrorOrNil()`. Call it
  first thing in the use-case. Field names in validation errors are camelCase (`transactionId`)
  to match the JSON API.
- Max lengths come from constants in `internal/domain/model/limits.go` — add new ones there, don't
  hardcode numbers.
- Errors are `*apperror.Error` built with `apperror.Invalid / NotFound / Conflict / Forbidden /
  Unauthorized`; the `Kind` decides the HTTP status (see `http-adapter.md`). Put errors shared by
  several zones in `application/error/common.go`; zone-specific ones in that package's
  `error.go` as exported `Err...` vars. Import the shared package as `apperror`.
- Return repository errors as-is (unknown errors become 500). A repository returning
  `nil, nil` means "not found" — the use-case converts that to the proper `Err...NotFound`.
- Anything that touches several writes (balances, transfers) goes through the transaction manager
  (`pkg/transaction.Manager`) so it is atomic.
- Ownership: every lookup of a user's data is scoped by `UserID` (`GetByIDAndUserID`). Never fetch
  by bare ID and trust the caller.
- No `gin`, `http`, or `ent` types in this layer.
