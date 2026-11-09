---
description: Go code style conventions beyond what gofmt and go vet enforce
paths:
    - '**/*.go'
---

# Code style

- `gofmt` is mandatory: `gofmt -l .` must print nothing before a task is done.
- Imports: standard library first, then module-internal (`finance-app-go-demo/...`), then third-party,
  separated by blank lines; the surrounding files are the reference. Established aliases:
  `apperror` (`internal/application/error`), `repomock` (`domain/repository/mock`),
  `httptransaction`-style `http<zone>` when importing an adapter package from a test.
- File names are snake_case (`get_transaction_by_id.go`); directory names are snake_case when
  multi-word (`transaction_category`). Exported identifiers are PascalCase, no stutter
  (`transaction.Service`, not `transaction.TransactionService`).
- Wrap errors with context using `fmt.Errorf("...: %w", err)` where context helps; never discard an
  error silently. Don't log-and-return the same error.
- Pass `context.Context` as the first parameter and propagate it to every repository/IO call.
- Prefer small, concrete types and constructor injection; avoid package-level mutable state.
  Package-level `var Err... = apperror.X(...)` sentinels are fine.
- No comments in production code. Prefer code that doesn't need explaining: clear names, small
  functions, and reaching for the right language/API feature over prose. Allowed:
  - swagger annotations (`// @Summary`, `// @Router`, ...) above handlers;
  - a short comment for a genuinely non-obvious reason that the code cannot express (a workaround,
    a subtle ordering or concurrency constraint). These should be rare; don't restate what the
    code does;
  - `// todo` markers, when something is deliberately deferred. Mention it to the user as well.
- Tests may have short comments, one line where possible, that explain a non-obvious setup, for
  example why a call is not mocked because the method returns earlier ("DeleteRecovery is not
  expected: the user can retry with the same code"). Don't narrate what each step does.
- Don't use `panic` outside of startup/wiring in `app/` and `main.go`.
- Don't add dependencies (`go get`) for something the standard library or an existing dependency
  already does; mention it to the user before adding a new one.
- Don't fix style or lint issues in generated code (`docs/`, `ent/`, `**/mock/`) — regenerate it.
