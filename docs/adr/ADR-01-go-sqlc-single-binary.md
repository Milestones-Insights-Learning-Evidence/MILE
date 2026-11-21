# ADR-01: Go + SQLc + Embedded SvelteKit Single-Binary Architecture

- **Status:** Accepted (Amended by `docs/IMPLEMENTATION_PLAN.md` v2.1). `D3` withdrawn and §2.1 item 5 amended on 2026-11-20 to match plan v4.0 (DEC25); see §2.2. `D1` amended on 2026-11-21 (DEC83): the product, module path, binary, environment prefix and metric prefix were renamed from MOAJ to MILE; §2.1 and §2.2 use the new names.
- **Date:** 2026-11-01
- **References:** `IMS-BRD-1.0` (`NFR-01`, `NFR-02`, `NFR-03`), `IMS Full System Architecture.pdf` (Blueprint v1.0 §1–§2), `docs/IMPLEMENTATION_PLAN.md` (§3.1, §6.1, §6.2, §6.3, Appendix A `D1`–`D3`)

---

## 1. Context

**MILE (Milestones, Insight, Learning & Evidence)** targets mid-market engineering leads, mentors, and HR teams that require a self-hosted workflow and intern appraisal platform without complex microservice orchestration or heavy runtime memory overhead.

The source specifications (`IMS-BRD-1.0` and Blueprint v1.0) establish the following architectural goals:
- **`NFR-01` (Single-Binary Deployment):** Ship the HTTP API, asynchronous background worker, database migration runner, and static web frontend inside a single compiled executable (`mile`) with no Node.js runtime required in production.
- **`NFR-02` (Relational + JSONB Persistence):** Use PostgreSQL 16 as the primary ACID relational database for strict workflow state transitions and append-only event ledgers, paired with `JSONB` columns for department-customizable 30/60/90-day evaluation rubrics.
- **`NFR-03` (Permissive Open-Source Stack):** Ensure all runtime dependencies and infrastructure components carry permissive open-source licenses (`MIT`, `BSD`, `Apache-2.0`, `PostgreSQL`) so the codebase can be distributed under both the root MIT License (Community Edition) and the commercial Enterprise EULA (`ee/EULA.md`).

During Phase 0 architecture review (`docs/IMPLEMENTATION_PLAN.md`), three additional foundation decisions (`D1`–`D3`) and a key operational scope amendment (Community observability) required formal recording.

---

## 2. Decision

### 2.1 Core Application & Persistence Stack
1. **Language & Runtime:** Use **Go** (`go 1.26.0` as the minimum in `go.mod`, with CI selecting the latest `1.26.x` patch, Decision `D2`) with the standard library `net/http` router, `log/slog` structured JSON logger, and `embed` package.
2. **Project Name & Module Path (`D1`):** Adopt the official project name **MILE (Milestones, Insight, Learning & Evidence)** across documentation, CLI (`cmd/mile`, `bin/mile`, `bin/mile-ee`), environment variables (`MILE_*` only; no unprefixed, `IMS_*`, or `MINTERN_*` aliases), enterprise license path (`/etc/mile/mile_enterprise.lic`), Prometheus metrics (`mile_*`), and the Go module path `github.com/Milestones-Insights-Learning-Evidence/MILE` in `go.mod` matching `https://github.com/Milestones-Insights-Learning-Evidence/MILE.git`.
3. **Database Access (`SQLc` + `pgx/v5`):** Write explicit, parameterized SQL queries in `db/query/*.sql` and compile them into type-safe Go code in `internal/db/` via **SQLc** (`sqlc.yaml`), rather than using reflection-heavy ORMs (`GORM`, `Ent`). Schema migrations live in `db/migrations/*.sql` and are embedded into the binary via `goose`.
4. **Embedded Frontend (`SvelteKit` Static Adapter):** Build the frontend in `ui/` (`mile-ui`) using **SvelteKit** with `@sveltejs/adapter-static`, embedding the compiled static bundle directly into the Go binary (`NFR-01`).
5. **Single Binary Subcommands (`cmd/mile`):** The `mile` executable provides explicit operational entry points (`mile serve`, `mile version`, `mile help`, and later `migrate`, `backup`, `restore`, `verify-backup`, `rotate-master-key`, `support-bundle`) from one binary artifact. *Amended 2026-11-20 (DEC25):* background work runs inside the `serve` process; there is no separate worker process or `worker` subcommand.
6. **Open-Core Build-Tag Boundary (`!ee` vs. `ee`):** Isolate proprietary enterprise features (`SAML`, `OIDC`, `SCIM`, `PBAC`, `SIEM`, `RS256` licensing) inside `ee/` behind `//go:build ee` constraints and the `internal/enterprise` bridge interface so the default Community build (`//go:build !ee`) never imports `ee/`.

