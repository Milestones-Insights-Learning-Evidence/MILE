# MILE - Production Implementation Plan

Version: 4.1
Updated: 2026-11-21
Status: The design for Phase 2 is ahead of the code. ADR-07 and ADR-08 are written, and the Phase 2 schema is drafted and tested on PostgreSQL 16. CI runs and passes. None of the Phase 2 work is in the repository yet: it enters after the rest of the Phase 0 gate. Phase 0 and Phase 1 remain reopened for re-verification.

> **Naming.** The product is MILE (Milestones, Insight, Learning & Evidence). It replaces MOAJ (DEC83). Phase 0 renames everything that still says MOAJ, in one change. The Phase 2 schema already uses the `mile` prefix. This plan writes subcommands without the binary name.

## 1. Purpose and authority

Build a self-hosted, open-core tool that turns everyday task movements on a board into understandable, reviewable evidence for mentors and HR. The Community edition must be deployable, recoverable and monitorable at its first release.

This document replaces versions 2.1 to 3.2. It is a reorganised rewrite: every rule appears once, under its topic, and the layered change tables of earlier versions are gone. Appendix B lists what changed since 3.2. Every work item and gate of version 3.2 is carried over, changed only where a later decision changed it.

Authority, in order:

1. This plan.
2. The decisions workbook, version 9. `DEC` and `O` identifiers refer to it.
3. The ADRs in `docs/adr/`, and `db/SCHEMA.md` for the schema.
4. The Data Handling and Integrity document, which states section 4 for customers.
5. The four specification PDFs in `docs/specs/`, as source requirements subject to the amendments recorded here.

Do not silently resolve conflicts by mixing incompatible requirements. Record the decision and update the relevant ADR before implementing the affected behaviour. Anything marked "to confirm" or carrying an open-item number is a recommendation, not a decision.

## 2. Where the project stands

| Item | State | Next step |
| --- | --- | --- |
| Phase 0, foundation | Reopened. CI runs and passes (DEC84). Fixes F1 to F12, the rename to MILE and the document updates are outstanding. | Do it now (O30). |
| Phase 1, HTTP runtime | Reopened for re-verification in CI. | After Phase 0. |
| Phase 2, persistence | Design done: data inventory, secrets inventory, ADR-07, ADR-08, seven migrations with 39 tables, 133 tests passing on PostgreSQL 16. Not in the repository. The threat model is not written. | Enters the repository after the Phase 0 gate. |
| Phases 3 to 10 | Not started. | |
| Validation track | Not started. | Start with Phase 2. |

| Document | State |
| --- | --- |
| ADR-01, ADR-02 | Need corrections (Phase 0). |
| ADR-03 | Needs a rewrite (Phase 0). |
| ADR-04, ADR-05, ADR-06 | To write (Phase 0). |
| ADR-07, boards, columns and moves | Accepted 15 November. Seven of its rules and one question await confirmation (O60, O63). |
| ADR-08, secrets and keys | Accepted 15 November. The backup key and the backup role are proposals (O61, O66). |
| Data Handling and Integrity | Draft 0.1, with the data inventory. Final version in Phase 7. |
| `db/SCHEMA.md` | Current with the migrations. |
| Threat model | Not written (O18). Due in Phase 2. |
| Permissions matrix | Phase 3. DEC68 is its starting point. |

Eight points await confirmation (O59 to O66). The design already assumes the recommended answer to each, and four of them would change the schema if answered differently. They are listed first in section 10.

## 3. Product rules

### 3.1 Editions and licence

| Topic | Rule |
| --- | --- |
| Editions | Community: MIT, free, unlimited interns, one department, free support. Enterprise: paid, private code (DEC1 to DEC5, DEC29, DEC36, DEC37). |
| Stays free | The board, tasks, tags, reason capture, confirmation, undo, history, local sign-in, basic OIDC, the progress view for one cohort, snapshot freezing and amendments, notebooks, retention and erasure, an intern's view of their own evidence, CSV and JSON export, monitoring, Slack and email notifications, scaling to several server processes, the S3 backup target (DEC20, DEC54). |
| Paid | Several departments and cross-department dashboards; comparing cohorts; SAML, SCIM, role and group mapping, several identity providers, SSO-only enforcement, policy-based access; the branded PDF review pack and mentor sign-off; audit retention beyond 12 months and audit streaming; branding; connectors (DEC20, DEC23, DEC54, DEC57). The final list and prices are O6. |
| Licence | An offline signed file holding customer, slots, edition and expiry; no network call. A 14-day grace period after expiry, then paid features become read-only. Data, exports and administrator recovery are never locked (DEC43). |
| Slots | Enterprise only. Reusable, measured by peak use in fixed bands (DEC19). A slot is used by each person with at least one active intern membership. One shared pool per licence, with an optional cap per department. A scheduled cohort that would exceed the limit gives a warning; the block applies to adding interns only, never to existing data (DEC43, DEC76, DEC77). |
| First build | The Kanban board and a department progress view. No connectors and no AI (DEC8, DEC9). |
| First users | Software companies in Egypt with 5 to 30 interns and engineers as mentors (DEC35). |

### 3.2 Boards, tasks and moves

ADR-07 holds the full rules.

| Topic | Rule |
| --- | --- |
| Boards | One board per cohort inside the department, on one version of a workflow definition. Mentors filter it by intern (DEC48). |
| Workflow versions | Every column edit creates a new, immutable version. A board moves to it through an explicit column map, and each open card gets a remap event that is never backward (DEC78). A column holding cards cannot be removed unless the map moves them (DEC79). |
| Tasks | One assignee per task, an intern, with reassignment history; evidence follows the assignee at the time of each event. Each membership names its mentors. A card can be moved by its assignee, that intern's mentors and administrators. Interns can create tasks for themselves (DEC50). |
| Task content | Plain-text comments and links; no file attachments in the first release (DEC56). An optional external reference and an optional time allowance (DEC9, O31). Typed text is limited to a closed list of fields (DEC66). |
| Skill tags | Each task needs at least one tag; "General" is seeded and pre-selected (D7). Tag changes are kept as history, because tags decide which skill a reason counts against. |
| Text edits | Who changed a task's text, when, and which fields is recorded. The old text is not kept. |
| Backward moves | Columns have an order, and one or more are terminal. A move to an earlier column, or out of a terminal column, is backward (DEC49). The move is held until the person who made it picks a reason. |
| Reordering | Reordering a card inside a column is not a transition: no event and no version change. A drop into the same column does nothing (DEC80). |
| Blocked cards | Blocking is a flag on the card with a start and an end, not a column, so it never looks like a backward move (DEC49). Setting it needs a free-text note. A block set by an intern counts as excluded time only after a mentor approves it, under the confirmation rules. One set by a mentor or administrator is approved when set (DEC69). |
| Scope changes | A mentor or administrator marks them on a task. No note (DEC38, DEC66). |
| Abandoned tasks | A mentor or administrator closes a task with a required note. It is never deleted and is left out of the completed count (DEC81). |
| To confirm | Seven further ADR-07 rules (O60), and whether a forward move can be undone (O63). |

### 3.3 Reasons, confirmation and undo

