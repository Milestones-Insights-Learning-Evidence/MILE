# MILE — Milestones, Insight, Learning & Evidence

**MILE (Milestones, Insight, Learning & Evidence)** is a self-hosted, open-core Intern Management System that turns everyday Kanban task movements into understandable, reviewable, and fair HR performance evidence.

Built as a single Go binary (`mile`) with an embedded static UI (arriving in Phase 4/8) and strict privacy-by-design controls, MILE is engineered so that **Community Edition** is fully deployable, monitorable, and recoverable on its own.

---

## Architecture & Open-Core Boundaries

MILE follows a strict physical and compile-time separation between its MIT-licensed Community core and proprietary Enterprise extensions:

| Path | License | Responsibility |
| --- | --- | --- |
| `cmd/mile/` | MIT | Process entry point (`serve`, `version`, `help`; `migrate`, `backup`, `restore` and others in later phases). Background work runs inside `serve`; there is no separate worker process. |
| `internal/config/` | MIT | Validated environment configuration (`MILE_*` only), timeout bounds, and `_FILE` secret loading. |
| `internal/api/` | MIT | HTTP server, graceful drain lifecycle, security headers, request limits, and RFC 9457 Problem Details. |
| `internal/observability/` | MIT | Structured JSON logging (`log/slog`) and instance-scoped Prometheus metrics (`mile_*`). |
| `internal/enterprise/` | MIT | Compile-time bridge interface (`//go:build !ee` vs `//go:build ee`). Neither stub grants paid capabilities without Phase 9 license verification. |
| `db/`, `ui/` | MIT | SQL migrations/queries (Phase 2) and static SvelteKit frontend (Phase 4). |
| `deploy/`, `docs/` | MIT | Prometheus/Grafana monitoring assets, operational runbooks, ADRs, and governance documentation. See `docs/governance/LICENSING_BOUNDARIES.md` §2 for the exact per-path mapping. |
| `ee/` | Proprietary (`ee/EULA.md`) | Enterprise-only SAML/OIDC, SCIM, PBAC, SIEM, and RS256 `.lic` verification (Phase 9). Never imported by Community builds. |

---

## Implementation Status

| Phase | Scope | Status |
| --- | --- | --- |
| **Phase 0** | Repository foundation, licensing boundaries, ADRs, CI, and strict config loader | Reopened: realignment with plan v4.0 and a first green CI run |
| **Phase 1** | HTTP runtime (`mile serve`), graceful drain, RFC 9457 errors, `/healthz`, `/readyz`, `/metrics`, Prometheus/Grafana baseline | Reopened: re-verification in CI |
| **Phase 2** | PostgreSQL persistence: schema, database roles, append-only protection, migrations, and SQLc adapters | Designed; not yet in the repository |
| **Phase 3** | Authentication (Argon2id, sessions, CSRF, basic OIDC), roles and permissions, and first-run setup | Planned |
| **Phase 4** | First usable workflow slice: the board, task versioning (`409`), and reason prompts on backward moves (`428`) | Planned |
| **Phase 5** | Background work inside the server process: a PostgreSQL-backed job queue, scheduler, and notification outbox | Planned |
| **Phase 6–10** | HR review experience, retention and backup, Community release, Enterprise edition, release automation | Planned |

See [`docs/IMPLEMENTATION_PLAN.md`](docs/IMPLEMENTATION_PLAN.md) for full phase exit gates and specification amendments.

---

## Prerequisites & Building

- **Go**: `1.26+` (see `go.mod`; module `github.com/Milestones-Insights-Learning-Evidence/MILE`)

### Build Community Edition (Default)

```bash
go build -o bin/mile ./cmd/mile
./bin/mile version
# Output: mile dev (community)
```

### Build with Enterprise Bridge Tag

```bash
go build -tags ee -o bin/mile-ee ./cmd/mile
./bin/mile-ee version
# Output: mile dev (enterprise-preview)
```

### Run Tests

```bash
go test ./...
go test -tags ee ./...
```

---

## Running Locally

### CLI Subcommands

- `mile serve` *(default when no arguments are passed)*: Validates configuration and starts the HTTP runtime.
- `mile version`: Prints the version and build edition (`community` or `enterprise-preview`) without loading environment configuration.
- `mile help`: Prints command usage.

### Configuration Reference

MILE validates all settings at startup and fails fast on empty explicit variables or malformed values.

Secret-capable settings support `*_FILE` variants (up to 64 KiB, trailing `\r\n` stripped, mutually exclusive with the direct environment variable). Note that `MILE_LICENSE_FILE` is already a filesystem path and does not use a `_FILE` suffix.

