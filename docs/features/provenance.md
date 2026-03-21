---
title: Session Provenance
---

# Session Provenance

Session Provenance makes checkpoints, lineage, attribution, and provenance bundle exchange durable and inspectable.

## Coverage

- Persistent checkpoints anchored to RunLedger journal positions
- Persistent session tree for root and child session lineage
- Git-aware attribution for workspace operations
- Token-aware reports for sessions without workspace git evidence
- Signed provenance bundle export/import with `none`, `content`, and `full` redaction
- Dedicated P2P provenance transport for remote bundle exchange

## Commands

```bash
lango provenance status
lango provenance checkpoint list --run <id>
lango provenance session tree <session-key> --depth 10
lango provenance session list --limit 50 --status active
lango provenance attribution show <session-key>
lango provenance attribution report <session-key>
lango provenance bundle export <session-key> --redaction content --out bundle.json
lango provenance bundle import bundle.json
```

## Notes

- Bundle export requires a local wallet identity so the bundle can be signed with a DID-verifiable signature.
- Bundle import is verify-and-store only. It does not mutate existing session, run, or workspace state.
- Attribution reports join persisted provenance rows with token usage records to produce per-author and per-file summaries.
