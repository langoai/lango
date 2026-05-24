## Context

The broker and payload encryption primitives already exist, but coverage is incomplete. The safest corrective path is to keep the existing schema shape wherever possible and make the remaining storage paths conform to the same rules: ciphertext for original payloads, redacted projection for search/display support, decrypt-on-read, and no silent plaintext fallback for protected rows.

## Goals / Non-Goals

**Goals:**
- Close plaintext leakage in session messages, learning entries, inquiries, and agent memory.
- Use shared redaction/protection helpers to avoid per-domain drift.
- Preserve legacy plaintext read compatibility only for rows written before ciphertext fields existed.

**Non-Goals:**
- Removing storage facade raw handles or enforcing broker-only ownership everywhere.
- Backfilling all historical plaintext rows to ciphertext in this change.
- Introducing key rotation or multiple payload key versions.

## Decisions

- Session tool calls are split into two representations: redacted projection JSON in `tool_calls`, full JSON bundle in `tool_calls_ciphertext`.
- Learning entries store `{error_pattern, diagnosis, fix}` in one protected bundle and keep `trigger` plaintext.
- Inquiry entries reuse the existing single `payload_*` slot for `{question, context, answer}` so resolve operations preserve all three fields.
- Agent memory search remains `key + content projection` only; decrypt is return-time only.
- Projection generation failure stores `projection=""` and never synthesizes fallback plaintext.
- Rows with ciphertext present may not fall back to plaintext projection on decrypt failure.

## Risks / Trade-offs

- Search quality may drop because secrets are aggressively redacted from projections. This is acceptable because leakage prevention is the higher priority.
- Legacy plaintext rows remain readable until a future migration/backfill change. This is acceptable because the scope is corrective coverage, not historical re-encryption.
- Session recall summaries may become less rich after redaction, but they must not leak original payload text into the FTS table.
