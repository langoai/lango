## ADDED Requirements

### Requirement: Knowledge exchange runtime control plane reuses receipt-backed meta tools
The meta tools surface SHALL treat the first knowledge exchange runtime design slice as a composition of the existing receipt-backed tools, with `transaction receipt` as canonical control-plane state and `submission receipt` as canonical deliverable state.

#### Scenario: Runtime slice reuses existing tool contracts
- **WHEN** the knowledge exchange runtime slice is described through meta-tools behavior
- **THEN** it SHALL rely on the existing exportability, approval, submission-creation, upfront-payment, and escrow recommendation tools rather than introducing a duplicate receipt model
