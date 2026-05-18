# Domain Glossary

Ubiquitous language for this codebase. When naming variables, functions, UI labels, or API fields, use these terms consistently.

---

## Core entities

### Monitor

A configured check that runs on a schedule. Each monitor has a **type** (protocol), a **target** (URL/host/port), and an **interval**.

- **Active** — whether the scheduler should run checks for this monitor.
- **Interval** — seconds between checks.
- **Timeout** — max seconds a single check may take before reporting Down.
- **MaxRetries** — how many consecutive failures before transitioning to Down.
- **RetryInterval** — seconds between retry attempts (within the retry window).
- **UpsideDown** — inverts status: a successful check is reported as Down, failure as Up.
- **ParentID** — for group monitors, the parent that aggregates child statuses.

### Monitor types

| Type | What it checks |
|------|---------------|
| `http` | HTTP/HTTPS endpoint, status code + optional keyword/JSON path |
| `port` (TCP) | TCP connection to host:port |
| `ping` | ICMP echo |
| `keyword` | HTTP response body contains (or doesn't contain) a string |
| `json-query` | HTTP response body evaluated via JSON path |
| `grpc-keyword` | gRPC call response |
| `dns` | DNS resolution (A, AAAA, MX, etc.) |
| `push` | Passive — waits for an external system to POST a heartbeat |
| `mqtt` | MQTT subscribe, wait for message |
| `redis` | Redis PING |
| `group` | Aggregates child monitors; Up if all children are Up |
| `tailscale-ping` | Tailscale DERP ping |
| `snmp` | SNMP query |
| `sqlserver` / `postgres` / `mysql` / `mongodb` | Database connectivity via query |
| `radius` | RADIUS authentication |
| `steam` / `gamedig` | Game server query |
| `rabbitmq` | RabbitMQ connectivity |
| `real-browser` | Headless browser page load |
| `manual` | Status set manually by user (no automated check) |

### Heartbeat

A single check result recorded at a point in time.

| Field | Meaning |
|-------|---------|
| `Status` | 0 = Down, 1 = Up, 2 = Pending, 3 = Maintenance |
| `Ping` | Response time in milliseconds (null for non-applicable types) |
| `Msg` | Human-readable result message |
| `Important` | True when this heartbeat represents a status *transition* |
| `Duration` | Seconds the monitor has been in this status continuously |
| `Retries` | How many retry attempts preceded this result |

### Status

An integer enum representing the current state of a monitor:

| Value | Name | Meaning |
|-------|------|---------|
| 0 | **Down** | Check failed (after retries exhausted) |
| 1 | **Up** | Check succeeded |
| 2 | **Pending** | Initial state before first check completes |
| 3 | **Maintenance** | Monitor is inside a maintenance window; checks are suppressed |

### Notification

A configured alert channel. Notifications are **linked to monitors** via a many-to-many relationship (`MonitorNotification`). When a monitor transitions status, the dispatcher sends to all active notifications linked to that monitor.

- **Config** — JSON blob containing provider-specific settings (webhook URL, Slack channel, etc.).
- **Active** — whether this notification channel is enabled.

### Notification providers

`webhook`, `slack`, `discord`, `telegram`, `smtp`, `pagerduty`

### Tag

A user-defined label (name + color) that can be attached to monitors via `MonitorTag`. Tags carry an optional **value** per monitor (e.g., tag "environment" with value "prod").

### Proxy

An HTTP or SOCKS proxy that monitors can route outbound checks through.

- **Protocol** — `http`, `https`, `socks`, `socks5`
- **Default** — if true, applies to all monitors that don't specify a proxy

### Maintenance

A scheduled window during which monitors are suppressed (status = Maintenance, no alerts fire).

#### Maintenance strategies

| Strategy | Meaning |
|----------|---------|
| `manual` | User manually activates/deactivates |
| `single` | One-time window between start and end datetime |
| `recurring-interval` | Repeats every N days for a fixed duration |
| `recurring-weekday` | Repeats on specific weekdays within a time range |
| `recurring-day-of-month` | Repeats on specific days of the month |
| `cron` | Cron expression triggers maintenance for a fixed duration |

### Status Page

A public-facing page showing the health of selected monitors.

- **Slug** — URL path segment (e.g., `/status/my-service`)
- **Published** — whether the page is publicly accessible
- **Group** — a named section within a status page containing monitors (ordered by `weight`)
- **Incident** — a user-authored notice pinned to a status page (active/resolved, styled as info/warning/danger/success)

---

## Infrastructure concepts

### Checker

A Go interface (`internal/monitor/types/`) that knows how to perform one type of check. Returns a `CheckResult` with status, ping, and message.

### Manager

The orchestrator (`internal/monitor/`) that maintains a running goroutine per active monitor, schedules checks, records heartbeats, and triggers notifications on transitions.

### Scheduler (jobs)

Background cron-like runner (`internal/jobs/`) for periodic housekeeping: stat aggregation, data retention, session cleanup, SQLite vacuum.

### Hub

WebSocket broadcast hub (`internal/broadcast/`). Publishes real-time events (heartbeats, status changes) to connected dashboard clients.

### Uptime calculator

Computes percentage uptime from heartbeat and stat records over configurable windows (minutely, hourly, daily).

---

## Stat tables

Heartbeats are aggregated into rollup tables for efficient historical queries:

| Table | Granularity | Fields |
|-------|-------------|--------|
| `stat_minutely` | 1 minute | up count, down count, avg/min/max ping |
| `stat_hourly` | 1 hour | same |
| `stat_daily` | 1 day | same |

---

## Relationships (key many-to-many joins)

| Join table | Connects | Purpose |
|------------|----------|---------|
| `monitor_notification` | Monitor ↔ Notification | Which alerts fire for which monitors |
| `monitor_tag` | Monitor ↔ Tag | Labeling, with per-monitor value |
| `monitor_group` | Monitor ↔ Group | Status page grouping |
| `monitor_maintenance` | Monitor ↔ Maintenance | Which monitors a maintenance window suppresses |
| `maintenance_status_page` | Maintenance ↔ StatusPage | Which status pages show a maintenance notice |

---

## Naming conventions in code

| Domain concept | Go package | API path | Frontend hook |
|----------------|-----------|----------|---------------|
| Monitor | `domain/monitor` | `/monitors` | `useMonitors` |
| Heartbeat | `domain/heartbeat` | `/monitors/{id}/heartbeats` | `useHeartbeats` |
| Notification | `domain/notification` | `/notifications` | `useNotifications` |
| Maintenance | `domain/maintenance` | `/maintenance` | `useMaintenance` |
| Tag | `domain/tag` | `/tags` | `useTags` |
| Proxy | `domain/proxy` | `/proxies` | `useProxies` |
| Status Page | `domain/statuspage` | `/status-pages` | `useStatusPages` |
| Settings | `domain/settings` | `/settings` | `useSettings` |
| API Key | `domain/apikey` | `/api-keys` | `useAPIKeys` |
