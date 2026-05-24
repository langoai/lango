## MODIFIED Requirements

### Requirement: p2p_pay Tool for Peer-to-Peer USDC Payment

The system SHALL expose a `p2p_pay` agent tool (safety level: `Dangerous`) that sends a USDC payment on the Base blockchain to a connected peer identified by their DID. The tool SHALL require `peer_did`, `transaction_receipt_id`, and `amount` parameters and MAY accept optional `submission_receipt_id` and `memo`. The tool SHALL NOT be available if the payment service is not initialized.

#### Scenario: Missing transaction receipt id rejected at the wrapper
- **WHEN** `p2p_pay` is called without `transaction_receipt_id`
- **THEN** the tool SHALL return an actionable missing-parameter error before receipt-backed payment evaluation begins