| Topic | Rule |
| --- | --- |
| Who gives the reason | The person making the move (DEC10). The task's own intern chooses from the intern list; mentors and administrators choose from the mentor list. The database decides which list applies from who made the move, so a code from the wrong list is rejected whatever the client sends. |
| Code sets | Stored per installation and versioned. Editing labels, adding or retiring codes, or changing a flag creates a new version. An event stores the code and the version (DEC12, DEC13). One code carries an optional intern label and an optional mentor label; a code without an intern label is mentor-only, and paired wordings are reported as one reason (DEC52). No `OTHER` code and no free text (DEC11, DEC41). |
| Flags | `counts`, `does_not_count`, `partly_counts`, plus `needs_confirmation`. `partly_counts` is a third group shown beside the other two, never a fraction or a weight (DEC51). |
| Confirmation | A reason marked `needs_confirmation` chosen by an intern is pending until a mentor confirms or rejects it. A rejection requires a reason from the mentor list. Pending moves are shown as awaiting confirmation and are not counted either way. Any mentor of that intern can confirm; a reminder goes out after 2 working days; a department administrator can confirm after 5 (DEC28, DEC53). A reason chosen by a mentor or administrator is final when chosen. An intern can never settle their own reason. The intern sees the state of each of their own reasons, including a rejection and the reason that replaced it. |
| Undo | The mover can undo within 1 minute (admin-adjustable). It is stored as a linked pair and never deletes history; HR sees the move as undone. It is refused with `409` if the task has changed since the move. It cancels any pending confirmation. Confirmation requests and notifications for a move are held until its undo window has closed (DEC14, DEC39). |
| Prompt triggers | Phase 4 builds the backward-move prompt only. The time-overrun and blocked-oscillation prompts wait for O31. |

Seeded reason codes (DEC41). Companies edit them.

| Intern sees | Mentor sees | Flag | Needs confirmation |
| --- | --- | --- | --- |
| I need to research this more | Needs more practice on this skill | counts | no |
| I misunderstood the task | Misread the requirements | counts | no |
| I missed a detail | Missed details | counts | no |
| The requirements changed | Requirements changed | does not count | yes |
| I'm waiting on someone else | Blocked by a third party | does not count | yes |
| By mistake | By mistake | does not count | yes |
| (mentor list only) | My brief was unclear | does not count | no |

### 3.4 Reviews, evidence display and notebooks

| Topic | Rule |
| --- | --- |
| Evidence display | Lead with accomplishments, support needs, exclusions and concise evidence. No leaderboard and no automatic score. The return-offer decision is written by a person. |
| Review checkpoints | Days are set per cohort; 30, 60 and 90 are the seeded default (DEC55). They count from the membership join date, falling back to cohort start (D12). A confirmation still pending at a cutoff is shown as pending, and settling it later creates an amendment (DEC53). |
| Frozen reviews | One per membership per checkpoint, holding that intern's evidence only. Cohort baselines are computed when shown. Community stores no long-term aggregate (DEC63). Snapshot freezing and amendments are Community; the paid review pack is the branded PDF and the sign-off (DEC54). |
| Notebooks | Mentors create notebooks, each with a list of interns and notes per intern. Visibility per notebook: private, shared with the intern, or shared with HR. Private by default. The administrator sets which levels are allowed; the mentor chooses among them (DEC15, DEC42). Private notes stay out of review material unless released. The administrator override is off by default and audit-logged. Notes are encrypted at rest and can be exported and erased per intern (DEC16). A departed mentor's private notebooks stay hidden and expire with each intern (DEC61). |
| Agendas | Shared 1-on-1 agendas are separate from notebooks. |

### 3.5 People, roles and access

| Topic | Rule |
| --- | --- |
| Roles | Intern, mentor, department administrator, HR administrator, system administrator. A person may hold more than one role. |
| Who sees what | An intern sees only their own records. A mentor sees the interns named on their memberships. Department and HR administrators see the department. A system administrator sees no evidence (DEC68). The permissions matrix in Phase 3 carries the detail. |
| Personal data | A display name and a work email, both required; a password hash or the identity provider's subject; roles and account state. No other personal field (DEC59). |
| Sign-in | Email and password with Argon2id and database sessions (D4). Basic OIDC in Community: one provider, accounts linked by verified email, and a local administrator login always remains (DEC23). |
| Settings | Admins can switch optional features on or off (DEC21). Retention rules, the settings change log and an intern's access to their own evidence cannot be switched off (DEC22). |
| Customer's duties | The customer is responsible for lawful use of intern data. The software supports that with export, erasure and retention (DEC7). |

### 3.6 Language, time and notifications

| Topic | Rule |
| --- | --- |
| Language | English interface first, with every string externalised. Text that users type may be Arabic or mixed and must display correctly in either direction. |
| Time | Each installation has one time zone and a configurable working week (Sunday to Thursday, for example). The database clock decides every event time, every duration, the undo window and every review cutoff. A client clock is never trusted. Timestamps are stored in UTC. Time-zone data is embedded in the binary. |
| Notifications | An in-app inbox by default. Slack and email only when an administrator configures them. External messages carry the kind of item and a link back: no names other than the recipient's, no task text, no reasons (DEC70). |

## 4. Data contract

The Data Handling and Integrity document states this section for customers. Its final version describes tested behaviour (Phase 7).

### 4.1 Sovereignty

- The customer runs the software and owns the database, in both editions (DEC6, DEC32).
- The software makes no outbound connection unless an administrator configures one. It must run correctly with all outbound traffic blocked, including on an air-gapped network (DEC74).
- No telemetry of any kind, no usage reporting, and no use of customer data to train any model. Metrics, logs and tracing exist for the customer to monitor the application's health: metrics are read by the customer's own monitoring, logs go to the server's output, and traces go only to a collector the customer runs, off until configured, with no personal content (DEC74).
- The Enterprise licence is verified offline (DEC43).
- Backups are written to storage the customer chooses and are never sent to the vendor (DEC33).
- Support is an email to the vendor. Nothing in the application sends, uploads or opens a support channel. The `support-bundle` command only writes a redacted local file, which the customer may attach; support never uses a database dump (DEC73).
- The web interface loads no fonts, scripts or other resources from third-party origins, and a Content-Security-Policy enforces it. Links that users add open without a referrer, and the server never fetches them.
- The software does not check for updates. Security releases are announced through the repository's advisories and release feed.

Destinations an administrator can configure. A new one needs its own setting and a row in the Data Handling document.

| Destination | Edition and phase | What is sent |
| --- | --- | --- |
| The customer's mail server | Community, Phase 5 | Notification and password-reset emails (DEC70). |
| Slack | Community, Phase 5 | The same notifications. A cloud service; left off on an air-gapped network. |
| Identity provider (OIDC) | Community, Phase 3 | Sign-in requests. The server keeps the provider's subject and the verified email, and no tokens (DEC59). |
| S3-compatible backup storage | Community, Phase 10 | Encrypted backup files, to the customer's own endpoint (DEC54). |
| The customer's trace collector | Community, Phase 10 | Timing traces with no personal content (DEC74). |
| SAML provider, audit streaming, connectors | Enterprise, Phases 9 and 10 | Sign-in requests; audit events; task data for the connected tool. |

### 4.2 What the application can promise

During a configured retention period, personnel evidence is identifiable and access-controlled. At expiry, a controlled process removes expired personal records and links. Long-term benchmarks, which are an Enterprise feature, use assessed aggregate data, not individual timelines with renamed users (DEC63).

Do not describe UUID replacement, hashing or salt destruction as proof of anonymity. A rare task, an exact timestamp, a tiny cohort or a recognisable note can identify someone without a name.

The ledger is append-only for ordinary application operations during retention. A restricted, audited retention process may erase expired records. The first release has no tamper-evident chain, and the documents say plainly that a database owner can alter records undetected (DEC72).

The software encrypts stored secrets, notebook notes and backups, and nothing else. A private notebook is hidden in the application, not from someone who holds the database and the master key (DEC75).

### 4.3 Retention and lifecycle

Retention cannot be switched off and is always finite (DEC22, DEC62). The clock starts when a membership ends: its leaving date, or the cohort's end date. A cohort needs an end date at most five years after its start, and a membership cannot end after its cohort (O64).

