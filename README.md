# besops

A self-hosted uptime monitoring tool — inspired by [Uptime Kuma](https://github.com/louislam/uptime-kuma), rebuilt from the ground up as a full-stack Go + React application.

## Origin

Uptime Kuma is an excellent, battle-tested monitoring tool with a great community. This project draws heavily from its design philosophy — simple, self-hosted, no agents required — but takes a different path on implementation.

**besops is a ground-up rewrite**, not a fork. The original Node.js/Vue.js codebase served as a reference for feature scope and UX decisions, but no source code was carried over.

## What's different

| | Uptime Kuma | besops |
|---|---|---|
| Backend | Node.js + Socket.IO | Go + OpenAPI (ogen) |
| Frontend | Vue.js 3 | React + Ant Design + TanStack Query |
| API | WebSocket-first, implicit contract | OpenAPI 3.1 spec, generated types, REST + SSE |
| Database | better-sqlite3 (Node binding) | SQLite via `database/sql` + bob query builder |
| Metrics | Internal only | OpenTelemetry export (Prometheus-compatible) |
| Architecture | Monolithic module | Domain-driven packages, composition root |

### Backend

- **Go** with a domain-driven package structure (`internal/domain/{resource}/`)
- **OpenAPI-first**: the API spec drives both server routing (ogen) and client types (openapi-typescript)
- **SQLite** for zero-dependency deployment, with bob for type-safe queries and goose for migrations
- **OpenTelemetry** metrics exporter for monitor status, latency, and uptime
- 11 monitor types: HTTP, TCP, Ping, DNS, Docker, MQTT, gRPC, Redis, SMTP, Push, and Tailscale-Ping
- Notification providers: Webhook, Slack, Discord, Telegram, SMTP, PagerDuty
- Scheduled jobs for uptime aggregation (minutely/hourly/daily), data retention, session cleanup

### Frontend

- **React 19** with Vite, TypeScript strict mode
- **Ant Design** component library with a custom dark/light theme system
- **TanStack Query** for server state with SSE-driven real-time invalidation
- **openapi-fetch** with generated types — every API call is type-checked against the spec
- Playwright E2E tests and Vitest unit tests

## An experiment in AI-assisted development

This project is also a deliberate experiment in building a non-trivial application with AI coding assistants. The entire codebase — backend, frontend, tests, infrastructure — was developed collaboratively with Claude.

The goals of the experiment:

1. **Test the ceiling**: Can AI assistance produce production-quality architecture, not just boilerplate?
2. **Maintain craft standards**: The project enforces TDD, domain separation, and design principles (documented in `.claude/CLAUDE.md`) — can AI follow these consistently?
3. **Measure velocity**: A full-stack rewrite of a mature application is a known-scope problem. How does development speed compare?

This is not a demo or toy project. It follows real engineering practices: test coverage for behavior, error handling at boundaries, continuous integration, and code that's designed to be changed safely tomorrow.

## Quick start

### Prerequisites

- Go 1.26+
- Node.js 22+
- SQLite 3

### Run

```sh
# Backend
cd go
go run ./cmd/server

# Frontend (dev)
cd frontend
npm install
npm run dev
```

The server listens on `:3001` by default. The frontend dev server proxies API requests there.

### Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3001` | HTTP listen port |
| `DATABASE_URL` | `sqlite://./data/besops.db` | SQLite database path |
| `JWT_SECRET` | (required) | Secret for signing auth tokens |
| `OTEL_ENDPOINT` | (disabled) | OpenTelemetry collector gRPC endpoint |
| `KEEP_DATA_PERIOD_DAYS` | `180` | Days to retain heartbeat data |
| `LOG_LEVEL` | `info` | Log verbosity: debug, info, warn, error |

### Tests

```sh
# Go tests
cd go && go test ./...

# Frontend unit tests
cd frontend && npm test

# Frontend E2E tests
cd frontend && npx playwright test
```

## Project structure

```
go/
  cmd/server/          Entry point
  internal/
    app/               Composition root, adapters
    api/               OpenAPI server, security, WebSocket
    auth/              JWT + OIDC + session management
    domain/            One package per resource (monitor, tag, notification, ...)
    monitor/           Check scheduler, manager, result recording
    notification/      Dispatcher + provider registry
    otel/              OpenTelemetry metrics exporter
    jobs/              Scheduled background jobs
  lib/                 Shared types (status, telemetry, errors)

frontend/
  src/
    api/               Generated types + fetch client
    components/        Shared UI components
    contexts/          Auth, theme
    hooks/             React Query hooks (one per domain)
    pages/             Route-level page components
  e2e/                 Playwright tests
```

## Status

Active development. Core monitoring functionality works end-to-end. Some settings pages and advanced features (status pages, maintenance scheduling) are in progress.

## Acknowledgments

[Uptime Kuma](https://github.com/louislam/uptime-kuma) by Louis Lam — for proving that self-hosted monitoring can be simple, beautiful, and community-driven. This project wouldn't exist without that foundation of ideas.

## License

MIT
