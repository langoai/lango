## 1. Shared Payload Rules

- [x] 1.1 Add shared protection/redaction helpers for text and structured JSON bundles.
- [x] 1.2 Standardize legacy fallback behavior so only rows without ciphertext may use plaintext reads.

## 2. Session And Recall Coverage

- [x] 2.1 Protect session message content and tool-call payloads across create, append, compaction, and reload paths.
- [x] 2.2 Update session recall generation so it uses decrypted content as input and stores only redacted summaries.

## 3. Learning, Inquiry, And Agent Memory Coverage

- [x] 3.1 Protect learning payload fields with bundle storage, redacted projections, and atomic FTS updates.
- [x] 3.2 Protect inquiry payloads with `{question, context, answer}` bundles across save, resolve, and read helper paths.
- [x] 3.3 Restrict agent memory persistence and search to redacted projections with decrypt-on-return behavior.

## 4. Verification

- [x] 4.1 Add payload leakage and decrypt-failure regression tests for session and session-recall flows.
- [x] 4.2 Add payload leakage and round-trip regression tests for learning, inquiry, and agent memory.
- [x] 4.3 Update affected docs/specs and run `go build ./...`, `go test ./...`, `go build -tags fts5 ./...`, `go test -tags fts5 ./...`, and `openspec validate --type change payload-protection-leakage-followup`.
