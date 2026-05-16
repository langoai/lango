## ADDED Requirements

### Requirement: RPC wallet address cleanup holds across timeout and companion-error exits

RPC wallet coverage SHALL verify that pending address request bookkeeping is cleaned up not only after success, sender error, and cancellation, but also after timeout and companion-error exits.

#### Scenario: Address timeout leaves no pending request entry
- **WHEN** `RPCWallet.Address` exits because no response arrives before the timeout
- **THEN** the pending address request map SHALL be empty after the call returns

#### Scenario: Address companion error leaves no pending request entry
- **WHEN** `RPCWallet.Address` exits because the companion responds with an error
- **THEN** the pending address request map SHALL be empty after the call returns