| Data | During retention | At expiry |
| --- | --- | --- |
| Working evidence: assignments, transitions, reasons, confirmations, undo links, block events and approvals, scope changes, closures, tag and edit history, task text, comments, notebooks, 1-on-1 agendas | Restricted personnel evidence. Task text and notes are treated as potentially identifying; minimise collection. | Erased 12 months after the membership ends, in dependency order. The customer sets 1 to 120 months (DEC62). |
| Review record: frozen reviews, amendments, assessments, self-appraisals, the recommendation | Restricted | Erased after its own period of 1 to 120 months; the default is open (O59). A frozen review keeps its copy of task text until then (DEC62). |
| Membership | The subject of all evidence | Deleted with the last record that points to it. |
| Intern account | May serve several active memberships | Erased when the last of that person's records expires. Never erased because one cohort expires (DEC62). |
| Mentor or administrator account | | Erased 12 months after deactivation. Records they wrote stay with the person they are about, with the author link emptied (DEC61, DEC62). |
| Aggregates | Community stores none. Cohort baselines are computed when shown, under small-group rules (D10, DEC63). | The Enterprise cohort comparison adds its own aggregate table: only approved aggregates, with small groups suppressed or combined. |
| Job payloads, outbox, notifications, idempotency keys | Identifiers and codes only (DEC64). Idempotency keys 24 hours. | Delivered messages and finished jobs 7 days, failed ones 30. In-app notifications 90 days, or sooner with their record. Retries must not recreate erased information. |
| Sessions, invitation and reset links | Stored as hashes (DEC65). The IP address on sessions and sign-in audit entries only (DEC71). | Sessions removed within 24 hours of expiry. Invitations 7 days, reset links 24 hours. |
| Logs | Record identifiers, never names, emails or typed text (DEC64). Written to the server's output. | Kept or dropped by the operator's log system. |
| Audit log | Restricted to administrators; identifiers and codes | Expires after 12 months in Community; longer and configurable in Enterprise (DEC57). |
| Settings change log, erasure log | Written by the database. No secret values; identifiers and times only. | Kept for the life of the installation (DEC60, DEC62). |
| Legal-hold records | | Deleted 12 months after the hold is lifted (DEC67). |
| Backups | Encrypted, customer-held, with a stated lifetime (30 days by default) | Pruned on schedule (DEC33). |
| External notifications and downloaded exports | Minimise content; document destinations | The software cannot erase independently held copies. |

### 4.4 Erasure, export and legal hold

- Every personal record has one subject, the person it is about. Erasing the subject deletes the record. Erasing the author keeps it and empties the author link (DEC61).
- A retention job runs under its own database role, the only role that may delete evidence. The database writes an erasure-log entry on every delete of evidence or of an account, and refuses a delete with no declared run. The log holds table names, record identifiers and times only, and can never be changed.
- The erasure audit proves operation, time, scope and outcome without keeping the content it erased.
- The same restricted process can export everything held about one person, and erase one person on request before expiry. An HR administrator starts either in the application; a system administrator can do the same from the command line. An export is streamed, not stored, and includes private notes about the person unless the requester leaves them out (DEC7, DEC64, DEC67).
- A legal hold suspends erasure for named people or cohorts. It outranks expiry and an erasure request, it does not stop backup pruning, and placing and lifting a hold are audited (DEC67).

### 4.5 Backups and restore

- Commands: `backup`, `restore` and `verify-backup`. A backup is an encrypted dump written to a folder the administrator chooses, with a stated lifetime and pruning. Backups can run on a schedule inside the server. An automatic backup is taken before every migration. S3-compatible storage comes in Phase 10 (DEC33, DEC54).
- The `backup` command connects with a read-only backup role (O66). The backup key follows ADR-08 (O61).
- A copy of the erasure log goes to the backup folder at every retention run and every backup (DEC60).
- A restore needs the backup key, and the restored database is fully readable only with the master key as well. Before service reopens, the restore replays the erasure log, revokes every session and unused sign-in link, cancels messages that were waiting in the dump, re-plans scheduled work and sets the database's connection rights again (DEC65).
- The replay covers every entry, not only entries newer than the dump. Identifiers are never reused, so a replay is always safe, and a time filter can miss an erasure whose transaction was still open when the dump was taken. This corrects DEC60 (O65).
- Operator guidance: not on the database server, key stored apart from the backups, more than one copy, and the lifetime applied to every copy.

### 4.6 Database roles

Use separate migration, application, retention and backup database roles. Triggers and permissions must stop the normal application role bypassing append-only rules, including by `TRUNCATE` and DDL. `serve` refuses to start when connected as the migration role, an owner or a superuser. Nothing needs superuser rights, so a managed PostgreSQL service works. `db/roles.sql` creates the roles. Supported PostgreSQL versions are stated and tested in CI (O62).

## 5. Workflow and evidence contract

### 5.1 Transitions and concurrency

- Every task mutation increments a task-wide `version`, except reordering inside a column (DEC80).
- A transition request carries the expected version and an idempotency key. Lock the task, check authorisation and version, evaluate the prompt rule, then commit the event, task pointer, version and any downstream intent in one transaction.
- A stale version returns `409` with the current authorised task. The UI puts the card in its real column at once and explains the conflict. Do not replay the rejected move automatically.
- A backward move without a valid reason returns `428`; nothing is committed. Release the transaction while the user chooses. Recheck authorisation, version and the rule on resubmission.
- The `428` response carries the reason list for the caller's role and the current code-set version. The `409` response carries the current task as the caller is allowed to see it.
- An idempotency key is scoped to one user and one task and kept for a stated period (24 hours by default). The same key with a different request body is rejected. The body is kept as a hash (DEC64).
- The database clock sets every event time and duration. A client clock is never trusted.
- Create an initial state-entry event when a task is created, so first-column durations have an anchor.
- Every version of a task has at least one event, and each thing that changed has the event that records it. The database refuses to commit otherwise.
- Validate department, cohort and tag consistency. Version every definition that affects interpretation, so concurrent column edits cannot change the meaning of a committed event.
- Retry a request on a deadlock or serialization failure. Both can occur when a board version switch meets a card write.
- ADR-07 records the board, column, move, block and closure rules. `db/SCHEMA.md` lists which rules the database enforces, which stay with the application, and the order in which a change is written.

### 5.2 Fairness and metric definitions

Keep raw movement counts separate from counted revisions. Show how exclusions affect the result.

Accepted rules (DEC38):

1. Record from the first schema: when a task became blocked and unblocked, and when its scope changed. The Phase 2 schema does.
2. A backward move whose reason is `does_not_count` is excluded once confirmed.
3. Blocked time is excluded from cycle time. A block set by an intern is excluded only once a mentor has approved it (DEC69).
4. Overlapping exclusion periods are merged, so no time is subtracted twice.
5. Undone moves are not counted and are shown separately.

Before implementing aggregation, write worked examples covering scope changes, third-party delays, overlapping exclusions, undo, pending and rejected confirmations, `partly_counts`, multi-skill tasks, final-column bounces, abandoned tasks, review-boundary crossings, reassignment during a task, non-working days, a daylight-saving change, and a cohort shorter than its checkpoint schedule. The decisions since version 3.2 add four: a rejected or unapproved block, a remap after a workflow change, a tag change during a task, and a task whose text was edited after work began.

Define cycle-time start and end, calendar versus working time, the completed-task denominator, skill attribution (from the tags in force at the time of each event), the cohort baseline and minimum evidence thresholds. Fairness rule versions and code-set versions are part of every snapshot's provenance. Choose small-group thresholds with the first users in mind: a cohort of 5 to 30 can fall below a threshold that suits a large company, and would then show nothing.

### 5.3 Make the review understandable

The primary view answers: what was accomplished, where support is needed, what was outside the intern's control, what evidence supports that, and what the mentor recommends.

- Show the review window, sample size, missing data and exclusions. Avoid confident statements from sparse evidence.
- A support suggestion (for example targeted pairing after three counted reasons on one skill) is a prompt for the mentor, not a judgement.
- Drill-down exposes tasks, reasons, dates, calculation explanations and definition versions, subject to role permissions.
- Keep system evidence, intern self-appraisal, mentor assessment and the human return-offer recommendation separate.
- Freeze published review evidence with a deterministic cutoff and a unique snapshot key. Corrections create an attributed amendment.
- Build explicit input fixtures and expected calculations before adopting golden files.

## 6. Architecture and operational contract

### 6.1 Boundaries

