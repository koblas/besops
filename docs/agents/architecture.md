# Architecture

System map and module boundaries for the uptime monitoring tool.

## Stack

| Layer | Tech | Entrypoint |
|-------|------|-----------|
| Backend | Go 1.26, SQLite, [ogen](https://ogen.dev) (OpenAPI codegen), [bob](https://bob.stephenafamo.com) (query builder) | `go/cmd/uptime/` |
| Frontend | React 19, Ant Design 6, TanStack Query 5, Vite | `frontend/src/App.tsx` |
| API contract | OpenAPI 3.1 | `go/api/openapi.yaml` |
| Migrations | golang-migrate, SQLite | `go/internal/database/migrations/` |

## High-level layout

```
go/
├── cmd/uptime/          # main binary
├── cmd/migrate/         # migration CLI
├── api/openapi.yaml     # source of truth for API
├── internal/
│   ├── api/             # composed handler, security, websocket, OAS codegen
│   ├── app/             # application bootstrap & wiring (composition root)
│   ├── auth/            # OIDC, JWT, sessions, 2FA
│   ├── broadcast/       # WebSocket event hub
│   ├── database/        # open, migrate helpers
│   ├── domain/          # ← all business logic lives here (see below)
│   ├── jobs/            # scheduled background work
│   ├── monitor/         # check-loop manager & registry
│   ├── notification/    # dispatcher, provider registry
│   ├── otel/            # OpenTelemetry metrics exporter
│   ├── proxy/           # outbound proxy support
│   ├── tls/             # certificate utilities
│   └── uptime/          # uptime percentage calculator
├── models/              # generated bob ORM models (DO NOT EDIT)
├── factory/             # generated test factories (DO NOT EDIT)
├── dbinfo/              # generated DB metadata (DO NOT EDIT)
└── dberrors/            # generated constraint-error helpers (DO NOT EDIT)

frontend/src/
├── api/                 # openapi-fetch client + generated types
├── components/          # shared UI components
├── contexts/            # React contexts (theme, auth)
├── hooks/               # React Query hooks per domain
├── layouts/             # shell layouts (app, public, empty)
└── pages/               # route-level pages grouped by feature
```

## Domain packages (`go/internal/domain/`)

Each domain is a self-contained package following this structure:

| File | Purpose |
|------|---------|
| `model.go` | Plain domain structs (no DB/OAS imports) |
| `repository.go` | Consumer-side interface |
| `sqlite_repo.go` | Bob-based implementation |
| `handler.go` | OAS handler methods (returns OAS types directly) |

**Domains:** `apikey`, `badge`, `heartbeat`, `maintenance`, `monitor`, `notification`, `proxy`, `settings`, `stats`, `statuspage`, `system`, `tag`, `user`

Cross-domain dependencies use local interfaces injected at the composition root (`internal/app/app.go`), never direct package imports between domains.

## Monitor check loop

`internal/monitor/` owns the scheduling loop:

1. **Manager** loads active monitors, starts/stops per-monitor goroutines.
2. **Registry** maps monitor type strings → `Checker` implementations.
3. **Checkers** (in `internal/monitor/types/`): `http`, `tcp`, `ping`, `dns`, `grpc`, `mqtt`, `redis`, `push`, `smtp`, `tailscale-ping`, `group`.
4. Each check produces a heartbeat → stored → broadcast via WebSocket hub → notification dispatcher fires if status changed.

## Notification dispatch

`internal/notification/` dispatches alerts:

- **Registry** maps provider type → `Notifier` implementation.
- **Providers** (`internal/notification/providers/`): `webhook`, `slack`, `discord`, `telegram`, `smtp`, `pagerduty`.
- **Dispatcher** receives monitor status transitions, loads notification rules for the monitor, and fans out.

## Scheduled jobs (`internal/jobs/`)

| Job | Cadence | Purpose |
|-----|---------|---------|
| Aggregate minutely/hourly/daily | 1m / 1h / 1d | Roll up heartbeats into `stat_*` tables |
| Clear old data | daily | Prune heartbeats older than retention window |
| Clear expired sessions | hourly | Remove stale auth sessions |
| Vacuum | daily | SQLite VACUUM |

## API layer

- `go/api/openapi.yaml` is the single source of truth.
- `go generate ./internal/api/` runs ogen to produce `internal/api/oas_generated/`.
- `ComposedHandler` delegates to domain handlers (functional-option injection).
- Security is handled by `SecurityHandler` (JWT/session validation).
- Frontend types: `npm run generate-api` in `frontend/` runs `openapi-typescript`.

## Data flow (request lifecycle)

```
HTTP request
  → ogen router (oas_generated)
  → SecurityHandler (auth check)
  → ComposedHandler
  → domain Handler (business logic)
  → Repository interface → sqlite_repo (bob queries)
  → response marshalled by ogen
```

## Code generation (regeneration commands)

From `go/`:
- **OAS server**: `go generate ./internal/api/` (requires openapi.yaml changes)
- **Bob models**: `go generate ./internal/database/bobgen/` (requires running DB with current schema; `make bobgen` runs migrate-up first)
- **Frontend types**: `cd ../frontend && npm run generate-api`

## Key conventions

- Generated files (`models/`, `factory/`, `dbinfo/`, `dberrors/`, `internal/api/oas_generated/`, `frontend/src/api/generated/`) are committed but never hand-edited.
- Nullable DB columns map to `*T` in domain models and `omitnull.Val[T]` in setters.
- IDs are UUIDs stored as TEXT in SQLite.
- The composition root (`app.go`) is the only place that knows about all domains simultaneously.
