# SpendWise Backend (`spendwise-be`)

A self-hostable personal finance API. This is the **Go + Gin REST backend** for [SpendWise](https://github.com/aayushsiwa/spendwise-fe), responsible for storing financial data, computing summaries and balances, and serving the web dashboard and MCP server.

Related repositories:

- Frontend dashboard: <https://github.com/aayushsiwa/spendwise-fe>
- MCP server for AI assistants: <https://github.com/aayushsiwa/spendwise-mcp>

## Features

- **Transaction records** — create, read, update, and delete income, expense, and transfer records. Listing supports pagination, filtering (date range, category, type, amount), case-insensitive description search, and optional grouping (by category or month).
- **Categories** — manage categories with custom names, icons, and colors; used for classification and analytics.
- **Budgets** — set per-category budgets by month and year, and query budget progress (spent vs. budgeted, percentage consumed).
- **Savings goals** — define goals with target/optional current amounts, target dates, and statuses; contribute progress over time.
- **Financial summaries** — aggregated income, expense, net, opening/closing balances, and category breakdowns for a date range. Rebuilt transactionally whenever records change.
- **Running balance** — balances are recomputed as a running total (window function) on every record mutation, with a manual `/refresh` endpoint.
- **Import / export** — flexible CSV import (alias-aware column mapping; requires at least `date` and `amount`) and JSON import, plus CSV export for portable data.
- **Multi-database support** — SQLite by default (WAL mode), with Postgres and MySQL available via `DB_TYPE` / `DB_URL`.
- **Health check** — a `/health` endpoint reports database connectivity.
- **Consistent errors** — all errors use an `{"error": {"type", "message", "details"}}` envelope.

> **Note:** Authentication/authorization is not yet implemented. For personal/local use this is fine; before exposing the port, add auth and tighten CORS (currently `AllowOrigins: ["*"]`).

## Architecture

Layered design with a clear request flow:

```text
routes/  →  handlers/  →  services/  →  db/ (GORM)
```

- **`main.go`** — entrypoint: loads `.env`, initializes the database, wires middleware and routes, runs with graceful shutdown.
- **`routes/`** — declarative route table mapping HTTP method + path to handlers (all mounted under `/api/v1`).
- **`handlers/`** — HTTP layer: request parsing, validation, calling services, and response shaping. One file per resource (e.g. `createrecord.go`, `getbudgets.go`).
- **`services/`** — business logic: records, categories, budgets, goals, summaries, balance, and import. Talks to the database via GORM.
- **`models/`** — GORM models and request/query structs (`record`, `category`, `budget`, `goal`, `summary`, `queryparams`, `update_request`).
- **`db/`** — database initialization, connection, and health check (SQLite/Postgres/MySQL). Schema is managed by GORM `AutoMigrate` on startup; `sql/` holds manual reset scripts.
- **`middleware/`** — panic recovery, security headers, and validation error handling.
- **`errors/`** — typed error definitions and the error envelope.
- **`config/`** — environment-based configuration loading.
- **`validation/`**, **`utils/`** — input validation and helpers (date parsing, month math, IDs).
- **`mocks/`** — mock service used in handler tests.

## API surface (under `/api/v1`)

| Group | Endpoints |
|---|---|
| Records | `GET/POST /records`, `GET/PATCH/DELETE /records/:id` |
| Summary | `GET /summary` |
| Import / Export | `POST /import/csv`, `POST /import/json`, `GET /export/csv` |
| Categories | `GET/POST /categories`, `PATCH/DELETE /categories/:id` |
| Budgets | `GET/POST /budgets`, `GET /budgets/progress`, `PATCH/DELETE /budgets/:id` |
| Goals | `GET/POST /goals`, `GET/PATCH/DELETE /goals/:id`, `POST /goals/:id/progress` |
| Maintenance | `POST /refresh`, `GET /health` |

## Getting started

### Prerequisites

- Go 1.26+ (toolchain per `go.mod`).
- SQLite (default), or Postgres/MySQL for other `DB_TYPE`s.
- The [`sqlite3`](https://www.sqlite.org/cli.html) CLI for migration targets.

### Run

```sh
cp .env.example .env
make dev        # hot-reload via Air, falls back to `go run .`
```

The server listens on **port 8090** by default (override with `PORT`) with all routes under `/api/v1`.

### Environment variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8090` | HTTP port. |
| `GIN_MODE` | `release` | Gin mode; empty falls back to `release`. |
| `DB_TYPE` | `sqlite` | `sqlite`, `postgres`, or `mysql`. |
| `DB_URL` | _(empty)_ | Connection string; for SQLite falls back to `records.db`. |

### Make targets

| Command | What |
|---|---|
| `make dev` | Dev server with Air hot reload. |
| `make test` | `go test ./...` |
| `make coverage` | Tests + coverage report. |
| `make lint` | `golangci-lint`. |
| `make check` | `go fmt ./...` then `go vet ./...` |
| `make modernize` | Go `modernize` fix. |
| `make pre-push` | modernize → lint → check → test → coverage-check. |
| `make migrationup` / `make migrationdown` | Apply/revert `sql/init.sql` / `sql/dropData.sql` (destructive: `migrationup` deletes the SQLite DB first). |
| `make build` | Cross-compile linux amd64 (`CGO_ENABLED=0`). |
| `make run` | Build Docker image and run the container. |

> `.githooks/pre-push` runs `make pre-push`. Enable with `git config core.hooksPath .githooks`.

## Testing

Backend tests are extensive — handlers use `httptest` + `mocks.MockService`, services use in-memory SQLite for integration-style tests, and `utils`/`validation` have their own suites. Run them with `make test` or `make coverage` (enforces a minimum coverage threshold).
