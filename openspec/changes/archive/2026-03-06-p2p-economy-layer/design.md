## Context

Lango has a complete P2P infrastructure layer (libp2p networking, DID identity, Noise handshake, reputation scoring, USDC settlement, X402 payments, paygate). However, agents cannot autonomously make economic decisions — budgets are unmanaged, risk is unassessed, prices are static, and there is no negotiation or escrow mechanism. This change adds 5 economy subsystems that sit between the P2P infrastructure and agent tools.

## Goals / Non-Goals

**Goals:**
- Enable agents to allocate, track, and enforce per-task spending budgets
- Assess transaction risk using trust scores, amounts, and output verifiability
- Support dynamic pricing with trust/volume discounts
- Allow P2P price negotiation with auto-negotiation capability
- Provide milestone-based escrow for high-value transactions
- Wire all subsystems into the existing app lifecycle, event bus, and P2P protocol

**Non-Goals:**
- Symphony orchestration (multi-agent workflow coordination)
- On-chain escrow smart contracts (uses no-op settler placeholder)
- Ent schema for escrow persistence (in-memory store only)
- Real-time market-based pricing (rule-based only)

## Decisions

### 1. Callback pattern over direct imports
Economy packages define local function types (e.g., `risk.ReputationQuerier`, `budget.RiskAssessor`) instead of importing P2P packages directly. This avoids import cycles and keeps the economy layer independently testable.

**Alternative**: Interface-based dependency injection. Rejected because function types are simpler for single-method callbacks and match existing patterns in the codebase (paygate.PricingFunc, protocol.ToolExecutor).

### 2. math/big.Int for all monetary values
All USDC amounts use `*big.Int` in the smallest unit (6 decimals). This prevents floating-point precision errors in financial calculations.

**Alternative**: Custom Money type. Rejected as over-engineering for current needs; big.Int is sufficient and widely understood.

### 3. In-memory stores with interface-backed persistence
Budget and escrow use in-memory stores behind interfaces (budget.Store, escrow.Store). This allows future migration to Ent/DB persistence without changing engine logic.

### 4. Event bus integration for cross-system coordination
Economy events (budget alerts, negotiation state changes, escrow milestones) are published through the existing eventbus.Bus rather than direct callbacks. This decouples producers from consumers.

### 5. Interface{} fields in App struct for economy components
Economy engine fields in App use `interface{}` type to avoid importing economy packages in the core app/types.go file, keeping the dependency graph clean. The wiring file holds the concrete types.

## Risks / Trade-offs

- **[In-memory store data loss]** → Budget and escrow data is lost on restart. Mitigation: Designed with Store interface for future Ent persistence migration.
- **[No-op escrow settlement]** → Escrow funds are not actually locked on-chain. Mitigation: SettlementExecutor interface allows wiring real settlement in a future change.
- **[Negotiation state not persisted]** → Active negotiations are lost on restart. Mitigation: Sessions are short-lived (5min default timeout), acceptable for MVP.
- **[Single-node negotiation]** → Negotiation engine is local; P2P negotiation requires both peers to have the protocol handler wired. Mitigation: P2P protocol messages (RequestNegotiatePropose/Respond) enable cross-node negotiation.
