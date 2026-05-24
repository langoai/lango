## ADDED Requirements

### Requirement: Knowledge exchange runtime architecture page describes the first control-plane slice
The `docs/architecture/knowledge-exchange-runtime.md` page SHALL describe the first transaction-oriented runtime control-plane design slice for `knowledge exchange v1`, centered on transaction receipt and submission receipt, and SHALL list the current limits of that slice.

#### Scenario: Runtime page shows the bounded slice
- **WHEN** a user reads `docs/architecture/knowledge-exchange-runtime.md`
- **THEN** they SHALL find sections covering the runtime design slice, canonical state, current limits, and follow-on work

### Requirement: P2P knowledge exchange track links the runtime design slice
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL reference `knowledge-exchange-runtime.md` as the first transaction-oriented runtime design slice and SHALL state that the remaining work is runtime implementation and broader progression handling.

#### Scenario: Track page points to the runtime slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find the runtime design slice referenced by name and linked to `knowledge-exchange-runtime.md`
- **AND** the follow-on work SHALL be described as implementation, not redesign of the landed slice