### 2.2 Resolution of `D1`, `D2`, and `D3`
- **`D1` (Project Naming & Module Identity):** Official project name is **MILE - Milestones, Insight, Learning & Evidence** (`cmd/mile`, `bin/mile`, `MILE_*` env vars only, `mile_*` metrics) with Go module path `github.com/Milestones-Insights-Learning-Evidence/MILE` matching the GitHub organization remote.
- **`D2` (Go Version & CI Matrix):** Declare `go 1.26.0` as the minimum in `go.mod` (no `toolchain` line) while CI resolves the toolchain with `go-version: 1.26.x` so security patches are picked up, and enforce Linux CI verification (`gofmt`, `go vet`, `staticcheck`, `govulncheck`, `go test -race` for both Community and `-tags ee` builds).
- **`D3` — Withdrawn on 2026-11-20 (DEC25, plan v4.0 Appendix A).** MILE has no Redis-family dependency. Background jobs, scheduling and the notification outbox run in PostgreSQL, inside the server process (ADR-05, to be written in Phase 0). Deployment is one binary plus PostgreSQL. The original decision is kept below, superseded, for the record:
  - *Superseded text:* Where the original Blueprint referenced Redis 7 alongside `hibiken/asynq`, Redis 7.4+ switched to non-permissive `RSALv2`/`SSPLv1` (and Redis 8 added `AGPLv3`), conflicting with `NFR-03`. MILE adopts **Valkey** (`BSD-3-Clause`, `valkey/valkey:8.0-alpine`) as its intended in-memory store for `asynq` queues and scheduled tasks. Valkey is a `BSD-3-Clause` fork of Redis 7.2.4 that retains the Redis wire protocol per its own project documentation; whether `asynq` works against it for the specific queue and scheduler features MILE needs is **not yet tested** and is a Phase 5 integration-test obligation, not an established property of this decision.

---

## 3. Specification Amendments Recorded in This ADR

### 3.1 Community Monitoring & Operability Amendment (`docs/IMPLEMENTATION_PLAN.md` §3.1 & §6.3)
- **Original Specification Gap:** The initial specifications treated external telemetry mainly as an Enterprise SIEM feature (`BR-EE-04`) and deferred operational packaging until the end of development.
- **Amended Rule:** Prometheus-compatible `/metrics` (`mile_build_info`, `mile_process_start_time_seconds`, `mile_http_requests_total`, `mile_http_request_duration_seconds`, `mile_http_requests_in_flight`, protected by `MILE_METRICS_TOKEN` bearer authentication), structured JSON logs (`log/slog`), process liveness (`/healthz`) vs. dependency readiness (`/readyz`), a versioned Grafana dashboard (`deploy/monitoring/grafana-dashboard.json`, `uid: mile-phase1-runtime`), Prometheus scrape and alert rules (`deploy/monitoring/prometheus.yml`, `deploy/monitoring/alerts.yml`), and operational runbooks (`docs/operations/monitoring.md` §5) are **core Community Edition features** starting in Phase 1.
- **Privacy Invariant on Telemetry:** Metric labels must remain strictly bounded (route templates, HTTP methods, status codes, operation names) and must never include `user_id`, `intern_id`, `task_id`, email addresses, or raw URLs.

---

## 4. Consequences

### Positive
- **Zero Production Runtime Bloat:** Operators deploy a single statically linked Go binary (`mile`) plus PostgreSQL 16, with no Node.js or external ORM runtime dependencies.
- **Compile-Time SQL & Boundary Safety:** SQL query mismatches fail at code-generation/compile time, and `make boundary-check` proves the Community binary contains zero proprietary `ee/` packages.
- **License Cleanliness (`NFR-03`):** Depending on no Redis-family service and enforcing `make license-check` eliminates copyleft/SSPL contamination risks across both Community and Enterprise builds.
- **Day-One Operability:** Community operators gain immediate visibility into HTTP latency, expected `409`/`428` workflow responses vs. `5xx` faults, database pool health, and worker outbox backlogs.

### Trade-offs & Mitigations
- **SQLc Boilerplate for Dynamic Queries:** Complex analytical queries require carefully authored SQL in `db/query/` rather than dynamic query builders; mitigated by PostgreSQL 16 CTEs, window functions, and explicit repository interfaces in `internal/`.
- **Measured Footprint Verification (Phase 8 Gate):** Source targets (`<20 MB` binary, `<25 MB` idle RSS, `<5 min` install-to-first-board) will be explicitly benchmarked under documented conditions in Phase 8 rather than assumed without measurement.
