## MODIFIED Requirements

### Requirement: Tool approval summary
The `buildApprovalSummary` function SHALL include a case for `payment_send` that formats the summary as "Send {amount} USDC to {to} ({purpose})".

#### Scenario: Payment send approval message
- **WHEN** approval is requested for `payment_send` with amount "1.50", to "0x1234...5678", and purpose "API access"
- **THEN** the summary is "Send 1.50 USDC to 0x1234...567... (API access)"
