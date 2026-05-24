---
title: Approval Flow
---

# Approval Flow

Lango's first approval-flow slice covers artifact release approval for `knowledge exchange v1`. It sits on top of exportability: exportability decides whether an artifact is tradeable, and approval flow decides how the release request is classified and recorded.

## What Ships in This Slice

The landed surface is narrow and operator-facing:

- structured artifact release decisions: `approve`, `reject`, `request-revision`, and `escalate`
- outcome records with reason, issue classification, fulfillment assessment, and settlement hint
- audit-backed approval receipts stored as `artifact_release_approval` audit entries
- a release path that consumes the exportability decision state and the requested artifact scope, then reconstructs a minimal receipt for this slice

## Decision Model

- `approve` means the artifact can be released under the current slice.
- `reject` means the request is blocked and should not release.
- `request-revision` means the submitted artifact does not match the requested scope and should be revised.
- `escalate` means the release request needs human review because the receipt or request context is not safe to resolve automatically.

## Operator Notes

This slice is intentionally narrow.

It does not yet include:

- a human approval UI,
- upfront payment approval runtime,
- dispute orchestration,
- partial settlement execution,
- full settlement automation.

The approval outcome is an input to later settlement and dispute work, not a replacement for those systems.

## Related Docs

- [Security Overview](index.md)
- [Exportability Policy](exportability.md)
- [P2P Knowledge Exchange Track](../architecture/p2p-knowledge-exchange-track.md)
- [Trust, Security & Policy Audit](../architecture/trust-security-policy-audit.md)
