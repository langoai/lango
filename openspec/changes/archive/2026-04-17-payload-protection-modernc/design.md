## Context

With brokered DB ownership in place, page-level SQLCipher protection is no longer the preferred boundary. Sensitive payloads can instead be protected at the application layer while keeping FTS5-capable redacted projections available to search and recall flows.

## Goals / Non-Goals

**Goals:**
- Move the primary SQLite driver to modernc with FTS5 support.
- Encrypt user-content payloads with broker-managed AEAD.
- Keep recall/search usable through redacted plaintext projections.

**Non-Goals:**
- Supporting existing SQLCipher databases.
- Implementing key rotation or re-encryption workflows in v1.
- Moving graph persistence off BoltDB.

## Decisions

- Keep `key_version=1` fixed in v1 and defer rotation to a follow-up change.
- Treat projection generation failure as `projection empty`, never as plaintext fallback.
- Commit ciphertext, projection, and FTS/recall rows in the same broker transaction.

## Risks / Trade-offs

- [Risk] Search quality may drop if redaction is aggressive. → Mitigation: use summaries/snippets rather than raw-field projection.
- [Risk] Legacy encrypted DB users lose in-place upgrade. → Mitigation: fail fast with a clear remediation message.
