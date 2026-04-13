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

## Config and Hook Provenance

At session start, the provenance subsystem records a snapshot of the runtime configuration and hook registry as checkpoint metadata. This enables operators to verify that a session's behavior can be attributed to a specific configuration state.

### Config Fingerprint

A SHA-256 hex digest is computed from three configuration inputs relevant to session reproducibility:

1. **Explicit keys** -- Config keys the user explicitly set (sorted for deterministic ordering)
2. **Auto-enabled flags** -- Context subsystems that were auto-enabled during config resolution
3. **Hooks config** -- The full hooks configuration section

The resulting fingerprint is stored as `config_fingerprint` in the checkpoint metadata. Two sessions with the same fingerprint had identical configuration inputs at startup.

### Hook Registry Snapshot

The current state of the tool hook registry is serialized as a JSON array and stored as `hook_registry` in the checkpoint metadata. Each entry contains:

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Hook name |
| `priority` | int | Execution priority |

The snapshot includes both pre-hooks and post-hooks registered at the time of checkpoint creation.

**Example checkpoint metadata:**

```json
{
  "config_fingerprint": "a1b2c3d4e5f6...",
  "hook_registry": "[{\"name\":\"exec-policy\",\"priority\":100},{\"name\":\"output-gatekeeper\",\"priority\":200}]"
}
```

**Note:** The hook registry snapshot may be empty (`[]`) during early module initialization because hooks are registered in a later bootstrap phase. The full snapshot is available once all modules have initialized.

## Notes

- Bundle export requires a local wallet identity so the bundle can be signed with a DID-verifiable signature.
- Bundle import is verify-and-store only. It does not mutate existing session, run, or workspace state.
- Attribution reports join persisted provenance rows with token usage records to produce per-author and per-file summaries.
- Config-backed provenance behavior (`enabled`, auto-checkpoint settings, retention, per-session limits) can be edited through `lango settings` in the Automation section.
- Agent-level `session_isolation` is not part of provenance settings. It remains an `AGENT.md` metadata field.
