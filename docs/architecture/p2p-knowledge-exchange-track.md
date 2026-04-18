# P2P Knowledge Exchange Track

## Goal

Define the first concrete external market activity for sovereign Lango agents:

- expertise access,
- reviewable deliverables,
- bounded exportability,
- trusted settlement,
- dispute-ready receipts.

## Why This Track Comes First

Lango should reach meaningful external economic activity before it tries to support broad external execution or long-running multi-agent organizations.

Knowledge exchange is the narrowest useful slice because it:

- creates real external value,
- forces trust and exportability boundaries to become explicit,
- produces reviewable artifacts,
- creates a clean bridge toward later team execution.

## In Scope

- pseudonymous but cryptographically continuous identities,
- owner-root trust plus agent-specific reputation,
- allowlist plus explicit exportability policy,
- expertise access and reviewable deliverables,
- small upfront payment plus approval-based final settlement,
- on-chain stablecoin as the trust anchor,
- off-chain accrual only after trust is established,
- dispute handling grounded in signed logs, provenance, acceptance criteria, escrow state, and immutable receipts.

## Out of Scope

- unrestricted remote execution,
- full team-based role orchestration,
- live shared workspaces by default,
- complete reputation formula design,
- final smart-contract design.

## Default Transaction Shape

1. A leader agent discovers or selects an external counterparty.
2. The leader agent confirms the target artifact is tradeable under exportability policy.
3. The parties agree on price, delivery window, and deliverable scope.
4. A small upfront payment or escrowed commitment is created.
5. The external agent delivers a reviewable artifact.
6. The leader agent approves, rejects, or disputes the artifact.
7. Final settlement is released on approval or handled through dispute rules.

## Required Follow-On Plans

1. `P2P identity / trust / reputation` detailed audit
2. `pricing / negotiation / settlement` detailed audit
3. exportability policy and approval flow design
4. first implementation plan for the `knowledge exchange` runtime path