Use `api -> application services -> repository interfaces`, with SQLc/pgx adapters in the infrastructure layer. Pure workflow, reason and HR calculations do not import HTTP or pgx. Transaction orchestration belongs in application services.

| Location | Responsibility |
| --- | --- |
| `cmd/<binary>` | A thin `main` that calls the exported entry point. Subcommands: `serve` (HTTP plus background work), `migrate`, `backup`, `restore`, `verify-backup`, `rotate-master-key`, `support-bundle`, `version`. |
| `app/` (exported) | The application entry point. This repository's `main` and the private Enterprise `main` both call it. |
| `extension/` (exported) | Bridge interfaces with safe Community stubs, and the points where an Enterprise build registers routes, migrations, job handlers, settings, entitlements and UI assets. No Enterprise implementation lives in this repository. |
| `internal/config` | Validated environment settings and `_FILE` secrets. One prefix; no legacy aliases. |
| `internal/api`, `internal/auth` | HTTP, sessions, authorisation, CSRF, request limits, basic OIDC. |
| `internal/workflow`, `internal/reason`, `internal/hr` | Domain rules and evidence calculations. |
| `internal/db` | Generated SQLc access; never edited by hand. |
| `internal/jobs`, `internal/notify` | PostgreSQL-backed queue, scheduler, delivery, retries, outbox processing. |
| `internal/retention` | Retention planning, restricted erasure, erasure log, restore reconciliation. |
| `internal/backup` | Encrypted dump to a folder, restore, verification. |
| `internal/realtime` | Scoped SSE delivery and resynchronisation. |
| `internal/observability` | Logs, metrics, health. |
| `db/`, `ui/`, `deploy/`, `docs/` | Migrations and queries, embedded SvelteKit UI, operations assets, documentation. |

SvelteKit uses a static adapter and is embedded in the executable. No Node runtime in production.

The package names `app` and `extension` are placeholders for ADR-06 to fix. They are the contract between the two repositories: version them with the module and treat a breaking change as a coordinated release. The Enterprise interface is a separate bundle, built in the private repository from the Community UI plus the Enterprise screens, and the Enterprise binary embeds that bundle. No Enterprise screen ships here.

### 6.2 Durable delivery, background work and realtime

Commit the business change and the durable notification intent in one PostgreSQL transaction, and deliver asynchronously. Business transactions write outbox and job rows. In-process workers claim rows with `FOR UPDATE SKIP LOCKED` and a lease, record attempts, back off, and expose failures and delivery age. Do not claim exactly-once external delivery; document the possible duplicate after a successful send followed by a crash.

The scheduler takes a PostgreSQL advisory lock, so two accidentally running processes do not both run sweeps. Failed jobs have inspectable state and a controlled replay.

In-app notifications are rows in PostgreSQL and need no outbound connection. They are the default channel; Slack and email repeat them when configured.

Use LISTEN/NOTIFY as a wake-up, not a replay log. SSE reconnect resynchronises from persisted state. Apply authorisation to subscriptions and to every delivered resource.

Deployment is one binary plus PostgreSQL. Background work runs inside the server process (DEC25).

### 6.3 Monitoring as a Community feature

Protected Prometheus-compatible metrics, structured JSON logs, a versioned dashboard, scrape examples, alert rules and runbooks. Required visibility: HTTP, database, jobs and outbox, scheduled jobs, runtime, realtime. Use bounded labels; never user IDs, task IDs, raw URLs, notes or emails. Health checks separate liveness from readiness.

All of it is for the customer's own monitoring. Tracing, when it arrives in Phase 10, is off by default and goes only to a collector the customer runs (DEC74).

### 6.4 Edition boundary

- This repository is the whole Community edition and builds without any private code.
- Community manages exactly one department. The schema allows many, and nothing in Community offers a second one.
- The Enterprise edition is built in a private repository that imports this module's exported `app` and `extension` packages and registers its implementations there. It cannot import anything under `internal/`; the Go toolchain refuses.
- CI checks that this repository's dependency graph contains no Enterprise module.
- CI also builds a small test module from outside this module's tree that registers a stub through the bridge and starts the application. This proves another repository can use the bridge.
- Enterprise migrations keep their own history, separate from the Community migrations. Each Enterprise table follows section 6.6 and declares its inventory row through the bridge.
- Downgrade: the Community binary started on a database that holds Enterprise data must start, keep every existing record readable and exportable, and hide nothing silently. Its retention still covers the Enterprise tables. Tested in Phase 9.
- Editions of later items (DEC54): snapshot freezing and amendments are Community; comparing cohorts is Enterprise; Phase 10 scaling and the S3 backup target are Community; connectors are Enterprise (DEC20).

### 6.5 Secrets and keys

ADR-08 holds the inventory and the rules. Keep it current: every new secret is added there. In short:

- Database credentials for the four roles, the master key and the other process secrets come from the environment or a `_FILE`. They are never in the database.
- The master key, supplied by the operator, encrypts every secret stored in the database and wraps the notebook key. Once an encrypted value exists, `serve` refuses to start without it. Losing it makes encrypted settings and notebooks unreadable; the install guide says so plainly and tells the operator to keep it with the backup key, not with the backups.
- Mail, Slack and identity-provider secrets are never returned by the API after saving, and are left out of exports, logs, the settings change log and the support bundle.
- Session, invitation and reset tokens are stored as hashes. Rotating the session and CSRF secrets ends sessions safely.
- The backup key is kept apart from the backup folder. A restored database is fully readable only with this key and the master key, and the restore test proves both are needed and sufficient. ADR-08 proposes a public key, so that the server can write backups it cannot read (O61).
- The Enterprise licence signing key stays offline and never on a build server.

### 6.6 Database design rules

Six rules turn section 4 into table design, and every later migration follows them:

1. **One subject.** Every table with personal data names the person it is about: a membership for evidence, a user for account data (DEC61).
2. **Authors are links.** A column that says who did something can be emptied without deleting the row (DEC61).
3. **Random identifiers.** Every record the erasure log can name has a random UUID that is never reused (DEC60).
4. **Identifiers, not content.** Operational rows carry record identifiers and codes, never names or typed text (DEC64).
5. **Hashes, not tokens.** Anything that signs a person in is stored as a hash (DEC65).
6. **A retention class per table.** Working evidence, review record, account, operational or permanent (DEC62).

Every new table also gets its row in the `data_inventory` table, the no-truncate trigger, its grants, and the erasure-log trigger if it holds evidence or accounts. The schema tests fail if one is forgotten.

## 7. Development phases and exit gates

Phases are ordered by dependency, not by calendar. A gate must pass, with commands and results recorded in `docs/gates/` (one file per phase: commands, results, date, commit and accepted exceptions), before a phase is declared complete. Relative size is a rough guide only: S, M, L or XL.

### Phase 0 - Foundation realignment (reopened) - size M

Outcome: a clean repository, named MILE, whose documents match the decisions, with CI green.

Work:

- Keep CI green on Linux: build, format, vet, static analysis, vulnerability check, tests with the race detector. It runs and passes (DEC84); record the run in `docs/gates/`.
- Apply repository fixes F1 to F12 from the decisions workbook: the empty dashboard panel, legacy configuration aliases, the Go patch-level pin, the mislabelled CLA step, the container port and health check, document paths, database ports bound to all interfaces, static analysis of all files, plan record errors, overstated governance claims, and the security policy.
- Remove Valkey: settings, Compose service, documentation and dependency inventory entries.
- Replace `SECURITY.md` with the version written on 21 November.
- Rewrite ADR-03 for the new reason design. Add ADR-04 (reason vocabularies, versions, confirmation, undo, with the rules of DEC51 to DEC53 and block approval under DEC69), ADR-05 (single process with a PostgreSQL queue) and ADR-06 (edition boundary and private Enterprise repository: exported bridge and entry point, Enterprise UI bundle, separate migration history, downgrade behaviour including retention). Correct ADR-01 paths and ADR-02 (the erasure log written by the database, its copy in the backup folder, the full replay on restore, backup lifetime).
- Add ADR-07 and ADR-08 to `docs/adr/`. Both were written on 15 November. Copy this plan and the Data Handling draft into `docs/`.
- Replace the CLA wording with DCO sign-off in `CONTRIBUTING.md` and CI.
- Remove the `ee` build tag and folder from this repository. Move the bridge interfaces and stubs to an exported package and add the exported entry point (section 6.1). Add the out-of-module build test (section 6.4). Update `LICENSING_BOUNDARIES.md` and the boundary check.
- Rename from MOAJ to MILE in one change: module path, command folder, binary, environment prefix (`MILE_`), metric prefix, dashboard identifier, image and documents (DEC83). The Phase 2 schema already uses `mile` for roles, the database and the `mile.erasure_*` settings. Keep the product name in one constant afterwards.

