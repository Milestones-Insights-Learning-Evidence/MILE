# Open-Core Licensing Boundaries & Template Comparison (`LICENSING_BOUNDARIES.md`)

**Document Reference:** `IMS-GOV-01` §1–§4 & `docs/IMPLEMENTATION_PLAN.md` §3.1, §6.4, §7 (Phase 0 & Phase 9), Appendix A (`D17`)  
**Status:** Active Phase 0 Governance Record (Pre-Release Commercial Terms Gate Tracked Under `D17`)

---

## 1. Purpose

**MILE (Milestones, Insight, Learning & Evidence)** uses a **Single-Repository Open-Core** structure to support frictionless bottom-up adoption of the Community Edition while isolating commercial enterprise governance, identity, and compliance capabilities inside the top-level `ee/` directory.

This document defines:
1. The authoritative directory-to-license boundary mapping across the repository.
2. The Go build-tag isolation mechanism (`!ee` vs. `ee`) and automated CI enforcement ensuring the Community Edition binary (`bin/mile`) never links or imports proprietary `ee/` packages.
3. A comparison of reusable open-core legal templates and the contribution sign-off (DCO)—including their upstream reuse terms—and the explicit checklist required to close Decision `D17` prior to any Phase 9 Enterprise Edition distribution.

---

## 2. Authoritative Directory-to-License Mapping (`IMS-GOV-01` §1)

| Repository Path | Governing License | License File | Permitted Rights | Commercial & Technical Boundary |
| :--- | :--- | :--- | :--- | :--- |
| `cmd/`, `internal/`, `db/`, `ui/` | **MIT License** (Permissive Open-Source) | `LICENSE` | Free use, modification, self-hosting, private forking, and redistribution with copyright notice. | Core Kanban workflow state machine (`cmd/mile`), append-only `transition_ledger`, `HTTP 428` friction intercepts, SLA nudges, 30/60/90-day HR evaluation snapshots, and Prometheus/JSON observability (`mile_*`). |
| `deploy/`, `sqlc.yaml`, `Makefile` | **MIT License** | `LICENSE` | Free use, modification, and redistribution with copyright notice. | Dockerfile and Compose manifests, Prometheus scrape/alert configuration, the Grafana dashboard, SQLc configuration, and build tooling. These are machine-readable configuration, so MIT applies for the same reason it applies to `cmd/` and `internal/`; `CC-BY-4.0` is a content licence and is not used for them. |
| `docs/` | **MIT License** (`LICENSE`), with `CC-BY-4.0` dual-licensing of prose still open under `D17` | `LICENSE` | Free use, modification, and redistribution with copyright notice. | Architecture Decision Records (`docs/adr/`), governance and operations manuals. `IMS-GOV-01` §1 asks for `CC-BY-4.0 / MIT` here, but **no `CC-BY-4.0` notice file has been added to this repository**, so documentation currently ships under the root MIT licence only. Adding that notice is tracked in §5. |
| `ee/auth/`, `ee/rbac/` | **MILE Enterprise EULA** (Proprietary) | `ee/EULA.md` | **Source-available for customer security inspection** and audit only; production execution requires an active signed RS256 enterprise license (`/etc/mile/mile_enterprise.lic`). Zero modification or redistribution rights. | SAML 2.0, OIDC, SCIM 2.0 automated user/cohort provisioning, and fine-grained Policy-Based Access Control (PBAC). |
| `ee/audit/`, `ee/license/` | **MILE Enterprise EULA** (Proprietary) | `ee/EULA.md` | **Zero modification or redistribution rights; tamper-protected.** | SIEM audit log streaming (Datadog/Splunk) and the offline RSA-2048 (`RS256`) cryptographic JWT license verification engine. |

---

## 3. Build-Tag Isolation & Community Graph Verification

### 3.1 Build Tag Architecture (`internal/enterprise/`)
To guarantee clean compile-time separation without preprocessor hacks or runtime dynamic plugin loading:

- **`internal/enterprise/bridge.go`** (no build constraint): Defines the pure Go `Bridge` interface (`Edition()`, `LicenseFile()`, `ValidateStartup()`, and future hooks for SSO/SCIM/PBAC/SIEM) consumed by `cmd/mile` and `internal/api`.
- **`internal/enterprise/bridge_community.go`** (`//go:build !ee`): Compiled by default into all standard `go build` / `go test` invocations. Provides the MIT-licensed Community Edition implementation with zero dependencies on `ee/`.
- **`internal/enterprise/bridge_ee.go`** (`//go:build ee`): Compiled only when `-tags ee` is explicitly passed to the Go toolchain (`bin/mile-ee`). Wires the proprietary `ee/` packages (`ee/license`, `ee/auth`, `ee/rbac`, `ee/audit`) in Phase 9.
- **Files under `ee/**`**: Every `.go` file added under `ee/` must begin with `//go:build ee`.

### 3.2 Automated Boundary Verification (`make boundary-check`)
Every pull request and build runs a deterministic dependency graph inspection on the default Community build:

```sh
go list -deps ./cmd/mile
```

If any package in the default (`!ee`) transitive dependency graph matches `github.com/Milestones-Insights-Learning-Evidence/MILE/ee` (or `ee/...`), `make boundary-check` and `.github/workflows/ci-license-gate.yml` fail the build immediately.

---

## 4. Open-Core Legal Template Comparison (`D17`)

