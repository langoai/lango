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
lango p2p provenance push <peer-did> <session-key> --redaction content
lango p2p provenance fetch <peer-did> <session-key> --redaction content
```

## Notes

- Bundle export requires a local wallet identity so the bundle can be signed with a DID-verifiable signature.
- Bundle import is verify-and-store only. It does not mutate existing session, run, or workspace state.
- Attribution reports join persisted provenance rows with token usage records to produce per-author and per-file summaries.
- Config-backed provenance behavior (`enabled`, auto-checkpoint settings, retention, per-session limits) can be edited through `lango settings` in the Automation section.
- Agent-level `session_isolation` is not part of provenance settings. It remains an `AGENT.md` metadata field.
