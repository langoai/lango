## Context

P2P networking, A2A protocol, and Payment subsystems exist independently. Agents can discover peers and establish sessions, but cannot execute paid tool invocations. The protocol handler lacks an executor callback, ZK circuits are compiled but unwired, and no pricing/payment flow exists.

## Goals / Non-Goals

**Goals:**
- Enable paid tool invocations between P2P agents using EIP-3009 USDC authorization
- Protect owner privacy with a hard-block PII filter on all P2P responses
- Track peer trust via reputation scoring to prevent abuse
- Wire existing ZK circuits for handshake verification and response attestation
- Connect the protocol handler executor so P2P tool calls actually work

**Non-Goals:**
- Full on-chain transaction submission (MVP uses placeholder; real tx requires seller-side signing)
- Cross-chain payment bridging
- Dispute resolution or escrow mechanisms
- Automated pricing optimization

## Decisions

### Pre-Signed Authorization (EIP-3009) over Direct Transfer
**Decision:** Buyer signs an EIP-3009 `transferWithAuthorization` off-chain; seller verifies and submits on-chain after service delivery.
**Rationale:** Prevents buyer fraud (seller holds auth, can submit anytime before deadline). Eliminates gas costs for buyer. Standard USDC mechanism supported on all major chains.
**Alternative:** Direct ERC-20 transfer before tool execution — rejected because seller could take payment without delivering service.

### Adapter Pattern for PayGate ↔ Handler
**Decision:** Use `payGateAdapter` struct to bridge `paygate.Gate` (concrete) to `protocol.PayGateChecker` (interface).
**Rationale:** Avoids import cycle between protocol and paygate packages. Handler only depends on its own interface type.

### Reputation Score Formula
**Decision:** `score = successes / (successes + failures*2 + timeouts*1.5 + 1.0)`
**Rationale:** New peers start at 0.0 (not trusted by default). Failures weigh double, timeouts weigh 1.5x. The +1.0 denominator prevents division-by-zero and requires at least a few successful exchanges to build trust. Default min threshold of 0.3 allows peers through after ~1-2 successful exchanges.

### Closure-Based Executor over App Method
**Decision:** Wire P2P executor as a closure capturing the tools slice, rather than an App method.
**Rationale:** Tools list is a local variable in `New()`. A closure avoids adding a new field to the App struct or leaking the tools slice. The closure directly dispatches to tool handlers by name.

### Owner Shield as Firewall Layer
**Decision:** Owner Shield is integrated into the Firewall's `SanitizeResponse()` rather than as a separate middleware.
**Rationale:** Single point of enforcement. All P2P responses pass through firewall sanitization. No way to bypass the shield regardless of payment amount.

## Risks / Trade-offs

- **[MVP: No real on-chain submission]** → SubmitOnChain returns placeholder hash. Mitigation: documented as TODO, auth is verified and can be submitted when seller-side signing is implemented.
- **[Reputation cold start]** → New peers have 0.0 score but are allowed through (benefit of doubt). Mitigation: minTrustScore default 0.3 is low enough that 1-2 successful exchanges suffice.
- **[ZK circuit compilation cost]** → All 4 circuits compiled at startup. Mitigation: compilation is one-time; results can be cached to disk via ProofCacheDir config.
- **[Single-chain assumption]** → Payment gate assumes one chain per node. Mitigation: canonical registry supports multiple chains; multi-chain can be added later.