For Decision `D17` (`docs/IMPLEMENTATION_PLAN.md` Appendix A: MIT core, DCO, `CC-BY-4.0` docs, MIT deploy files, lawyer to confirm), the table below compares industry-standard open-core licensing, directory boundary, and contribution sign-off templates alongside their reuse terms:

| Template / Pattern | Primary Scope | Upstream License / Reuse Terms | Strengths for MILE | Adaptation Notes & Limitations |
| :--- | :--- | :--- | :--- | :--- |
| **Developer Certificate of Origin 1.1 (DCO)** | Contribution sign-off (`Signed-off-by:` trailer on every commit) | Published by the Linux Foundation at developercertificate.org. Its own header permits copying and distributing verbatim copies and forbids changing the text, so MILE links to it rather than adapting it. | Lowest friction for contributors; no signature database or bot account; checkable in CI from the commit history alone. Used by the Linux kernel, GitLab and Docker. | Grants only the project's own licence (MIT), not a right to relicense. That fits MILE because the Enterprise edition is built in a private repository that consumes this repository as an MIT dependency (§6.4 of the plan) and never relicenses contributed code. Code from outside contributors cannot later be moved into the Enterprise edition without each author's permission. |
| **Mattermost / GitLab Single-Repo Open-Core Boundary Model** (`LICENSE` + `ee/LICENSE`) | Single-repository directory split between MIT/Apache core and `/ee` commercial directory | Directory-isolation architecture is an industry convention; specific commercial subscription texts are proprietary to their vendors and must not be copied verbatim without permission. | Cleanly informs users in the root `LICENSE` that everything outside `ee/` is open-source while `ee/` is governed by `ee/EULA.md`. Matches `IMS-GOV-01` §1 directly. | MILE uses its own drafted `ee/EULA.md` provisions derived from `IMS-GOV-01` rather than copying third-party proprietary subscription agreements verbatim. |
| **PolyForm Shield / Fair Source / Elastic v2 (ELv2)** | Source-available non-compete licenses | Reuse terms for the licence **texts** are stated by their stewards (polyformproject.org and elastic.co) and have **not been verified for this repository**; no CC0 dedication or attribution-only permission is asserted here. Confirm the current terms at the source before copying either text. | Clear, concise legal drafting for source-available repositories. | **Not a full substitute for `ee/EULA.md`:** PolyForm/ELv2 focus on managed-service non-compete rather than per-seat (`max_intern_seats`) RS256 cryptographic key gating, 14-day grace periods, and `IMS-GOV-01` §4.2 anti-circumvention rules. Best used as structural inspiration only. |
| **Creative Commons Attribution 4.0 (`CC-BY-4.0`)** | Documentation & architecture specs (`docs/`, `deploy/`) | Standard public license by Creative Commons; freely usable by reference. | Named by `IMS-GOV-01` §1 (`CC-BY-4.0 / MIT`) for documentation. Appropriate for architectural prose. | **Not yet in effect.** `CC-BY-4.0` is a content licence, so it is applied to prose only - `deploy/` configuration stays MIT. No `CC-BY-4.0` notice file exists in the repository today, so `docs/` ships under the root MIT licence; adding the notice and the per-file attribution header is tracked in §5. |

### Selected Direction for MILE
1. **Contribution Sign-off:** Every commit carries a DCO `Signed-off-by:` trailer, checked in CI (see `CONTRIBUTING.md` §2). There is no Contributor License Agreement.
2. **Repository Boundary Notice:** Use the explicit root `LICENSE` (MIT) + `ee/EULA.md` directory split with automated `go list -deps ./cmd/mile` verification.
3. **Documentation & Deploy Assets:** Govern `deploy/`, `sqlc.yaml`, and `Makefile` under the root **MIT License** (they are configuration, not content). For `docs/`, `IMS-GOV-01` §1 specifies dual `CC-BY-4.0 / MIT`; until the `CC-BY-4.0` notice is added (§5), `docs/` ships under MIT only and must not be described as CC-BY-licensed.

---

## 5. Checklist to Close Decision `D17` Before Phase 9 Release

Phase 0 establishes the technical edition boundary, dependency allowlist, DCO sign-off requirement, and specification-aligned `ee/EULA.md`. Before any Phase 9 Enterprise Edition artifact is distributed or sold, the following items must be completed and verified:

- [ ] **Final Licensor Legal Entity:** Replace `MILE Contributors / Project Owner (see D17 in docs/IMPLEMENTATION_PLAN.md)` in `LICENSE` and `ee/EULA.md` with the registered corporate entity or copyright holder name.
- [ ] **Governing Law & Jurisdiction:** Specify the governing law, courts/venue, and dispute resolution clause in `ee/EULA.md`.
- [ ] **Template Reuse Clearance:** Have counsel confirm the reuse terms of any PolyForm/ELv2 text consulted for `ee/EULA.md`, and strip third-party trademarks.
- [ ] **Documentation Licence Notice:** If `docs/` is to be dual-licensed per `IMS-GOV-01` §1, add a `CC-BY-4.0` notice (for example `docs/LICENSE-CC-BY-4.0`) and state the attribution requirement. Until then `docs/` is MIT-only and no document may claim otherwise.
- [ ] **Commercial Master Subscription Schedule:** Attach commercial warranty, indemnification, SLA support tiers, and limitation-of-liability schedules reviewed by legal counsel to `ee/EULA.md`.
- [ ] **Production RS256 Key Isolation:** Confirm the vendor RSA-2048 private signing key resides in an isolated offline signing vault (`IMS-GOV-01` §3) and only the public key is embedded in `ee/license/`.
