# MOAJ (Mentor of all Jacks) - Change History

This file records what the repository *was*, so that
[`IMPLEMENTATION_PLAN.md`](IMPLEMENTATION_PLAN.md) can describe only the intended end
state and the gates still to be met. Entries are historical: they are accurate as of
their stated date and are **not** descriptions of the current tree. Always recheck the
working tree before an implementation step rather than relying on a snapshot here.

---

## Pre-implementation baseline - 2026-10-31

Repository inspection before Phase 0 began. Every row below has since been superseded;
the "Superseded by" column records where the current behaviour lives.

| Area | State on 2026-10-31 | Superseded by |
| --- | --- | --- |
| Git | Initialized; single commit `b4afedd` (`first commit`). | Phase 0 and Phase 1 gate commits. |
| Go | `module mintern`, `go 1.22`; local toolchain `go1.27.2 windows/amd64`. | `module github.com/MOAJ-Mentor-of-all-Jacks/MOAJ`, `go 1.26.0` floor in `go.mod`, CI on `1.26.x`. |
| Entry point | `cmd/ims/main.go` held practice TCP code with syntax and identifier errors. | `cmd/moaj/main.go` (`serve`, `version`, `help`). |
| Config | Formatting, validation, secret loading, and a `.llc` license extension needed correction. | `internal/config` with strict validation, `_FILE` secrets, and the `/etc/moaj/moaj_enterprise.lic` default. |
| Empty Go files | 14 zero-byte files; `go build ./...` failed. | Zero-byte files removed in Phase 0; both build variants compile. |
| Infrastructure and UI | Scaffolds only - no migrations, queries, CI, container setup, or SvelteKit. | Phase 0 CI and `deploy/`; persistence is Phase 2 and the UI is Phase 4. |
| Documentation and licensing | README, CONTRIBUTING, root LICENSE, ADRs, and `ee/EULA.md` were empty. | Populated in Phase 0; see `docs/governance/`. |
| Specifications | Root PDFs duplicated `docs/specs/` byte-for-byte. | Root copies removed; `docs/specs/` is canonical. |

---

## Phase 0 and Phase 1 - 2026-11-01

Gate verification records for both completed phases are kept inline with their phase
sections in [`IMPLEMENTATION_PLAN.md`](IMPLEMENTATION_PLAN.md) §7, since each one is
the evidence for that phase's exit gate rather than a change description.

---

## Audit remediation - 2026-11-07

Findings F1-F11 from the 2026-11-06 review, applied before the first push to the remote:

| ID | Change |
| --- | --- |
| F1 | Grafana GC panel now queries the quantiles `client_golang` actually exports (`0.75` and `1`) instead of a non-existent `quantile="0.99"`. |
| F2 | **Breaking:** all unprefixed (`DATABASE_URL`, `HTTP_ADDR`, `REDIS_ADDR`, `APP_DOMAIN`, `HTTP_*_TIMEOUT_SEC`), `IMS_*`, and `MINTERN_LICENSE_FILE` aliases removed. Only `MOAJ_*` and its `_FILE` variants are read. Nothing had shipped under an earlier name, and an ambient `DATABASE_URL` alongside `MOAJ_DATABASE_URL` previously halted startup as "mutually exclusive". |
| F3 | CI selects Go with `go-version: 1.26.x` (`check-latest: true`) instead of `go-version-file: go.mod`, which pinned exactly `1.26.0`. |
| F4 | The PR-body CLA check is renamed to an advisory acknowledgement prompt and no longer claims to verify a signature; real enforcement stays blocked on the hosted CLA bot under `D17`. |
| F5 | `deploy/Dockerfile`: `EXPOSE` aligned to `8000`, `HEALTHCHECK` added, both stages pinned by patch tag and digest, and the `MOAJ_HTTP_ADDR` / `MOAJ_METRICS_TOKEN` requirements for container runs documented. |
| F6 | ADR-01 §3.1 points at the real flat monitoring filenames; the Compose header no longer references a non-existent `docker-compose.monitoring.yml`; `runbook_url` annotations are absolute links. |
| F7 | Compose publishes PostgreSQL and Valkey on `127.0.0.1` only, so the fixed development password is not reachable from the network. |
| F8 | `staticcheck` runs a second pass with `-tags ee` in CI and in the new `make staticcheck` target, so files behind `//go:build ee` are analysed. |
| F9 | Phase 1 record corrected (`./cmd/moaj`, `uid: moaj-phase1-runtime`); the pre-implementation snapshot moved into this changelog. |
| F10 | Partly addressed. Verified locally on Windows: `gofmt`, `go vet` (both tag sets), `go build` (both editions), `go test -count=1` (both tag sets), `staticcheck` (both tag sets), `govulncheck`, boundary-check, the license allowlist, YAML/JSON parse checks, and runbook-anchor resolution. Still unverified: `-race` (needs cgo/gcc, absent locally), `promtool check config/rules`, `go-licenses`, and a live Grafana import. See the outstanding-verification note in `IMPLEMENTATION_PLAN.md` §7 (Phase 1). |
| - | **Found while verifying F10 (not in the original review):** `make license-check` would have failed on the first CI run. It gated `go list -m all`, which reports the pruned module graph including ~23 test-only dependencies of dependencies (`stretchr/testify`, `alecthomas/kingpin`, `json-iterator/go`, ...) that are never linked into `bin/moaj`. It now gates `go list -deps` over both editions at `GOOS=linux GOARCH=amd64`, which resolves to exactly the ten modules in the allowlist. |
| F11 | Governance claims reworded to what is verifiable: Valkey/Asynq compatibility is stated as untested until Phase 5, CLA/PolyForm/ELv2 reuse terms are no longer asserted, and `deploy/` is MIT. |
