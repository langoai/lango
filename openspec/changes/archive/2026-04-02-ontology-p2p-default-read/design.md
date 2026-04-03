## Context

The ontology ACL system currently defaults P2P peer principals (`peer:` prefix) to `PermWrite`. This was set during initial implementation but conflicts with the principle of least privilege. Remote peers can write to the local ontology without explicit trust or configuration.

The change touches 5 files across 3 packages: `ontology`, `config`, and `cli/settings`. No new interfaces or dependencies are introduced.

## Goals / Non-Goals

**Goals:**
- Change the default P2P peer permission from `PermWrite` to `PermRead`
- Ensure all default-setting paths (code, config loader, TUI) are consistent
- Maintain backward compatibility via explicit config override

**Non-Goals:**
- Changing the permission model itself (levels, Check logic)
- Adding trust-based automatic escalation
- Modifying the P2P handshake or exchange protocol

## Decisions

1. **Default at constructor level, not just config** — `NewRoleBasedPolicy()` itself defaults to `PermRead`. This ensures safety even if the config layer is bypassed or a caller constructs the policy directly. Alternative: only change config default → rejected because constructor would still be unsafe.

2. **Explicit default in `DefaultConfig()`** — Adding `P2PPermission: "read"` to the config defaults makes the value visible in generated configs and prevents ambiguity when the field is empty. Alternative: rely on zero-value + code default → rejected because empty string triggers different TUI behavior.

3. **No migration script** — Existing users who need write access add one line to config. This is simpler and safer than auto-detecting and preserving previous behavior.

## Risks / Trade-offs

- **[Risk] Existing P2P write workflows break silently** → Mitigation: P2P peers attempting writes will get `ErrPermissionDenied`, which surfaces clearly in logs. Documentation notes the migration path.
- **[Trade-off] Strictness vs convenience** → Chose strictness. Users who want P2P write must explicitly opt in, which is the correct posture for a security-sensitive runtime.
