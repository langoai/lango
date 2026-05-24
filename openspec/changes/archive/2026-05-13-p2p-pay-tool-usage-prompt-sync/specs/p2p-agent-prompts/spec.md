## ADDED Requirements

### Requirement: P2P pay prompt wording stays truth-aligned

The agent tool-usage prompt SHALL describe `p2p_pay` using the same receipt-linked required-input contract enforced by the tool wrapper.

#### Scenario: p2p_pay prompt lists required transaction receipt id
- **WHEN** the agent reads the P2P networking section in `TOOL_USAGE.md`
- **THEN** `p2p_pay` SHALL describe `peer_did`, `transaction_receipt_id`, and `amount` as required inputs
- **AND** SHALL describe `submission_receipt_id` and `memo` as optional inputs
