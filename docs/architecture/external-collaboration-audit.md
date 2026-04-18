# External Collaboration & Economic Exchange Audit

## Purpose

This document is the first detailed audit ledger under the Lango master document.

It exists to review the product area that most directly defines Lango:

- P2P identity,
- trust,
- reputation,
- pricing,
- negotiation,
- settlement,
- team formation,
- shared artifacts.

## Audit Order

1. P2P identity / trust / reputation
2. pricing / negotiation / settlement
3. team formation / role coordination
4. workspace / shared artifacts

## Audit Method

Each feature family should be judged by:

- capability area fit,
- product-path fit,
- current user-facing surface,
- duplication risk,
- trust or policy gaps,
- judgment,
- owning track.

Allowed judgments:

- `keep`
- `stabilize`
- `merge`
- `defer`
- `remove`

## Current Surface Map

| Feature family | Primary phase | Current surface clues | Audit status |
| --- | --- | --- | --- |
| P2P identity / trust / reputation | Phase 1 | `docs/features/p2p-network.md`, `docs/features/economy.md`, `internal/config/types_p2p.go`, `internal/cli/p2p/`, `internal/cli/settings/forms_p2p.go` | Ready for detailed audit |
| pricing / negotiation / settlement | Phase 1-2 | `docs/features/economy.md`, `docs/payments/usdc.md`, `docs/payments/x402.md`, `internal/config/types_economy.go`, `internal/cli/economy/`, `internal/cli/payment/` | Ready for detailed audit |
| team formation / role coordination | Phase 3 | `docs/features/p2p-network.md`, `docs/features/multi-agent.md`, `internal/config/types_p2p.go`, `internal/config/types_orchestration.go`, `internal/cli/p2p/`, `internal/cli/agent/` | Ready for detailed audit |
| workspace / shared artifacts | Phase 3-4 | `docs/features/p2p-network.md`, `docs/features/provenance.md`, `internal/config/types_p2p.go`, `internal/cli/p2p/`, `internal/cli/provenance/` | Ready for detailed audit |

## Baseline Decisions Already Locked

- External collaboration is economically native.
- Trust is mixed: cryptographic continuity first, transaction history at the center.
- Root accountability belongs to the owner.
- Agent-level reputation stays separate from owner-level root trust.
- Early trade is bounded by allowlists plus explicit exportability policy.
- The default early external exchange is deliverable-oriented, not broad execution.
- On-chain stablecoin is the trust anchor for settlement.
- Off-chain accrual opens only after trust is earned.
- Team formation is leader-led by default.
- Shared artifacts are leader-owned and selectively exposed by scope.

## Next Plan

The next implementation plan after this document lands should perform the detailed audit for the first row:

- P2P identity / trust / reputation