**Only `MILE_*` variables are read.** Generic names such as `DATABASE_URL` or `HTTP_ADDR` are ignored entirely, so another process's environment can never reconfigure MILE: they neither override a `MILE_*` setting nor cause a startup conflict.

| Variable | `_FILE` Support | Default | Description |
| --- | --- | --- | --- |
| `MILE_ENV` | Yes | `development` | Deployment environment: `development`, `test`, or `production`. |
| `MILE_LOG_LEVEL` | Yes | `INFO` | Minimum structured JSON log level: `DEBUG`, `INFO`, `WARN`, or `ERROR`. |
| `MILE_APP_DOMAIN` | Yes | `localhost` | Public hostname (without scheme or path). Must not be `localhost` / loopback when `MILE_ENV=production`. |
| `MILE_HTTP_ADDR` | Yes | `127.0.0.1:8000` | TCP `host:port` bind address (`1–65535`). |
| `MILE_DATABASE_URL` | Yes | *(empty)* | PostgreSQL connection URL (`postgres://` or `postgresql://`). Validated when set; required starting in Phase 2. |
| `MILE_LICENSE_FILE` | No *(is a path)* | `/etc/mile/mile_enterprise.lic` | Path to the Enterprise `.lic` file (verified in Phase 9). |
| `MILE_METRICS_TOKEN` | Yes | *(empty)* | Bearer token protecting `/metrics`. Must be `>= 32` non-whitespace characters when `MILE_ENV=production` or when `MILE_HTTP_ADDR` binds beyond loopback. |
| `MILE_HTTP_READ_TIMEOUT_SEC` | Yes | `10` | Maximum duration (seconds) for reading an entire request. |
| `MILE_HTTP_READ_HEADER_TIMEOUT_SEC` | Yes | `5` | Maximum duration (seconds) allowed to read request headers. |
| `MILE_HTTP_WRITE_TIMEOUT_SEC` | Yes | `15` | Maximum duration (seconds) before timing out writes of the response. |
| `MILE_HTTP_IDLE_TIMEOUT_SEC` | Yes | `60` | Maximum duration (seconds) to wait for the next request on keep-alive connections. |
| `MILE_HTTP_SHUTDOWN_TIMEOUT_SEC` | Yes | `10` | Grace period (seconds) to drain active HTTP requests on `SIGINT` / `SIGTERM`. |
| `MILE_HTTP_MAX_BODY_BYTES` | Yes | `1048576` | Maximum allowed request body size in bytes (default 1 MiB). |

---

## Smoke Checks (`/healthz`, `/readyz`, `/metrics`)

With `mile serve` running on the default loopback address (`127.0.0.1:8000`):

```bash
# 1. Process liveness check
curl -i http://127.0.0.1:8000/healthz
# HTTP/1.1 200 OK
# {"status":"ok"}

# 2. Readiness check (discloses Phase 1 scope)
curl -i http://127.0.0.1:8000/readyz
# HTTP/1.1 200 OK
# {"scope":"http-runtime","status":"ready"}

# 3. Prometheus metrics (unauthenticated on loopback in development when MILE_METRICS_TOKEN is unset)
curl -s http://127.0.0.1:8000/metrics | grep -E "^mile_(build_info|process_start_time_seconds|http_)"
```

When `MILE_METRICS_TOKEN` is configured:

```bash
curl -s -H "Authorization: Bearer $MILE_METRICS_TOKEN" http://127.0.0.1:8000/metrics | head -n 25
```

---

## Documentation Index

- **Implementation Plan & Amendments**: [`docs/IMPLEMENTATION_PLAN.md`](docs/IMPLEMENTATION_PLAN.md)
- **Change History**: [`docs/CHANGELOG.md`](docs/CHANGELOG.md)
- **Operations & Monitoring Runbook**: [`docs/operations/monitoring.md`](docs/operations/monitoring.md)
- **Monitoring Assets**:
  - Prometheus scrape config: [`deploy/monitoring/prometheus.yml`](deploy/monitoring/prometheus.yml)
  - Prometheus alert rules: [`deploy/monitoring/alerts.yml`](deploy/monitoring/alerts.yml)
  - Grafana dashboard (`mile-phase1-runtime`): [`deploy/monitoring/grafana-dashboard.json`](deploy/monitoring/grafana-dashboard.json)
- **Architecture Decision Records**: [`docs/adr/`](docs/adr/)
- **Governance & Licensing**: [`docs/governance/`](docs/governance/), [`LICENSE`](LICENSE), [`CONTRIBUTING.md`](CONTRIBUTING.md), [`SECURITY.md`](SECURITY.md)
- **Source Specifications**: [`docs/specs/`](docs/specs/)
