# ADR-001: Package Boundary Policy

## Status
Accepted (2026-04-02)

## Context

Lango is a single-binary AI agent runtime with 68+ packages under `internal/`. 
A comprehensive restructuring proposal suggested reorganizing into 7 bounded 
contexts (commerce, runtime, capability, knowledge, network, security, platform).

After code-level analysis, we determined that:

1. The existing `app/wiring_*.go` pattern (28 domain-specific wiring files) already 
   provides modular monolith boundaries
2. The `appinit` module system uses Kahn's algorithm for topological dependency resolution
3. The `automation` package provides shared interfaces (`AgentRunner`, `ChannelSender`) 
   for cron/background/workflow
4. The `eventbus` is already used only for integration events, not core flow control
5. Full package restructuring would touch ~1252 .go files with high merge conflict risk
6. The 90-day roadmap rates build modularization as P4 (lowest priority)

## Decision

Maintain the current package structure with **explicit boundary rules** enforced 
by automated tests and linter configuration, rather than wholesale directory 
restructuring.

### Layering Rules

```
types <- leaf packages <- domain packages <- app <- cli <- cmd/lango (composition root)
```

- `internal/types/`: shared constants and function types (no dependencies)
- `internal/finance/`: shared monetary types (no dependencies)  
- Domain packages: own their types, interfaces, and tool builders
- `internal/app/`: cross-domain wiring only (via `wiring_*.go` files)
- `internal/cli/`: presentation layer, calls app services
- `cmd/lango/`: composition root, assembles the application

### Boundary Rules (enforced by archtest + depguard)

1. `internal/economy/*` must NOT import `internal/p2p/*`
2. `internal/p2p/{discovery,handshake,firewall,protocol,agentpool}` must NOT import 
   `internal/economy/*` or `internal/wallet/*` (handshake exempted pending Signer 
   interface extraction)
3. Cross-domain orchestration belongs in `app/wiring_*.go`, not in domain packages
4. Domain packages own their tool builders (per OpenSpec `domain-tool-builders` spec)
5. `eventbus` is for integration events only; core flows use explicit service calls

### Enforcement

- `internal/archtest/boundary_test.go`: import graph analysis via `go list`
- `.golangci.yml`: depguard rules for IDE-level feedback
- OpenSpec `domain-tool-builders` spec: tool builder ownership

## Consequences

### Positive
- No import path churn across 1252+ files
- Boundary violations caught automatically in CI
- Compatible with existing 90-day roadmap priorities
- Each rule is independently enforceable and reversible

### Negative
- Package count remains high (68+); navigability depends on documentation
- Some coupling (e.g., handshake->wallet) requires future interface extraction
- Full bounded context restructuring is deferred, not eliminated

## Revisit When

- Package count exceeds 80
- A new bounded context is needed that doesn't fit current structure
- The handshake->wallet coupling blocks a security requirement
- The team grows beyond 3 concurrent contributors on `internal/`
