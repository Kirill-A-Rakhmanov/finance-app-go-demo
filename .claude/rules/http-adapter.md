---
description: Gin handlers, DTOs, mappers, error-to-status mapping
paths:
    - 'internal/adapter/http/**/*.go'
---

# HTTP adapter (`internal/adapter/http`)

- One file per endpoint in `<zone>/` (`delete_transaction.go`), named after the use-case it calls.
  `handler.go` holds the `Handler` struct and `NewHandler(service)`; the handler depends on the
  zone's `*Service` only.
- Handler flow, in this order: parse path/query/body → read the session → build the use-case
  input → call the service → map the result to a response DTO → write it.
  - Bad path/query/body → `response.BadRequest(c.Writer, fmt.Errorf("invalid x: %w", err))`.
  - No session → `response.Unauthorized`. Get it with `request.CurrentSession(ctx)`; the
    user ID comes **only** from the session, never from the request body/params.
  - Service errors → always `response.FromError(c.Writer, err)`. Don't pick status codes by hand
    for use-case errors. `FromError` maps `ValidationError` → 422 with `fields`, and
    `apperror.Kind` → 400/401/403/404/409 via `kindHandlers`; anything else is 500. To support a
    new kind, extend `kindHandlers` rather than special-casing in a handler.
  - Success → `response.OK / Created / NoContent`.
- Handlers write through `c.Writer` with the `response` package helpers, not `c.JSON`.
- Pass `c.Request.Context()` into the service as `ctx`.
- Request/response shapes shared between zones live in `dto/`; conversion between domain models and
  DTOs lives in `mapper/` (`<zone>_mapper.go`). Endpoint-specific request/response structs can sit
  in the handler file.
- Response struct fields have `json:"camelCase"` tags and `validate:"required"` for fields that are
  always present (this drives required-ness in the generated Swagger/TS types).
- No business logic here: no balance math, no ownership checks beyond passing the session user ID
  down. If a handler needs an `if` about domain rules, it belongs in the use-case.
- Routes are registered in `router.go`; authenticated routes go under the auth middleware group.
