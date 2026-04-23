# Architecture

This section describes the internal architecture of Lango, a Go-based AI agent built on Google ADK v1.0.0

<div class="grid cards" markdown>

-   :material-layers-outline: **[System Overview](overview.md)**

    ---

    High-level architecture layers, component interactions, and a visual diagram of how Lango processes messages from channels through the agent runtime to AI providers.

-   :material-file-tree-outline: **[Project Structure](project-structure.md)**

    ---

    Complete directory layout with descriptions of every package, explaining what each module owns and how packages relate to each other.

-   :material-swap-horizontal: **[Data Flow](data-flow.md)**

    ---

    End-to-end message flow, the bootstrap/wiring process, and async buffer patterns used for embedding generation and graph indexing.

-   :material-compass-outline: **[Master Document](master-document.md)**

    ---

    Top-level product constitution, product path, capability areas, and execution-track portfolio for Lango.

-   :material-clipboard-search-outline: **[External Collaboration Audit](external-collaboration-audit.md)**

    ---

    The first audit ledger for the product area that most directly defines Lango: trust, pricing, settlement, teams, and shared artifacts.

-   :material-shield-check-outline: **[Trust, Security & Policy Audit](trust-security-policy-audit.md)**

    ---

    Detailed audit ledger for the policy and safety boundaries that determine whether early knowledge exchange is actually safe to operate.

-   :material-account-check-outline: **[Identity Trust Reputation Audit](identity-trust-reputation-audit.md)**

    ---

    Audit ledger for identity continuity, trust entry, reputation, and revocation in `knowledge exchange v1`.

-   :material-cash-refund: **[Pricing Negotiation Settlement Audit](pricing-negotiation-settlement-audit.md)**

    ---

    Audit ledger for pricing surfaces, negotiation, settlement, and escrow in `knowledge exchange v1`.

-   :material-timer-sand: **[Knowledge Exchange Runtime](knowledge-exchange-runtime.md)**

    ---

    The first transaction-oriented runtime control plane for `knowledge exchange v1`, centered on transaction receipt and submission receipt with explicit current limits.

-   :material-bank-check-outline: **[Settlement Progression](settlement-progression.md)**

    ---

    The first transaction-level settlement progression slice for `knowledge exchange v1`, covering approve, revise, reject, and escalate with explicit current implementation limits.

-   :material-cash-sync: **[Actual Settlement Execution](actual-settlement-execution.md)**

    ---

    The first direct settlement execution slice for `knowledge exchange v1`, connecting `approved-for-settlement` state to real payment execution with explicit current limits.

-   :material-cash-multiple: **[Partial Settlement Execution](partial-settlement-execution.md)**

    ---

    The first direct partial settlement execution slice for `knowledge exchange v1`, executing one canonical partial amount with explicit current limits.

-   :material-safe-square-outline: **[Escrow Release](escrow-release.md)**

    ---

    The first funded-escrow release slice for `knowledge exchange v1`, connecting funded escrow and approved settlement state to real release execution with explicit current limits.

-   :material-cash-refund: **[Escrow Refund](escrow-refund.md)**

    ---

    The first funded-escrow refund slice for `knowledge exchange v1`, connecting review-path funded escrow to refund execution with explicit current limits.

-   :material-hand-back-right-outline: **[Dispute Hold](dispute-hold.md)**

    ---

    The first dispute-linked escrow hold slice for `knowledge exchange v1`, recording hold evidence for funded dispute-ready escrow with explicit current limits.

-   :material-scale-balance: **[Release vs Refund Adjudication](release-vs-refund-adjudication.md)**

    ---

    The first post-hold adjudication slice for `knowledge exchange v1`, recording canonical release-vs-refund branching without yet executing either path.

-   :material-cash-fast: **[P2P Knowledge Exchange Track](p2p-knowledge-exchange-track.md)**

    ---

    The first concrete product track for external sovereign-agent economic activity.

</div>
