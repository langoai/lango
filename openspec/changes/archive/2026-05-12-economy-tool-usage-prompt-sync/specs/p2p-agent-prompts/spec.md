## ADDED Requirements

### Requirement: Economy escrow tool usage wording stays truth-aligned

The agent tool-usage prompt SHALL describe `economy_escrow_create` using the same required-input contract enforced by the wrapper layer.

#### Scenario: Economy escrow create prompt lists required milestones
- **WHEN** the agent reads the economy tool section in `TOOL_USAGE.md`
- **THEN** `economy_escrow_create` SHALL describe `buyerDid`, `sellerDid`, `amount`, and `milestones` as required inputs
