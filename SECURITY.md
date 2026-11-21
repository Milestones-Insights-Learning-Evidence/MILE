# Security policy

MILE is a self-hosted tool for intern programmes. It holds personnel evidence about interns, so we treat security reports as a priority. This file covers how to report a vulnerability, what happens next, what the software protects today, and what it does not.

> **Project status.** There is no release yet. The code is at an early stage: much of the security design below is planned and not built. Section 4 says which parts exist today. Do not run a pre-release build with real intern data.

## 1. Reporting a vulnerability

Report privately. Do not open a public issue, discussion or pull request for a security problem.

- **Preferred:** GitHub private vulnerability reporting, from the repository's **Security** tab, using **Report a vulnerability**.
- **Fallback:** email **abdelrhmanalasser@gmail.com** with the subject line `SECURITY: <short summary>`.

There is no PGP key yet. If your report contains sensitive details, use GitHub private reporting rather than email.

Please include:

- the version or commit you tested
- how the software was deployed (binary, container, configuration that matters, with secrets removed)
- steps to reproduce, and what an attacker gains
- any proof-of-concept code, kept to the minimum needed

**Never send real personal data.** Do not attach a database dump, an export, real intern records or a support bundle from a production system. Use invented data. If you found a problem on a live installation, report the steps, not the data you saw.

Reports about the Enterprise edition go to the same address. Its code is private, so the report stays private between you and us.

## 2. What happens next

These are targets for a small team, not contractual promises. We will tell you if we expect to miss one.

| Step | Target |
| --- | --- |
| Acknowledge your report | 3 working days |
| First assessment: accepted, not a vulnerability, or more information needed | 10 working days |
| Fix released, critical severity | 30 days from acknowledgement |
| Fix released, high severity | 60 days |
| Fix released, medium or low severity | The next planned release |

Severity follows CVSS v3.1 or v4, adjusted for how the software is deployed: a flaw that needs an administrator account is usually less severe than one any signed-in intern can use.

We will keep you informed while we work on a fix, agree a disclosure date with you, and credit you in the advisory unless you ask us not to. The default is public disclosure when the fix is released, or 90 days after your report, whichever comes first. We may agree a different date with you if a fix needs longer.

### Good-faith research

We will not pursue or support action against anyone who:

- tests only installations they own or are authorised to test,
- avoids accessing, changing or keeping other people's data beyond the minimum needed to show the problem,
- does not degrade service for others, and
- reports privately and gives us reasonable time before disclosure.

This statement covers the project's own code. It cannot authorise testing of a customer's installation; that permission can only come from the customer.

## 3. Supported versions

There is no release yet, so no version receives security fixes.

From the first release, the latest release receives security fixes. Before version 1.0, fixes go into a new release only; older releases are not patched. A versioning policy stating which upgrade paths are supported and how long each version receives fixes will be published before version 1.0.

### How fixes reach you

The software contains no update check and makes no outbound connection to us (see section 5). Security releases are announced through:

- the repository's **Security advisories**, and
- the release feed (GitHub Releases).

Subscribe to one of them: we have no other way to reach you. Air-gapped installations need a process to bring these announcements and the release files across.

Each release is published with SHA-256 checksums, signatures and a dependency inventory. Verify the checksum and signature before installing. Release signing is not set up yet; the verification steps will be added here with the first release.

## 4. What the software protects today

Today this means the code in this repository. Everything in this section is covered by tests that run in CI.

- **Secrets from files.** Secret settings can be read from a file named by the matching `MILE_..._FILE` variable, so secrets need not sit in environment listings.
- **Start-up checks.** In production mode the server refuses to start with unsafe settings.
- **Metrics token.** The metrics endpoint compares its token in constant time.
- **No secrets in logs.** A test checks that configured secrets do not appear in log output.
- **Request size limit.** A request body larger than the configured limit is rejected with `413` and a problem-details response.
- **Bounded metric labels.** Metric labels hold only route patterns, HTTP methods and status codes. An unknown route is recorded as `unmatched` and an unknown method as `OTHER`, so a client cannot create new metric series or put a URL into a label.
- **No content sniffing.** Every response carries `X-Content-Type-Options: nosniff`.

## 5. Security design and its status

This section states the design so that reviewers and operators know what to expect. **Designed** means written down and accepted. **Built and tested** means code exists and has tests. Nothing below is in a release.

### Data sovereignty

| Control | Status |
| --- | --- |
| The customer runs the software and owns the database. | Design principle |
| No outbound connection unless an administrator configures one (mail, Slack, identity provider, S3 storage, a trace collector). The software runs correctly with all outbound traffic blocked, including on an air-gapped network. | Designed; a CI test with outbound traffic blocked is planned |
| No telemetry of any kind, no usage reporting, no update check, and no use of customer data to train any model. Metrics, logs and traces exist only for the customer's own monitoring. | Designed |
| The Enterprise licence is verified offline. | Designed |
| Support is by email. Nothing in the software sends, uploads or opens a support channel. The `support-bundle` command only writes a redacted local file that the customer may choose to attach. | Designed |
| The web interface loads nothing from third-party origins, enforced by a Content-Security-Policy. Links users add open without a referrer, and the server never fetches them. | Partly built: every response already carries a Content-Security-Policy limited to the server's own origin, `Referrer-Policy: no-referrer`, `X-Frame-Options: DENY` and a restrictive `Permissions-Policy`, but no test checks them yet. The web interface arrives in Phase 4. |

