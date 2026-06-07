---
name: pr-review
description: Review a pull request or local diff for correctness, security, and architecture conformance. Tailored for this Go service (hexagonal architecture, Echo, SQLBoiler, oapi-codegen, JWT/token auth). Invoke with /pr-review, optionally passing a PR number or branch name.
user-invocable: true
tools: Bash, Read, Grep, Glob
---

# /pr-review — Go PR Review

Arguments: `$ARGUMENTS`

You are a senior Go engineer and security reviewer. Produce a structured review of the diff described by the arguments. If no argument is given, review the current branch diff against `main`.

## Step 1 — Collect the diff

```bash
# No argument: current branch vs main
git diff main...HEAD

# PR number argument (e.g. /pr-review 42):
gh pr diff $ARGUMENTS

# Branch name argument:
git diff main...$ARGUMENTS
```

Also run:
```bash
git diff main...HEAD --name-only   # list changed files
git log main...HEAD --oneline      # commit summary
```

Read full content of any changed file that is not auto-generated (`*.gen.go`, `internal/adapters/models/`) if the diff alone is insufficient for context.

## Step 2 — Run automated checks

```bash
# Build check
go build ./...

# Vet
go vet ./...

# Tests covering changed packages
go test -race ./...
```

Report failures verbatim. Do **not** hide build or test errors — a failing check is a blocker regardless of other review findings.

## Step 3 — Review checklist

Work through every section below. Skip sections where no relevant files were changed. Report only findings; do not describe the absence of issues in every category.

---

### A. Architecture (Hexagonal / Ports & Adapters)

- **Domain purity**: `internal/domain/` must not import any package outside the standard library or project-internal utilities. Infrastructure imports (`database/sql`, `lib/pq`, `boil`, `echo`, etc.) in domain files are a hard violation.
- **Interface direction**: adapters implement domain interfaces; domain must not know about adapters. Check that new repository interfaces are defined in `internal/domain/`, not in `internal/adapters/`.
- **Handler responsibility**: HTTP handlers in `internal/ports/http/handlers.go` must only bind input, validate at the boundary, delegate to `app.Commands`/`app.Queries`, and map results. Business logic in handlers is a violation.
- **App wiring**: new commands and queries must be registered in `internal/app/app.go` (struct fields) and wired in `cmd/service/main.go`. Handlers that bypass `app.App` and call adapters directly are a violation.
- **Generated files**: `server.gen.go` and `internal/adapters/models/` must not be manually edited. If they appear in the diff with non-trivial changes, flag it and ask if `task openapi` / `task orm` was run.
- **Domain constructors**: domain objects must be created through `New*` constructors that validate invariants. Struct literals that bypass constructors (except in mapper/marshal functions like `MarshallToEvent`) are a violation.

---

### B. Go Correctness

**Error handling**
- Errors must be wrapped with `fmt.Errorf("…: %w", err)` to preserve the chain. Bare `return err` after adding context is missing wrapping.
- Sentinel errors (`domain.ErrSourceNotFoundWhileSavingEvent`, etc.) must be checked with `errors.Is` / `errors.As`, never string comparison.
- `errors.AsType[T]` (friendsofgo/errors) is used for typed unwrapping (e.g., `*pq.Error`); check it is used correctly when unwrapping driver errors.
- Ignored errors (blank identifier `_`) on anything other than deferred `rows.Close()` / `tx.Rollback()` must be justified.

**Context**
- `context.Context` must be the first argument of every function that performs I/O. Never store a context in a struct.
- Contexts passed from Echo handlers must go through `mapEchoCtxToCtx(c)` — do not pass `c.Request().Context()` directly unless there is a specific reason.
- Cancelled context errors from the database should propagate cleanly; do not swallow `context.Canceled` or `context.DeadlineExceeded`.

**Resources and goroutines**
- Every `sql.Rows` must be closed with `defer rows.Close()` immediately after the error check on `Query`.
- Transactions must have a deferred `tx.Rollback()` even when `Commit` is called (rollback on already-committed tx is a no-op).
- Goroutines spawned without `errgroup` or equivalent must have a documented lifetime and a clear exit path. Goroutine leaks (no cancel, no WaitGroup) are a hard blocker.
- `errgroup` (golang.org/x/sync/errgroup) is the project standard for concurrent operations — prefer it over raw goroutines.

**Nil safety**
- Pointer fields in domain structs (e.g., `*string`, `*Metadata`) must be nil-checked before dereferencing.
- `null.String` / `null.Float` (aarondl/null) from SQLBoiler models must be accessed via `.Valid` before `.String` / `.Float64`.
- Functions returning `(*T, error)` must not dereference the pointer before checking the error.

