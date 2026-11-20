# Security Policy

**MOAJ (Mentor of all Jacks)** will hold personnel evidence about interns: task history, the reasons given for moves, review records and mentor notes. This policy says how to report a vulnerability, what the code enforces today, and what it is designed to enforce in later phases. It does not claim protection that has not been built and tested.

---

## 1. Supported versions

MOAJ has not had a release. Development follows [`docs/IMPLEMENTATION_PLAN.md`](docs/IMPLEMENTATION_PLAN.md).

| Version | Status | Security fixes |
| :--- | :--- | :--- |
| `main` (pre-release, Phases 0 to 2) | In development, not for production use | Fixed on `main` |
| Community `v1.0.x` (Phase 8 gate) | Planned first release | Security advisories and patch releases |
| Enterprise (Phase 9 gate) | Planned | Security advisories and patch releases |

MOAJ does not check for updates. Security releases will be announced through this repository's security advisories and release feed.

---

## 2. Reporting a vulnerability

**Do not report vulnerabilities in public issues, discussions or pull requests.**

Use **GitHub Private Vulnerability Reporting**: on `github.com/MOAJ-Mentor-of-all-Jacks/MOAJ`, open `Security` → `Report a vulnerability`.

Please include:

- the affected package or path (for example `internal/api` or `internal/config`) and the commit you tested;
- the steps to reproduce it, or a failing test;
- the impact: who could read or change what they should not, or what stops working.

What to expect:

- **Acknowledgement:** we aim to reply within 48 hours.
- **Assessment and plan:** we aim to reply within 5 business days.
- **Disclosure:** please keep the report confidential until a fix is released and operators have had reasonable time to upgrade. We credit reporters in the advisory unless they prefer not to be named.

Out of scope: vulnerabilities that need an attacker who already controls the server, the database owner role or the master key (see section 5), and findings in the source PDFs under `docs/specs/`.

---

## 3. What the code enforces today (Phase 1)

| Area | Behaviour |
| --- | --- |
| Configuration | Only `MOAJ_*` variables are read. Empty or malformed values stop start-up with an error instead of falling back to a default. Error messages never include the value of a setting, so a malformed database URL cannot leak its password into logs. |
| Secret files | Secret-capable settings accept a `*_FILE` variant for Docker and Kubernetes secret mounts (at most 64 KiB). Setting both the variable and its `_FILE` variant is rejected. |
| Metrics | `/metrics` requires a bearer token of at least 32 characters in production or whenever the server listens beyond loopback, compared in constant time. Metric labels are bounded: route patterns, methods, status codes. |
| HTTP | Security headers on every response: `Content-Security-Policy` (own origin only), `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: no-referrer`, a `Permissions-Policy` that denies camera, microphone and location, and `Strict-Transport-Security` when MOAJ itself terminates TLS (a reverse proxy that terminates TLS must set it). Request bodies are size-limited and every phase of a request has a timeout. Error responses use RFC 9457 problem details and `Cache-Control: no-store`. |
| Logs | Structured JSON. The configuration logs whether the database and metrics token are set, never their values. |

---

## 4. Designed, not yet built

These rules are part of the design and will be tested at the gate of the phase shown. Until then, do not rely on them.

| Rule | Phase |
| --- | --- |
| Four database roles (migration, application, retention, backup), none a superuser. The application role cannot update, delete or truncate evidence. `serve` refuses to start as a privileged role. | 2 |
| Argon2id passwords, server-side sessions, CSRF protection, rate limits. Session, invitation and reset tokens stored only as hashes. | 3 |
| Authorisation checked at the service and query layer, including exports and realtime streams. An intern sees only their own records; a system administrator sees no evidence. | 3 |
| Stored secrets and notebook notes encrypted with an operator-supplied master key ([ADR-08](docs/adr/ADR-08-secrets-and-keys.md)). | 3, 6 |
| Logs, job payloads and notifications carry record identifiers, never names, emails or typed text. | 2 onward |
| Finite retention that cannot be switched off, erasure under a separate database role, an erasure log, and encrypted backups. | 7 |
| No outbound connection unless an administrator configures one, so MOAJ runs on an air-gapped network. No telemetry. | Every phase; release gate in 8 |
| Enterprise licence verified offline, with no network call. | 9 |

---

## 5. Limits we state plainly

- **The database owner is trusted.** The evidence ledger is append-only for the application, but the first release has no tamper-evident chain. Someone with owner rights on the database can change records without detection.
- **Encryption covers stored secrets, notebook notes and backups, and nothing else.** A private notebook is hidden inside the application, not from someone who holds both the database and the master key.
- **Copies outside MOAJ are outside its control.** Exports someone has downloaded, and messages already delivered to email or Slack, are not erased by MOAJ's retention.