### Database

The schema below has been written and tested on PostgreSQL 16, outside this repository. It enters the repository in Phase 2.

| Control | Status |
| --- | --- |
| Four database roles: migration, application, retention and a read-only backup role. Nothing needs superuser rights. | Built and tested (outside the repository) |
| The movement history is append-only for the application role: update, delete and `TRUNCATE` are refused by permissions and triggers. | Built and tested (outside the repository) |
| Only the retention role can delete evidence. The database writes an erasure-log entry for every such delete and refuses a delete with no declared run. The log holds identifiers and times only and cannot be changed. | Built and tested (outside the repository) |
| Settings changes are logged by the database, never with a secret value. | Built and tested (outside the repository) |
| Background job payloads may hold identifiers and codes only; the database rejects names, email addresses and free text. | Built and tested (outside the repository) |
| The server refuses to start when connected as the migration role, a table owner or a superuser. | Designed |

### Accounts and access

| Control | Status |
| --- | --- |
| Passwords hashed with Argon2id; server-side sessions; secure cookies; CSRF protection; rate limits on sign-in. | Designed |
| Session, invitation and password-reset tokens are stored only as SHA-256 hashes, so a stolen database gives no working sign-in. | Schema built; application designed |
| A local administrator login always remains when single sign-on is configured, so an identity-provider outage does not lock the customer out. | Designed |
| Authorisation by role and scope at every service and query boundary, including exports and live updates. An intern sees only their own records; a system administrator sees no evidence. | Designed |
| IP addresses are kept only on sessions and sign-in audit entries. A failed sign-in for an unknown name stores nothing that was typed. | Designed |

### Encryption and keys

The software encrypts three things itself: secrets stored in the database, notebook notes and backups. Everything else relies on the operator's disk and database encryption.

| Control | Status |
| --- | --- |
| Stored secrets (mail, Slack, identity provider) encrypted with AES-256-GCM under keys derived from an operator-supplied master key. They are never returned by the API after saving and never appear in exports, logs or the support bundle. | Designed |
| Once an encrypted value exists, the server refuses to start without the correct master key. | Designed |
| Notebook notes encrypted at rest, with their key wrapped by the master key. | Designed |
| Backups encrypted. The current proposal is a public key, so a server can write backups it cannot read. | Designed; key type under review |
| A restore replays the erasure log, revokes every session and unused sign-in link, and cancels messages waiting in the backup, before access reopens. | Designed |

### Messages and logs

| Control | Status |
| --- | --- |
| Logs, job payloads and in-app notifications carry record identifiers, never names, emails or typed text. | Designed; enforced in the schema for job payloads |
| Slack and email messages carry the kind of item and a link back: no names other than the recipient's, no task text, no reasons. | Designed |
| Metric labels are bounded and never carry user IDs, task IDs, URLs, notes or emails. | Built and tested for HTTP metrics (see section 4); applies to every metric added later |

## 6. Known limits

These are deliberate, and we state them so that nobody relies on a promise the software does not make.

- **A database owner can change records undetected.** The append-only rules stop the application from rewriting history. They do not stop someone with owner or superuser access to the database. There is no tamper-evident chain in the first release.
- **Private notebooks are private in the application only.** Someone holding both the database and the master key can read them.
- **Removing a name is not anonymity.** A rare task, an exact time or a small cohort can identify a person without a name. The software does not claim to anonymise data.
- **Copies outside the software are beyond its reach.** Emails and Slack messages already sent, downloaded exports, and the operator's own copies of backups are not erased by the software.
- **Losing the master key loses encrypted data.** Stored secrets and notebook notes cannot be recovered without it.

## 7. The operator's responsibilities

The software is self-hosted. The customer who runs it is responsible for the deployment, including:

- TLS in front of the server, normally through a reverse proxy
- the database server: access control, network exposure, patching and disk encryption
- keeping the master key and the backup key apart from the backups, and from each other's storage where possible
- backups: where they are stored, more than one copy, and how long copies are kept
- protecting the metrics endpoint and bootstrap secret
- applying security releases promptly
- lawful use of the intern data the software holds

An operator hardening guide will ship with the first release.

## 8. Enterprise licence signing key

Enterprise licence files are signed with an Ed25519 key. The signing key is kept offline and never placed on a build server or in a repository.

Licences are verified offline, so a licence cannot be revoked remotely. If the signing key is ever exposed, we will:

1. publish an advisory,
2. release new Enterprise builds with a new verification key, and
3. issue new licence files to every customer.

Old builds will continue to accept licences signed with the exposed key until they are upgraded. Whatever happens to a licence, a customer's data, exports and administrator recovery are never locked.

## 9. Scope

In scope:

- the code in this repository and the official release files and container images built from it
- the Enterprise edition's code, reported privately as in section 1

Out of scope:

- vulnerabilities in a customer's own deployment that come from their configuration, such as an exposed database or a missing TLS proxy (report those to the customer)
- findings that need physical access to the server, or an already compromised administrator or database owner account
- reports from automated scanners with no demonstrated impact
- denial of service by volume against a single installation
- social engineering of maintainers or customers
- third-party services an administrator chooses to connect, such as Slack or an identity provider