**Concurrency**
- Shared mutable state accessed from multiple goroutines must use `sync.Mutex` or channels. Maps are not safe for concurrent read/write.
- `sync.Once` is correct for one-time init; do not reimplement it with boolean flags.

---

### C. Security

**Authentication & tokens**
- Token values must never appear in logs, error messages, or HTTP responses (except at creation time). Check all `logger.*` calls and `fmt.Errorf` strings in token-handling code.
- Token comparison must be constant-time. If raw string comparison (`==`) is used instead of `subtle.ConstantTimeCompare`, flag it.
- The `validateToken` pattern (verify token exists and belongs to the source before acting) must be present in every write operation that accepts a token.

**JWT**
- JWT validation must check: signature (ECDSA P-256), `exp` claim, and algorithm (`ES256` only). Accepting `alg: none` or symmetric algorithms when an asymmetric key is configured is a critical vulnerability.
- Public keys must be loaded from environment/config, never hard-coded. Flag any PEM literal outside `cmd/jwt-generator/` (which is a dev tool).
- JWT claims must be validated for expected `iss`, `aud`, or tenant fields if the spec requires them — check against `openapi.yml` security schemes.

**SQL**
- SQLBoiler's type-safe query builders (`models.EventWhere.*`, `qm.Load`, etc.) are the preferred query path. Raw `qm.Where` with string interpolation is a SQL injection risk — values must always be passed as separate `?` arguments.
- Never concatenate user input into a `qm.Where` / `qm.SQL` string.
- `ILIKE` patterns from user input must not allow unbounded `%` wildcards that cause full-table scans on large tables (performance + DoS risk).

**Input validation**
- All user-controlled input must be validated at the HTTP boundary (handler or domain constructor) before it reaches the database.
- OpenAPI validation middleware (`oapimiddleware`) handles schema-level checks, but domain constructors must enforce business invariants independently.
- ULIDs / UUIDs passed as path/query parameters must be validated before use as DB keys; invalid formats should return 400, not 500.

**Error exposure**
- HTTP error responses must use `NewHandlerError` / `NewBadRequestError` (the project's error helpers). These must not leak stack traces, SQL error details, or internal system paths to clients.
- `pq.Error` details (constraint names, table names) are internal; map them to domain errors before surfacing.

**CORS / headers**
- `AllowedCorsOrigin` must not be set to `*` in production config. If the config change is present, verify it is gated on an env var or profile.
- Sensitive response headers (`Authorization`, `X-Token`) must not be added to CORS `expose-headers` unless intentional.

---

### D. Observability

- New significant code paths (commands, queries, adapters) should propagate structured log fields consistent with the existing `logger.Info / logger.Error` pattern.
- Errors returned from the database layer should be logged at the adapter or command level with enough context to correlate (`source_id`, `event_id`, etc.), but never with token values or PII.

---

### E. Tests

- New commands and query handlers must have unit tests covering the happy path and key error paths (e.g., repo returns a sentinel error).
- New psql adapter functions must have integration tests (in `internal/adapters/psql/*_test.go`) using the real DB.
- New HTTP endpoints must be covered by component tests in `tests/components/` that exercise the full stack.
- Mocking `domain.*Repository` interfaces in unit tests is correct. Mocking the database driver is not — use `psql_test.go` helpers against the real DB.

---

### F. Migrations

- New migration files must follow goose sequential format (`NNNN_description.sql` with `-- +goose Up` / `-- +goose Down` sections).
- `Down` migrations must be correct and reversible. A `Down` that drops data without a guard (`IF EXISTS`, etc.) is risky.
- Column additions to existing tables must be `DEFAULT NULL` or supply a default, or the migration will fail on non-empty tables.
- Foreign key constraints added without `NOT VALID` / `VALIDATE CONSTRAINT` on large tables can cause lock escalation.

---

## Step 4 — Output format

Present findings under these headings. Omit a heading if there are no findings for it.

```
## Summary
One paragraph: what the PR does, overall risk level (low / medium / high), and whether it is ready to merge.

## Blockers
Issues that must be fixed before merging (security holes, broken builds, test failures, hard architecture violations).

## Issues
Correctness bugs, missing error handling, test gaps. Should be fixed but won't block an emergency merge.

## Suggestions
Style, clarity, minor improvements. No action required.

## Approved files
List any files that are clean with no findings, if it helps orient the author.
```

Be direct and specific: include file paths and line numbers. Do not praise the author or add filler text. One finding per bullet.
