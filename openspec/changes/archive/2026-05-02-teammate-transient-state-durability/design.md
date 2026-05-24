# Design

## Archived Audit Closure

This change closes the archived `approval-blocked conditions` follow-up from `openspec/changes/archive/2026-05-01-dynamic-multi-agent-hard-cut/design.md` by cross-reference. The archived document itself remains immutable.

Archive note: this change was validated before archive with `openspec validate teammate-transient-state-durability --strict`, and the resulting spec deltas were synced into main OpenSpec specs before the archive step. Archived changes are not directly re-validatable by name through the current CLI.

## Carried-Forward Follow-Ups

| Source audit item | Status in this change | Notes |
|-------------------|-----------------------|-------|
| `approval-blocked conditions` | `closed` | Mirrored via `teammate_approval_blocked` / `teammate_approval_unblocked` journal events plus `RunSnapshot` teammate blocked fields. |
| `recovery states` | `follow-up` | No production `RecoveryState` writer exists yet, so there is no source state to mirror in this change. |

## Reference Design

This archived OpenSpec design is intentionally concise. The fuller internal decision record for this change lives at:

- `internal-docs/superpowers/specs/2026-05-02-teammate-transient-state-durability-design.md`

That internal design records the substantive implementation decisions that shaped this change, including:

- mirror strategy: journal events plus latest snapshot state
- mirror choke point: `AgentRunStore.UpdateProjection(...)`
- event types and typed payload shape
- snapshot extension through `RunSnapshot` / `snapshot_data`
- best-effort mirror failure policy using logs and metrics
- state-delta rules for block, unblock, replace, and terminal precedence
- live read model boundary (`AgentRunStore` / `AgentRunProjection` remain authoritative)

In concrete terms, this archived change used:

- `teammate_approval_blocked` / `teammate_approval_unblocked` as the durable transition events
- `RunSnapshot` teammate fields as the latest durable blocked-state surface
- best-effort mirroring from the `AgentRunStore.UpdateProjection(...)` choke point
- terminal snapshot clearing without requiring a separate durable unblock event
