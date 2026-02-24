## Context

The `p2p-approval-gaps-fix` change implemented three features: SpendingLimiter auto-approval (`IsAutoApprovable`), inbound P2P tool owner approval (`ToolApprovalFunc` callback), and outbound payment auto-approval. These features are fully wired and tested but undocumented. Seven files need updates: feature docs, README, prompts, HTTP API docs, payment docs, example README, and Makefile.

## Goals / Non-Goals

**Goals:**
- Document the 3-stage inbound approval pipeline (firewall → owner approval → execution) with Mermaid diagram
- Document auto-approval behavior for `autoApproveBelow`, `autoApproveKnownPeers`
- Add missing REST API endpoints (`/api/p2p/reputation`, `/api/p2p/pricing`) with curl examples and JSON responses
- Add missing CLI commands (`lango p2p reputation`, `lango p2p pricing`)
- Add missing config fields to README reference table
- Update tool usage prompts with approval semantics
- Add `test-p2p` Makefile target

**Non-Goals:**
- No code changes to the P2P implementation
- No new features or behavioral changes
- No changes to existing test suites

## Decisions

1. **Documentation structure**: Add "Approval Pipeline" as a new top-level section in `p2p-network.md` positioned between Knowledge Firewall and Discovery, since it builds on the firewall concept and is a core P2P feature.

2. **Mermaid diagram**: Use a flowchart showing the 3-stage gate with auto-approve shortcut path, matching the actual flow in `handler.go` (`RequestToolInvoke` and `RequestToolInvokePaid` methods).

3. **Cross-reference approach**: The `autoApproveBelow` threshold is a cross-cutting concern (payment + P2P). Document it in `usdc.md` with a P2P integration note and link to the P2P approval pipeline section, rather than duplicating content.

4. **Makefile target scope**: `test-p2p` runs `./internal/p2p/...` and `./internal/wallet/...` together because the spending limiter auto-approval is a wallet feature directly consumed by P2P.

## Risks / Trade-offs

- [Docs drift] Documentation may diverge from implementation if approval logic changes → Mitigation: All examples reference actual config field names and endpoint paths verified against source code.
- [Example config values] The p2p-trading example uses high thresholds (`50.00`) for convenience → Mitigation: Added explicit production warning note.
