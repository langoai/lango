---
title: Exportability
---

# Exportability

Lango's first exportability slice decides early knowledge-exchange tradeability from source lineage, not from the final rendered artifact alone.

## What It Covers

This first slice treats exportability as a source-primary policy. It is designed for early knowledge exchange, where the key question is whether an artifact is tradeable based on where it came from and how it was assembled.

The current surface includes:

- source-class metadata on knowledge assets,
- source-primary exportability evaluation,
- audit-backed exportability decision receipts,
- operator visibility in `lango security status`.

## Operator Surface

Enable exportability policy with the security config:

```json
{
  "security": {
    "exportability": {
      "enabled": true
    }
  }
}
```

When enabled, `lango security status` shows:

- `Exportability: enabled`
- `"exportability_enabled": true` in JSON output

When disabled, the operator surface stays conservative and the exportability policy path remains off.

## What It Is Not Yet

This is the first slice, not the full long-term policy system.

It does not yet include:

- a policy-rule DSL,
- human override UI,
- dispute-ready unified receipt handling,
- a claim that sanitization alone determines tradeability.

## Related Docs

- [Security Overview](index.md)
- [Trust, Security & Policy Audit](../architecture/trust-security-policy-audit.md)
- [P2P Knowledge Exchange Track](../architecture/p2p-knowledge-exchange-track.md)
