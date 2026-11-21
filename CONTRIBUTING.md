# Contributing to MILE (Milestones, Insight, Learning & Evidence)

Thank you for your interest in contributing to **MILE (Milestones, Insight, Learning & Evidence)**. MILE is a self-hosted, open-core intern management system that converts everyday task workflow movements into transparent, reviewable HR and mentorship evidence.

Before submitting a pull request, please review the repository boundaries, the Developer Certificate of Origin (DCO) sign-off rule, build tags, and verification gates below.

---

## 1. Repository Structure & Open-Core Licensing Boundaries

MILE uses a **Single-Repository Open-Core** layout (`IMS-GOV-01` §1). Contributions must respect the boundary between the MIT-licensed Community Edition and the proprietary Enterprise Edition (`ee/`):

| Path | Edition & License | Description |
| :--- | :--- | :--- |
| `cmd/mile/` | Community (MIT) | Single-binary CLI entry point (`mile serve`, `mile version`, `mile help`; `migrate`, `backup`, `restore` and others in later phases). |
| `internal/` | Community (MIT) | Today: `config`, `api`, `observability`, `enterprise`. Planned (plan §6.1): `auth`, `workflow`, `reason`, `hr`, `db`, `jobs`, `notify`, `retention`, `backup`, `realtime`. |
| `db/` | Community (MIT) | Embedded Goose SQL migrations (`db/migrations/`) and SQLc query files (`db/query/`). |
| `ui/` | Community (MIT) | SvelteKit static frontend (`mile-ui`) embedded into the Go binary. |
| `docs/`, `deploy/` | Community (MIT) | Architecture Decision Records (`docs/adr/`), governance docs (`docs/governance/`), operations runbooks (`docs/operations/`), and container/monitoring manifests. `IMS-GOV-01` §1 additionally calls for `CC-BY-4.0` on documentation prose, but no `CC-BY-4.0` notice exists in the repository yet, so everything here is MIT today (tracked in `docs/governance/LICENSING_BOUNDARIES.md` §5). |
| `ee/` | **Enterprise (Proprietary EULA)** | Enterprise-only modules (`ee/auth/`, `ee/rbac/`, `ee/audit/`, `ee/license/`) governed exclusively by `ee/EULA.md`. |

See `docs/governance/LICENSING_BOUNDARIES.md` for full details on build-tag separation and licensing rules.

---

## 2. Developer Certificate of Origin (DCO)

Every commit must be signed off under the [Developer Certificate of Origin 1.1](https://developercertificate.org/). By signing off, you certify that you wrote the change, or otherwise have the right to submit it, under this repository's MIT License.

Sign off by committing with `-s`:

```sh
git commit -s -m "fix(p2): retry a transaction on serialization failure"
```

This adds a trailer with the name and email from your Git configuration:

```text
Signed-off-by: Your Name <you@example.com>
```

CI rejects a pull request that contains a commit without a sign-off. To fix one, amend the last commit with `git commit --amend -s --no-edit`, or sign off a whole branch with `git rebase --signoff main`, then force-push the branch.

There is no Contributor License Agreement. Contributions stay under the MIT License. The Enterprise edition is built in a separate private repository that uses this repository as an MIT-licensed dependency, so it never needs to relicense contributed code.

---

## 3. Development Prerequisites

- **Go Toolchain:** Go `1.26.x`, latest patch (`go.mod` declares `go 1.26.0` as a minimum, not a pin; CI uses `go-version: 1.26.x`). Module path `github.com/Milestones-Insights-Learning-Evidence/MILE` per `D1`.
- **Formatting & Static Analysis:** `gofmt` (standard library), `go vet`, `staticcheck`, and `govulncheck`. Run `make vet` and `make staticcheck`: both run twice, once with default tags and once with `-tags ee`, because files behind `//go:build ee` are invisible to a default-tag analysis.
- **Database (Phase 2+):** PostgreSQL 16. PostgreSQL is the only service MILE depends on: background jobs, scheduling and the notification outbox also run on it (Phase 5).
- **Running tests on Windows:** Smart App Control blocks the unsigned test binaries that `go test` builds. Run the suite in a Linux container instead, which also matches CI: `docker run --rm -v "${PWD}:/src" -v gomod:/go/pkg/mod -w /src golang:1.26 go test -race ./...`
- **Third-Party Dependency Policy (`NFR-03`):**
  - Distributed runtime dependencies must use approved permissive licenses (`MIT`, `BSD-2-Clause`, `BSD-3-Clause`, `Apache-2.0`, `ISC`, `PostgreSQL`).
  - Strongly reciprocal or source-available copyleft licenses (`GPL-2.0`, `GPL-3.0`, `AGPL-3.0`, `LGPL-*`, `SSPL-*`, `RSAL-*`, `EUPL-*`) are strictly prohibited in distributed artifacts.
  - If a PR modifies `go.mod` or `ui/package.json`, update `docs/governance/DEPENDENCY_LICENSES.md` and ensure `make license-check` passes.

---

## 4. Community vs. Enterprise (`-tags ee`) Build Rules

MILE compiles two distinct editions from the same repository using Go build constraints:

1. **Community Edition (Default — `//go:build !ee`):**
   - Built with standard commands (`go build -o bin/mile ./cmd/mile` or `make build`).
   - Uses `internal/enterprise/bridge_community.go` safe no-op/Community stubs.
   - **Strict Rule:** No package compiled in the default Community build may import any package under `github.com/Milestones-Insights-Learning-Evidence/MILE/ee/...`. This is enforced by `make boundary-check` and CI.
2. **Enterprise Edition (`-tags ee` — `//go:build ee`):**
   - Built with `go build -tags ee -o bin/mile-ee ./cmd/mile` or `make build-ee`.
   - Uses `internal/enterprise/bridge_ee.go` to wire enterprise modules once implemented in Phase 9.
   - Both editions must compile cleanly and pass all tests on every pull request.

---

## 5. Pull Request Verification Gates

Before opening or updating a pull request, run the full local verification suite via the `Makefile` (or equivalent Go commands):

```sh
# 1. Verify formatting (must produce zero unformatted files)
make fmt-check

# 2. Run go vet on both Community and Enterprise build tags
make vet

# 3. Verify Community build never imports ee/
make boundary-check

# 4. Verify dependency licenses against the permissive allowlist
make license-check

# 5. Run unit and race tests for both Community and Enterprise builds
make test
make test-ee
make test-race

# 6. Build both Community and Enterprise binaries (./bin/mile and ./bin/mile-ee)
make build
make build-ee
```

### Coding & Privacy Guidelines for Contributors
- **No Secrets or PII in Logs/Metrics:** Never log raw passwords, session tokens, database credentials, cryptographic keys, or free-text intern evaluation notes. Never use `user_id`, `task_id`, raw URLs, or email addresses as Prometheus metric labels (`mile_*`).
- **Specification Amendments:** Always check `docs/IMPLEMENTATION_PLAN.md` and `docs/adr/` before implementing domain rules. Where the implementation plan amends the original PDF specifications in `docs/specs/`, follow the recorded ADR and implementation plan contract.
