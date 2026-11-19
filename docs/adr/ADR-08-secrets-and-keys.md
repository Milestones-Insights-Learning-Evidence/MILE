# ADR-08: Secrets and keys

- Status: Accepted, 15 November 2026, except section 5 (proposed, O61). Revised the same day after the schema audit.
- Decisions: DEC33, DEC42, DEC60, DEC65, DEC73, DEC74, DEC75
- Plan: implementation plan v4.0, section 6.5; needed before the Phase 2 schema

## Context

The customer holds every key (DEC32). The software encrypts three things itself: secrets stored in the database, notebook notes and backups (DEC75). Everything else relies on the operator's disk or database encryption. The product must run on an air-gapped network, so no key service, licence server or update channel can be assumed (DEC74).

## Decision

### 1. Inventory

| Secret | Where it lives | Form | Rotation |
| --- | --- | --- | --- |
| Database credentials for the migration, application, retention and backup roles | Environment or `_FILE` | As issued by the operator | Operator's procedure |
| Master key | Environment or `_FILE` | 32 random bytes, base64 | `rotate-master-key` command re-encrypts settings secrets and re-wraps the notebook key |
| Mail, Slack and identity-provider secrets | `settings.secret` | Encrypted with a key derived from the master key | With the master key, or when the administrator replaces the value |
| Notebook data key (Phase 6) | Database, wrapped by the master key | 32 random bytes | Rewrap with the master key; full re-encryption procedure written in Phase 6 |
| Backup key | Outside the backup folder | See section 5 | New key for new backups; old key kept until its last backup is pruned |
| Session, invitation and reset tokens | Database, as SHA-256 hashes (DEC65) | 32 random bytes each, sent to the browser once | Expire; a restore revokes all of them |
| Passwords | Database, as Argon2id hashes (plan D4) | Parameters stored with each hash | Re-hashed on sign-in when parameters change |
| Session and CSRF signing secrets, bootstrap secret, metrics token | Generated locally, or `_FILE` | 32 random bytes | Rotation ends sessions safely |
| Enterprise licence signing key | Offline, never on a build server | Ed25519 | Phase 9 |
| Enterprise licence verification key | Embedded in the Enterprise binary | Ed25519 public key | With a new release |

### 2. Encryption of stored values

- AES-256-GCM. Each purpose gets its own key, derived from the master key with HKDF-SHA256 and a purpose label (`settings`, `notebook-wrap`, `key-check`).
- Each encrypted value starts with a one-byte format version and a key identifier, so a value encrypted under an old master key can be found and re-encrypted.
- The row identifier is bound as associated data, so an encrypted value copied to another row fails to decrypt.

### 3. Start-up checks

- `installation.key_check` holds a known value encrypted with the master key, written when the first secret is saved.
- Once `key_check` exists, `serve` refuses to start if the master key is missing or does not decrypt it. The error says which.
- `serve` refuses to start when connected as the migration role, a table owner or a superuser.

### 4. What never holds a secret

- The API never returns a stored secret after it is saved; it reports only whether one is set.
- Exports, logs, the audit log, the settings change log and the support bundle never contain a secret (DEC73). The settings change log is written by the database and records only that a secret changed.
- The erasure log and the job and outbox payloads hold identifiers only (DEC64). For job payloads the database enforces the shape: a name, an email address or a sentence does not fit.

### 5. Backup encryption (proposed, O61)

Backups are encrypted to a public key in the `age` format (X25519). The server holds only the public key, so it can write backups it cannot read. The private key is needed only for `restore` and `verify-backup`, and the operator keeps it offline, with the master key and away from the backups.

A restored database is fully readable only with both the backup private key and the master key. The Phase 7 restore test proves both are needed and that together they are enough.

The alternative is a symmetric key given to the `backup` command (plan v3.2 as written). It is simpler, but a stolen server then also yields the means to read every backup.

Dependency: `filippo.io/age` (BSD-3-Clause), to be added to the dependency inventory.

### 6. The backup role (proposed, O66)

The `backup` command connects with its own database role, which can read every table and sequence and change nothing. The application role cannot take a backup, because it cannot read the erasure log. The retention role is not used for backups, because it can delete evidence.

### 7. The erasure log (DEC60)

- The database writes an erasure-log entry on every delete of evidence or of an account. The retention process declares its run first, and a delete without a declared run is refused, so the log cannot be incomplete.
- At every retention run and every backup, the server writes the erasure log to the backup folder as JSON lines, with a SHA-256 checksum file.
- The log holds table names, record identifiers, times and run identifiers. It holds no personal content, so it is not encrypted, which keeps it readable even if a key is lost.
- A restore replays every entry, in order. Identifiers are random and never reused, so replaying an entry whose record is already gone does nothing. DEC60 said "entries newer than the dump"; that filter can miss an erasure whose transaction was still open when the dump was taken, so it is dropped (O65).

## Consequences

- Losing the master key makes stored secrets and notebooks unreadable. The install guide says so plainly and tells the operator to keep it with the backup key, not with the backups.
- No key leaves the customer's installation, and no key service is needed, so the design works air-gapped.
- Token hashes mean a stolen database does not give working sessions or sign-in links.
- A database restored under a new name needs its connection rights set again (`db/roles.sql` shows the three statements); `pg_dump` does not carry them. The `restore` command does this.
