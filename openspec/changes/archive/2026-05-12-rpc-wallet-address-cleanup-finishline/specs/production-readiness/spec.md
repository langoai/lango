## ADDED Requirements

### Requirement: RPC wallet address cleanup covers timeout and companion-error exits directly

RPC wallet address lifecycle coverage SHALL explicitly verify timeout and companion-error teardown in addition to the previously covered success and cancellation paths.

#### Scenario: Address timeout clears pending state
- **WHEN** `RPCWallet.Address` exits because the request times out
- **THEN** the pending address request map SHALL be empty after the call returns

#### Scenario: Address companion error clears pending state
- **WHEN** `RPCWallet.Address` exits because the companion responds with an error
- **THEN** the pending address request map SHALL be empty after the call returns
