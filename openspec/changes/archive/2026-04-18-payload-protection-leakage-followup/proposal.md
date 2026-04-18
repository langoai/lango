## Why

The previous payload-protection work added ciphertext fields and broker-managed AEAD primitives, but several high-sensitivity domains still persist plaintext content in repo-backed storage paths. Session messages, learning entries, inquiries, and agent memory all retain plaintext content in storage, projection, or helper conversion paths that are supposed to be protected.

## What Changes

- Apply broker-backed payload protection to session messages and tool-call payloads across create, append, compaction, and reload paths.
- Apply structured bundle protection to learning and inquiry domains while keeping only redacted projections searchable.
- Restrict agent memory to redacted searchable projections and decrypt-on-return behavior.
- Tighten legacy fallback rules so only legacy plaintext rows may use plaintext reads.

## Capabilities

### Modified Capabilities
- `session-store`: session messages and tool-call payloads are stored as ciphertext plus redacted projections.
- `session-recall`: recall summaries are built from decrypted content but only redacted summaries are stored/indexed.
- `knowledge-store`: learning entries use protected bundles and redacted plaintext projections.
- `proactive-librarian`: inquiry persistence uses protected bundles that preserve question/context/answer together.
- `agent-memory`: persistent ent store keeps searchable redacted projections only and decrypts on return.
- `master-key-envelope`: payload protection remains rooted in the envelope with fixed `key_version=1`.

## Impact

- Affected code: session persistence, recall indexing, learning persistence/search, inquiry persistence, agent-memory persistence/search, shared security helpers, and downstream docs/specs.
- Out of scope: removing raw `EntClient()` / `RawDB()` escape hatches from the storage facade.
