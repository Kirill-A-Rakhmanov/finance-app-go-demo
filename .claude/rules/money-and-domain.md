---
description: Money handling and core finance domain invariants
paths:
    - 'internal/**/*.go'
---

# Money and domain invariants

- Money is `shopspring/decimal.Decimal` in the domain and application layers and `numeric(19,4)`
  in Postgres. Never use `float64` for arithmetic. `float64` appears only as a JSON field type in
  some HTTP DTOs; convert at the mapper/handler boundary (`decimal.NewFromFloat` /
  `.InexactFloat64()`) and do no math on it.
- Compare decimals with `.Equal`, `.Cmp`, `.IsNegative()`, `.IsZero()` — never `==`.
- Amounts must be strictly positive on create/update: zero → `ErrZeroAmountNotAllowed`, negative →
  `ErrNegativeAmountNotAllowed`.
- Balance is derived (`GetBalanceByAccountID`: confirmed transactions + account adjustment), not
  stored. Only **confirmed** transactions (`IsConfirmed`) affect it: income adds, expense
  subtracts. Deleting/updating a confirmed transaction must recompute the resulting balance.
- An account with `IsNegativeBalanceEnabled == false` must never end up with a negative balance
  after any operation (create, update, delete, transfer). Check the post-operation balance and
  return `apperror.ErrNegativeBalanceNotAllowed`.
- A transaction's type must match its category's type (`income` / `expense`) —
  `ErrInvalidTransactionCategoryType` otherwise.
- Transactions with a `TransferID` belong to a transfer: they can't be updated or deleted as plain
  transactions (`ErrTransferUpdateNotAllowed`, `ErrDeleteTransferNotAllowed`). Change them via the
  `transfer` use-cases.
- Accounts carry a 3-letter currency; validate it via the currency value object / domain currency
  errors rather than ad-hoc string checks. Transfers between accounts of different currencies carry
  both `fromAmount` and `toAmount`.
- Users only ever see and modify their own accounts, transactions, categories and transfers. A
  resource that exists but belongs to someone else is reported as not found.
- Default categories are created on registration (`auth/default_categories.go`); changing them
  affects every new user.