Gate: CI is green with the commands and results recorded; the alert rules and scrape configuration pass their checkers and the dashboard loads; no reference to Valkey, Asynq, `OTHER`, CLA or legacy aliases remains in code, configuration, Compose files, CI or documents that describe the current state (ADRs, the changelog and this plan's appendix may name them as superseded); the out-of-module bridge test builds; the documents agree with decisions workbook version 9.

Depends on: nothing.

### Phase 1 - HTTP runtime and observability baseline (re-verify) - size S

Outcome: an inspectable service process with safe lifecycle behaviour, proven in CI.

Work:

- Re-run the Phase 1 tests in CI, including the race detector.
- Fix the container: matching port, documented metrics token for non-loopback binds, health check, pinned base image.
- Add a container smoke test: start, `/healthz`, `/readyz`, protected `/metrics`, graceful stop.
- Add a CI job that runs the tests with outbound traffic blocked, and keep it passing from here on. Phase 7 extends it to the full suite.

Gate: health and metrics smoke checks and handler tests pass in CI; shutdown drains requests; start-up failures are clear; secrets are absent from logs; the blocked-network job passes.

Depends on: Phase 0.

### Phase 2 - Persistence, schema and transactions - size L

Outcome: trustworthy data foundations that already record what later phases need.

Done in design and tested on PostgreSQL 16, outside the repository:

- The data inventory: every table, including those planned for later phases, who may see each row, who owns its retention, and the erasure order. It lives in the Data Handling document and in the `data_inventory` table.
- The secrets inventory (ADR-08), a first draft of the Data Handling and Integrity document and the list of outbound destinations (section 4.1).
- The model, in seven migrations with 39 tables: installation metadata with the time zone and working week; users; the organisation; departments; cohorts with their checkpoint schedule; memberships and the mentors of each with history; workflow definitions with immutable versions, column order, terminal columns, dwell limits and column maps; boards, one per cohort; tasks with `version`, an external reference and a time allowance; skill tags with a seeded "General" tag and the tag history of each task; text-edit records; assignment history; reason code sets, codes with two labels and flags, with the seeded defaults; transitions with `time_in_prev_state_ms` and an initial state-entry event; reason records with confirmation state; undo links; block events with approvals, read as intervals; scope-change markers; closures; outbox and job tables; idempotency keys; settings with limits and a change log; the audited events as a fixed list (sign-ins, account and role changes, membership changes, settings, code-set and workflow publication, exports, the notebook override, erasure, legal holds, backups and restores) and the audit log, kept 12 months in Community (DEC57); the erasure log.
- Foreign keys, checks and indexes; department consistency; column reordering made safe by immutable versions.
- Four database roles created without superuser rights (`db/roles.sql`), and append-only protection, including against `TRUNCATE` and DDL by the application role.
- 133 tests under the real roles: each rule in `db/SCHEMA.md`, two writers at once, a dump taken by the backup role and restored, and the erasure of one person.

Remaining work:

- Write the threat model from the inventory (DEC45, O18).
- After the Phase 0 gate, bring the migrations into the repository. Implement embedded migrations with an advisory lock, SQLc/pgx adapters, pool limits and transaction interfaces.
- Implement the write order of `db/SCHEMA.md` in the transaction layer, with a retry on deadlock and serialization failure.
- `serve` refuses to start when connected as the migration role, an owner or a superuser.
- State the supported PostgreSQL versions (O62) and run the integration tests against each. Use nothing that needs superuser rights.
- Move the schema tests into real PostgreSQL integration tests in Go; add database metrics; provide a development database bound to localhost.
- Tables for comments (DEC56), notifications and notebooks arrive with their features (D6) and are already listed in the data inventory. Nothing stores file attachments (DEC56).

Gate: migrations apply on a clean database; development rollback and reapply works; append-only and scoping tests pass under the real runtime role; editing a reason code creates a new version and leaves recorded events unchanged; block and scope-change times can be stored and read back; the schema supports the planned erasure order; every table is in the data inventory; a dump taken by the backup role restores cleanly; the data inventory, threat model and secrets inventory are written; the tests pass on every supported PostgreSQL version; `serve` refuses a privileged database role.

Depends on: Phase 0 (ADR-04, ADR-05), Phase 1, ADR-07 and ADR-08. The schema is final once O60, O63, O64 and O66 are answered.

### Phase 3 - Authentication and first-run setup - size L

Outcome: a protected installation that can create its organisation, its one department and its first users.

Work:

- One-time setup with an installation-local bootstrap secret and atomic first-admin creation.
- Argon2id passwords, server-side sessions, rotation and revocation, secure cookies, CSRF, rate limits (in memory), logout, and a documented recovery procedure.
- Tables for identity-provider links, role assignments, sessions, invitation links and reset links, each following section 6.6. Tokens are stored as hashes (DEC65). The IP address is kept on sessions and sign-in audit entries only; there is no "last active" indicator; a failed sign-in for an unknown name stores nothing that was typed (DEC71).
- Roles: intern, mentor, department administrator, HR administrator, system administrator, with assignment rules. A person may hold more than one role.
- A permissions matrix in `docs/` (role by action by resource, including who may create, assign and move tasks; DEC50, DEC68). The authorisation tests are derived from it.
- Account lifecycle that works with no mail server: invitation by a one-time link the administrator copies; password reset by an administrator-issued link, and by email only when mail is configured; deactivation that revokes sessions at once; role changes; an intern leaving early.
- Enforce authorisation at service and query boundaries, including exports and realtime subscriptions.
- Basic OIDC: one provider configured by issuer URL, client ID and secret; accounts linked by verified email; a local administrator login always remains. The server keeps the provider's subject and no tokens (DEC59).
- The master-key start-up check, from the first stored secret (ADR-08).
- Settings framework: optional feature switches, with the protected settings that cannot be switched off. Setup asks for the installation time zone and working week.
- Security headers, including a Content-Security-Policy that allows only the installation's own origin and a no-referrer policy. The server never fetches a link.
- Externalise every user-facing string from the first screen. Display text that users type with automatic direction, so Arabic and mixed text read correctly.
- Minimal setup and login UI; the API contract that generates the TypeScript client.

Gate: role and resource authorisation tests and setup-race tests pass; one intern cannot read another's data; first-run and recovery work without exposing credentials; OIDC sign-in works against a test provider and the local administrator can still sign in when the provider is down. With no mail server configured, an invited user can join and a deactivated user loses access at once. Tokens exist in the database only as hashes. The interface makes no request to a third-party origin.

Depends on: Phase 2.

### Phase 4 - First usable workflow slice: the board - size XL

Outcome: setup, task, board movement, reason capture and visible evidence work end to end for one department.

Work:

- Build in three slices, each one demonstrable: (a) board, tasks, tags, moves with versioning and the reason prompt; (b) confirmation, block approval, undo and the intern's own view; (c) code-set and workflow administration, the progress view and CSV import.
- Column, task and tag management; completion semantics; task versioning; transactional moves. One board per cohort (DEC48). Task assignment and reassignment (DEC50); plain-text task comments and links, with no file upload (DEC56); board filtering by intern, tag and text. Tag changes and text edits are recorded as history.
- Column management publishes a new workflow version: the administrator maps the old columns, and the board and its open cards move in one transaction (DEC78, DEC79).
- Blocking a card with a required note and clearing the block, recorded as an interval and not as a move (DEC49, DEC69). An action to mark a scope change on a task (DEC38). An action to close a task as abandoned, with a note (DEC81). The permissions matrix says who may use each.
- The backward-move prompt: `428` until the mover picks a reason from their own list.
- Confirmation: pending state for reasons that need it, a mentor queue, confirm and reject, with rejection requiring a mentor-list reason. Any mentor of the intern can confirm, and a department administrator can after 5 working days (DEC53). The same queue carries blocks set by interns (DEC69).
- Undo inside the window, only by the mover, stored as a linked pair, with the rules in section 3.3.
- The intern's own view: their moves, the reasons given, the outcome of each confirmation and of each block approval (DEC22). A count of pending items for mentors.
- Administrator screen for reason code sets: edit wording, add or retire codes, set flags; each save is a new version.
- Board UI: keyboard-accessible, with the reason dialog, immediate stale-card reconciliation, and scoped SSE with resynchronisation.
- Department progress view: per intern, tasks completed, backward moves by reason in three groups (counted, partly counted and not counted; DEC51), pending confirmations, undone moves, and abandoned tasks shown apart. It shows one cohort at a time (DEC54). No ranking and no score.
- Tables for comments and in-app notifications, holding identifiers only (DEC64).
- Generate the API client from the versioned contract. Instrument transitions, conflicts, prompts and realtime without individual identifiers in labels.
- Use the paper-prototype findings (O2) for the dialog wording before the UI is final.
- CSV import for tasks, keeping the external reference so that a repeated import updates instead of duplicating (DEC9). An import updates text fields only; it never moves a card and never back-dates an event.
- The board works at phone width. State the supported browsers.

Gate: tests cover the wrong-list rejection, every confirmation path, the undo window boundary and code-set versioning; two competing writes against one version give one success and one conflict; A to B to A also invalidates stale requests; retries do not duplicate events; a browser end-to-end path passes; an undo after a concurrent change is refused; an intern can see a rejection and the reason that replaced it; an intern's block is excluded only after approval; a board moved to a new workflow version keeps every card and its history; a repeated CSV import creates no duplicates and moves no card; a move out of a terminal column asks for a reason, and blocking a card does not. This is an internal demonstration, not a release.

Depends on: Phase 3. The final dialog wording depends on O2.

### Phase 5 - Background work and notifications - size L

Outcome: asynchronous work survives retries, restarts and delivery failures, inside one process.

Work:

- In-process worker pool on the PostgreSQL job and outbox tables: claim with a lease, retry with backoff, failed-job inspection and replay.
- Scheduler with advisory-lock leadership.
- Before building the sweep, record what an SLA nudge measures (a dwell limit per column, a time allowance per task, or both) and who receives it, from BR-WF-04.
- SLA sweeps every 15 minutes by default; neutral Slack and email notifications, sent only when an administrator has configured them, within the content limits of DEC70. Choose the Slack delivery method here.
- An in-app inbox as the default channel for pending confirmations and block approvals, their reminders after 2 working days (DEC53), rejections, SLA nudges and checkpoint dates. Reminders respect the working calendar.
- The time-overrun and blocked-oscillation prompts, if O31 is decided in favour.
- Development Compose with the server and PostgreSQL only, with health checks and volumes.
- Job health and metrics; queue and outbox dashboard panels and alert rules.

Gate: tests cover repeated sweeps, two processes running at once, a crash between commit and delivery, restarts, send failures, and a send that succeeds but is not acknowledged. No committed intent is silently dropped; duplicate-delivery limits are documented; nothing depends on a key-value store. Every notification can be read in the application with outbound traffic blocked. External messages carry no evidence content.

Depends on: Phase 4. O31 for the two extra prompts.

### Phase 6 - HR translation and review experience - size XL

Outcome: mentors and HR can understand, verify and act on fair evidence.

Work:

- Cohort lifecycle, milestone scheduling with checkpoint days set per cohort (DEC55; Day 30, 60 and 90 are the seeded default), versioned rubrics, 1 to 5 scales, weights, self-appraisal and schema validation.
- Worked fairness examples and calculation fixtures (section 5.2), then aggregation, exclusions, sample-size handling and support suggestions.
- Freeze evidence at each checkpoint at defined cutoffs with unique snapshot keys: one frozen review per membership per checkpoint (DEC63). A confirmation still pending at the cutoff is shown as pending; settling it later creates an amendment (DEC53). Freezing and amendments are Community (DEC54).
- Review cards, evidence drill-down, mentor responses, published amendments, and an explicit human return-offer recommendation.
- An intern's view of their own evidence.
- Shared 1-on-1 agendas.
- Notebooks: creation by mentors, a list of interns and notes per intern, three visibility levels with the allowed levels set by the administrator, private by default, private notes kept out of review material unless released, administrator override off by default and audit-logged, notes encrypted at rest with the key handling in ADR-08, the rotation procedure and `rotate-master-key`, export and erasure per intern.
- CSV and JSON exports, protected against formula injection. CSV is UTF-8 with a byte-order mark so that Arabic text opens correctly in Excel. Exports are streamed, not stored (DEC64).

Gate: fixture-based calculations and snapshot replay are deterministic; editing definitions does not rewrite published reviews; exclusions are explained; role tests cover notebooks and exports, including that a private notebook never appears in search, export or review output; a mentor can trace every summary claim to evidence.

Depends on: Phase 5.

### Phase 7 - Retention, backup and recovery - size L

Outcome: retention, backup and restore are implemented and tested, and the data-handling promises are proven.

Work:

- Finalise retention anchors, including early departures and multiple memberships, and the review-record default (O59).
- Configurable retention with its two periods, aggregate privacy rules, restricted erasure, non-identifying audit records, retries and checkpoints.
- Export of everything held about one person, and erasure of one person on request before expiry, through the same restricted process (DEC7), started by an HR administrator in the application or by a system administrator from the command line (DEC67).
- Legal hold: suspend erasure for named people or cohorts, with an audit entry when placed and when lifted. Enforce it in the database, at the same delete triggers that write the erasure log.
- Apply retention across tasks, events, reasons, notebooks, reviews, memberships, accounts, jobs, outbox and logs according to the data inventory, and expire the audit log after 12 months (DEC57).
- Coordinate erasure with background work so nothing recreates expired evidence.
- `backup`, `restore` and `verify-backup`: an encrypted dump written to a folder the administrator chooses, with a stated lifetime, taken through the backup role (O66). Key handling follows ADR-08 (O61). Backups can run on a schedule inside the server, and backups past their lifetime are pruned. The operator guide says: not on the database server, key stored apart from the backups, more than one copy (DEC33).
- The copy of the erasure log in the backup folder at every retention run and every backup (DEC60).
- Restore as in section 4.5, all before normal access resumes: the full erasure-log replay (O65), revoked sessions and links, cancelled waiting messages, re-planned scheduled work, and connection rights set again (DEC65).
- An egress test: the full test suite passes with all outbound traffic blocked, extending the Phase 1 job, and the interface makes no third-party request.
- Write the Data Handling and Integrity document from the tested behaviour.
- Retention and backup monitoring and procedures.

Gate: erased evidence is absent from active stores and authorised joins; active memberships stay intact; small-group protections work; interrupted erasure resumes safely; a restored backup is reconciled before access; restoring an older backup brings back nothing erased since, and no session from the backup works; the egress test passes; one person's data can be exported and erased on request; a record under legal hold survives a retention run; expired backups are pruned; a restore needs both keys and is complete with them; residual identification risks are documented.

Depends on: Phase 6.

### Phase 8 - Production-ready Community release - size XL

Outcome: a new organisation can install, operate, monitor, back up and upgrade Community using the shipped documentation.

Work:

- Complete UI accessibility, loading, empty and error states, the column and rubric editors, review workflows and critical-path end-to-end coverage.
- Embed UI assets and migrations; confirm no Node runtime is needed.
- Ship a non-root image and a Compose file with the server and PostgreSQL, health checks, volumes and examples. Document TLS, reverse proxy and secrets.
- Take an automatic backup before every migration.
- Sample data and a five-minute first install; a public demonstration with invented interns.
- Finish dashboard provisioning, scrape examples, alert rules and runbooks; test with a real Prometheus and Grafana.
- Test backup and restore, migrations, recovery guidance and upgrades.
- Write the size, memory, install-time and capacity targets into this plan, taken from the specifications. Define and run a capacity test; report results and bottlenecks. Resolve any target that is infeasible.
- Updated threat model and the operator hardening guide; a documentation site built from `docs/`.
- `support-bundle`: a redacted diagnostic file (version, configuration without secrets, migration state, job and outbox statistics), with a test that it holds no personal content. It only writes a local file. Support itself is by email (DEC73).
- Project files for a public repository: code of conduct, issue and pull-request templates, changelog, public roadmap, and a versioning policy that says which upgrade paths are supported and which versions receive security fixes.
- Security releases are announced through the repository's advisories and release feed. The software contains no update check.
- Publish a tagged release with checksums, a dependency inventory and signatures.

Gate: a clean machine reaches a working board using only the install guide; critical end-to-end, authorisation, retention, restore, monitoring and load gates pass; no open critical defect.

Public launch also requires, outside engineering: the employer release (O1), the licence confirmed by a lawyer (DEC36), and the recorded result of the adoption test (O2).

Before the repository is made public (DEC34): remove business, pricing, employer and legal notes from every document, including this plan's non-engineering launch conditions, which move to the private workbook. Review the commit history for the same material and for earlier project names, and start from a clean history if any is found. Until the selling structure exists (DEC30, O28), the repository and the documentation site carry no donation or sponsorship links and no prices.

Depends on: Phase 7.

### Phase 9 - Enterprise edition - size XL

Outcome: paid controls add multi-department governance while preserving recovery and access to existing records.

Work:

- Private repository and build pipeline producing the Enterprise binary from this module's exported packages plus the private code, with its own UI bundle and its own migration history (sections 6.1 and 6.4). A procedure for customers to inspect the private code under NDA (DEC37).
- Offline licence: a signed file holding customer, slots, edition and expiry; verification with an embedded public key and an algorithm allowlist; installation binding; a 14-day grace period after expiry, then paid features become read-only. Data, exports and administrator recovery are never locked. The signing key stays offline. A licence file is installed and renewed without a network. The installation identifier lives in the database, so it survives a restore and a move to a new host. A system clock set backwards is detected and reported.
- Slot counting: reusable slots measured by peak use in fixed bands. A slot is used by each person with at least one active intern membership and is freed when the membership ends. One shared pool per licence, with an optional cap per department. Over the limit: a warning when a cohort is scheduled, then a block on adding interns only, never on existing data (DEC19, DEC43, DEC76, DEC77; bands in O6).
- Several departments, cross-department dashboards and comparison across cohorts (DEC54), with its own aggregate table and small-group rules (D10, DEC63).
- SAML, SCIM with cohort mapping, role and group mapping, several identity providers, SSO-only enforcement, policy-based access.
- Formal review packs: branded PDF and mentor sign-off, built on the Community snapshots and amendments (DEC54). Choose a PDF method that shapes Arabic and right-to-left text correctly, and test it with mixed content before committing to a library.
- Audit retention beyond the Community 12 months, configurable (DEC57); audit streaming; branding.
- Final paid feature list and pricing (O6); EULA with the customer-responsibility clause (O26); the selling structure in place before any sale (O28).

Gate: this repository's build contains no Enterprise code; both editions pass their tests; entitlements and department scoping are enforced; slot counting is tested with fixed dates, including one person in two cohorts; licence expiry and identity-provider outages preserve tested recovery and export paths; the Community binary starts on an Enterprise database with every record readable and exportable, and its retention covers the Enterprise tables; commercial terms are complete.

Depends on: Phase 8.

### Phase 10 - Release automation and measured scaling - size L

Outcome: repeatable releases and capacity improvements based on evidence.

Work:

- Multi-architecture artifacts and automated release, inventory, signing and provenance workflows.
- Several server processes sharing the PostgreSQL queue; scheduler failover; SSE routing and resynchronisation; pool sizing and overload behaviour. Community (DEC54).
- S3-compatible backup target, in Community (DEC54).
- Tracing where needed, for the customer's own monitoring: off by default, sent only to a collector the customer runs, with no personal content (DEC74).
- Capacity baselines and performance regression tests.
- Connectors (Enterprise, DEC20), starting with the one the first customers ask for.

Gate: reproducible release procedures and documented capacity and failover results on declared infrastructure.

Depends on: Phase 8; Enterprise items on Phase 9.

### Validation track (runs alongside Phases 2 to 4)

- 10 discovery interviews with mentors and intern-programme owners: where intern work lives today, how interns are judged, whether AI has changed what interns do. No demonstration and no price.
- A paper prototype of the board and the reason dialog with 5 mentor and intern pairs.
- Go rule: 3 teams say they would install it. If the rule fails, stop before Phase 5 (DEC44), record what was learned, and decide whether to change the product or the first users before building further.
- Record the findings in the decisions workbook and adjust Phase 4 wording and defaults.

## 8. Cross-cutting completion rules

- Tests progress from pure domain tests to PostgreSQL integration and browser end-to-end. Run relevant gates as features land.
- All requests and jobs have bounded work, deadlines, validated input and scoped authorisation. Use parameterised SQL and redact secrets.
- Keep the API contract, migrations, metric definitions, dashboard panels and documentation in step with behaviour.
- Use integration tests for privilege boundaries, concurrent writes, crash recovery, erasure and migrations.
- Keep the edition boundary and check shipped dependency licences for each release, including the UI bundle.
- New outbound connections need an explicit setting and an entry in the Data Handling and Integrity document.
- Every new secret is added to the inventory in ADR-08.
- Every new table follows section 6.6 and the checklist in `db/SCHEMA.md`.
- Logs, job payloads and notifications carry record identifiers, never names, emails or typed text (DEC64).
- Every new user-facing string is externalised, and every screen is checked with right-to-left text.
- Do not mark a phase complete because files exist; record the commands, results and accepted exceptions for its gate in `docs/gates/`.

## 9. Requirement-to-phase map

| Requirement | Phase(s) |
| --- | --- |
| BR-WF-01 custom states; BR-WF-02 skill tags | 2, 4 |
| BR-WF-03 ledger, amended for retention | 2, 4, 7 |
| BR-WF-04 SLA nudges | 5 |
| BR-INT-01/02/03 triggers, dialog, taxonomy (amended: mover gives the reason, two lists, configurable codes, confirmation, undo) | 4; remaining triggers in 5; interpretation in 6 |
| BR-HR-01 cohort lifecycle | 2, 6 |
| BR-HR-02 rubrics and self-appraisal; BR-HR-03 evidence | 6 |
| BR-HR-04 1-on-1 and return-offer workflow (amended: notebooks) | 6 |
| BR-EE-01/02/03/04 licensing, identity, policy-based access, audit streaming (amended: basic OIDC in Community) | 3, 9 |
| NFR-01 embedded single binary | 1, 4, 8 |
| NFR-02 relational plus JSONB | 2, 6 |
| NFR-03 dependency licensing (met by removing the Redis-family dependency) | 0 and every release |
| NFR-04 retention, revised privacy contract | 2, 7, 8 |
| NFR-05 licence boundaries (amended: private repository, DCO) | 0, 9 |
| Governance role isolation | 3 and every new endpoint, job, export and stream |
| Governance grace, recovery and export guarantee | Community exports in 6; Enterprise behaviour in 9 |
| New: Community monitoring and operability | 1 onward; release gate in 8 |
| New: data sovereignty and backups | 1, 2, 7, 8 |
| New: one-department Community boundary | 2, 4, 9 |
| New: exported Enterprise bridge and out-of-module build test | 0, 9 |
| New: account lifecycle and permissions matrix | 3 |
| New: in-app notifications | 4, 5 |
| New: time zone, working calendar and right-to-left text | 2, 3, 6, 9 |
| New: secrets and key management | 2, 6, 7 |
| New: person-level export, erasure on request and legal hold | 7 |
| New: workflow versions, block approval, abandoned tasks, tag and edit history (ADR-07) | 2, 4 |
| New: database-written settings change log and erasure log; backup role | 2, 7 |
| New: support by email and air-gapped operation | 1, 7, 8 |
| New: slot counting per person, shared pool | 9 |

## 10. Open items

To confirm. Each has a recommended answer that the design already assumes. A different answer to O60, O63, O64 or O66 changes the schema; O59 changes one seeded setting; O61, O62 and O65 change ADR-08, the CI matrix and the restore procedure.

| ID | Question | Recommended answer |
| --- | --- | --- |
| O59 | The default period for the review record | 12 months, the same as working evidence. It is a setting. |
| O60 | Seven ADR-07 rules chosen to fill gaps (ADR-07 section 13) | Accept all seven. |
| O61 | Backup encryption with a public key, or a shared key | Public key: a stolen server cannot read old backups. |
| O62 | Supported PostgreSQL versions | 16, 17 and 18, plus each newer major version once CI passes on it. Only 16 is tested so far. Version 15 leaves community support in November 2027, so it is not worth starting on. |
| O63 | May the mover undo a forward move inside the window? | Yes. It is the cheapest fix for a slip, and both events are kept. |
| O64 | Limits added by the schema audit: a cohort of at most five years, a membership that cannot end after its cohort, an undo window of at most one hour, ranges for the other numeric settings | Accept. They close ways around finite retention and the undo rule. |
| O65 | Restore replays the whole erasure log, not only entries newer than the dump | Accept. It corrects DEC60. |
| O66 | A fourth, read-only database role for backups | Accept. No existing role can take a backup safely. |

Other open items.

| ID | Question | Needed by |
| --- | --- | --- |
| O30 | Phase 0: the repository fixes, the rename to MILE and the document updates (CI already passes) | Now |
| O18 | Threat model and operator hardening guide | Phase 2 (threat model); Phase 8 (guide) |
| O2 | Adoption test: 10 interviews and a paper prototype with 5 mentor and intern pairs | Before the Phase 4 UI is final |
| O31 | The other two prompt triggers (a task overrunning three times its time allowance; a third blocked entry in seven days): keep, change or drop. Recommendation: same rule as backward moves, built in Phase 5. | Phase 5 |
| O29 | Data Handling and Integrity document | Draft done; final in Phase 7 |
| O6 | Paid feature list and pricing | Phase 9 |
| O26 | Licence file details and EULA | Phase 9 |
| O1, O28 | Employer release, selling structure | Public launch (Phase 8) and first sale (Phase 9) |
| O21 | Legal checks that apply only if circumstances change | When a trigger happens |

## 11. Next implementation step

1. Answer O59 to O66. The schema is final once O60, O63, O64 and O66 are settled.
2. Phase 0: the repository fixes, the rename to MILE, the removal of Valkey, and the document changes listed above, including ADR-04 to ADR-06.
3. Phase 1: re-verify in CI.
4. Phase 2: write the threat model, bring the migrations into the repository and build the persistence layer on them.

Start the validation track at the same time as Phase 2.

## Appendix A. Plan defaults (the former D1 to D17)

| ID | Default | State | Phase |
| --- | --- | --- | --- |
| D1 | Project name and module path | Superseded: renamed to MILE (DEC83) | 0 |
| D2 | Pin a supported Go release and CI matrix | In force; fix the patch-level pin (F3) | 0 |
| D3 | Valkey | Withdrawn; no Redis-family dependency (DEC25) | 0 |
| D4 | Email and password with Argon2id, database sessions, secure cookies, CSRF, rate limits | Accepted | 3 |
| D5 | One organisation per installation | Accepted; Community manages one department, schema allows many | 2 |
| D6 | Add tables incrementally with their features | Accepted | 2 onward |
| D7 | Skill tags; each task needs at least one tag | Accepted, with a seeded default tag "General" pre-selected. In the schema, with tag history | 2, 4 |
| D8 | `time_in_prev_state_ms` internally | Accepted; in the schema | 2 |
| D9 | Retention-bounded evidence and controlled erasure | In force; designed in section 4 and the Phase 2 schema | 2, 7 |
| D10 | No anonymity claim from hashing; aggregates need suppression rules | In force; finalise before the Phase 7 gate. Community stores no aggregates (DEC63) | 2, 7 |
| D11 | SLA sweep every 15 minutes by default | Accepted | 5 |
| D12 | Review checkpoints from the membership join date, falling back to cohort start | Accepted; checkpoint days are set per cohort (DEC55) | 6 |
| D13 | Fairness rules need worked examples | Rules accepted (section 5.2); examples due before Phase 6 aggregation | 2, 6 |
| D14 | Blocked oscillation on the third entry in a rolling seven days | Accepted as a definition, where an entry is each time the blocked flag is set (DEC49); whether it prompts is open (O31) | 5 |
| D15 | Shared 1-on-1 agenda, private notes, human return-offer decision | Private notes replaced by notebooks | 6 |
| D16 | Installation identifier for licence binding | In force; in the schema | 9 |
| D17 | Licence terms | MIT core, DCO, CC-BY-4.0 docs, MIT deploy files; lawyer to confirm; EULA before Phase 9 | 0, 9 |

## Appendix B. What changed since version 3.2

| Area | Change | Source |
| --- | --- | --- |
| Name (4.1) | MILE is the product name; the rename is a Phase 0 work item with no condition; the schema uses the `mile` prefix | DEC83 |
| CI (4.1) | CI runs and passes; Phase 0 keeps it green | DEC84 |
| Security policy (4.1) | `SECURITY.md` rewritten against this plan and ADR-08 | Request of 21 Nov |
| Structure | Rules regrouped by topic in section 3; a status section added; the change tables of versions 3.0 to 3.2 removed; "Decisions in force" split between sections 3 to 6; the secrets table moved to ADR-08 | Request of 15 Nov |
| Data contract | Section 4 rewritten around the Data Handling document: personal data limited to name and work email, one subject per record, two retention periods, the full retention table, who starts export and erasure | DEC59 to DEC68 |
| Sovereignty | Air-gapped operation and support by email stated; tracing kept, for the customer's own collector only; the outbound list completed | DEC73, DEC74 |
| Aggregates | Community stores none; benchmarks move to the Enterprise comparison | DEC63 |
| Blocked cards | A free-text note, and mentor approval for an intern's block | DEC69 |
| Workflow | Immutable versions with column maps, the removal rule, the reordering rule and abandoned tasks; ADR-07 written | DEC78 to DEC81 |
| Slots | Counted per person; one shared pool with an optional cap per department | DEC76, DEC77 |
| Phase 2 | Split into what is designed and tested and what remains; the gate gains the inventory and restore checks | Schema work, 15 Nov |
| Phases 3 to 9 | Token hashes and sign-in data rules; workflow administration; block approval; CSV import limits; message content; streamed exports; the restore steps; legal hold in the database; retention under downgrade | DEC64, DEC65, DEC69 to DEC71, DEC78, schema audit |
| Schema audit | The reason list follows the mover; tag and text-edit history; no task change without an event; published sets closed; the settings change log and the erasure log written by the database; a backup role; limits on cohort length and settings | Schema audit, 15 Nov; O60 to O66 |
| Restore | Replays the whole erasure log | O65, correcting DEC60 |
| Phase 10 | "Tracing within the telemetry privacy policy" replaced: no such policy existed | DEC74 |

Earlier history, from version 2.1 to 3.2, is in the decisions workbook.
