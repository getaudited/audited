# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`audited` is an audit log management service for cloud-native applications. It exposes an HTTP API (Echo + oapi-codegen) and is backed by PostgreSQL.

## Commands

This project uses [Task](https://taskfile.dev) (`task`) as the task runner. All environment variables are loaded from `.env.local` (overrides) and `.env`.

### Infrastructure

```bash
task run          # Start postgres + rabbitmq + app in Docker
task dev          # Also starts swagger-ui on :8090
```
In 
### Running the service locally (outside Docker)

```bash
task mig:up       # Apply DB migrations (goose)
task run:service  # Run the service with live-reload (reflex)
```

### Testing

```bash
task test               # Unit tests: go test -v -race ./internal/...
task test:components    # Component/integration tests against a running service: ./tests/components/...
task test:all           # Both of the above
```

To run a single test:
```bash
go test -v -race ./internal/... -run TestName
```

Component tests require the service and its dependencies to be running (they hit `http://localhost:8080`).

### Linting & Formatting

```bash
task lint         # golangci-lint run
task lint:fix     # golangci-lint run --fix
task fmt          # gofmt -w -s ./
```

### Code Generation

```bash
task openapi      # Regenerate HTTP server + test client from api/openapi.yml (oapi-codegen)
task orm          # Regenerate SQLBoiler models from DB (resets + re-migrates first)
task events       # Regenerate protobuf code from events/events.proto
```

### Database Migrations

```bash
task mig:up           # Apply all pending migrations
task mig:down         # Roll back last migration
task mig:reset        # Roll back all migrations
task mig:create NAME  # Create a new sequential SQL migration file
task seed:up          # Apply seeds (non-versioned)
task seed:down        # Roll back seeds
```

### Setup (first time)

```bash
task setup    # Install required Go tools: sqlboiler, goose, oapi-codegen, golangci-lint
```

## Architecture

The codebase follows a **ports & adapters** (hexagonal) pattern strictly separated into:

```
internal/
  domain/       # Pure domain types (Event, EventType, Actor, Target) and repository interfaces
  app/
    command/    # Command handlers (write operations: CreateEvent, CreateEventType)
    query/      # Query handlers (read operations: EventTypeByAction, etc.)
    app.go      # App struct wiring Commands + Queries; handler interfaces defined here
  adapters/
    models/     # Auto-generated SQLBoiler ORM models — do not edit manually
    psql/       # PostgreSQL implementations of domain repository interfaces + mappers
  ports/
    http/       # Echo HTTP server; server.gen.go is auto-generated from openapi.yml — do not edit
  common/       # Shared utilities (logging, postgres connection, wait helpers)
```

### Key design rules

- **Domain layer has zero infrastructure dependencies.** `internal/domain/` defines interfaces; adapters implement them.
- **`app.App`** is the central wiring struct passed to the HTTP port. Commands/Queries are registered there in `cmd/service/main.go`.
- **`server.gen.go`** and `internal/adapters/models/` are generated files. Regenerate with `task openapi` and `task orm` respectively; never edit by hand.
- **Migrations** live in `misc/sql/migrations/` (goose sequential format). Seeds are in `misc/sql/seeds/`.
- **Component tests** (`tests/components/`) test the running service end-to-end using the auto-generated API client (`tests/client/client.gen.go`). They rely on `wait_for` helpers to poll for service readiness before running.

### Environment variables (key ones)

| Variable | Purpose |
|---|---|
| `DATABASE_URL` | PostgreSQL DSN |
| `HTTP_PORT` | HTTP listen port (default 8080) |
| `AMQP_URL` | RabbitMQ connection URL |
| `DEBUG_MODE` | Enables verbose logging |
| `GOOSE_DRIVER` / `GOOSE_DBSTRING` | Used by goose CLI for migrations |

The service auto-applies migrations on startup via `postgres.ApplyMigrations`.

### OpenAPI

The API contract lives in `api/openapi.yml`. The HTTP server (`internal/ports/http/server.gen.go`) and the test client (`tests/client/client.gen.go`) are both generated from this single source. Update the spec first, then run `task openapi`.
