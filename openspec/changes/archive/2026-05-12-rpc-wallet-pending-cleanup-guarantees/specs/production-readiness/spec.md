## ADDED Requirements

### Requirement: RPC wallet cleans pending request state on non-success paths

RPC wallet coverage SHALL verify that pending request bookkeeping is cleaned up not only after successful responses, but also after sender errors and cancelled contexts.

#### Scenario: Address sender error leaves no pending request entry
- **WHEN** `RPCWallet.Address` fails because the sender returns an error
- **THEN** the pending address request map SHALL be empty after the call returns

#### Scenario: SignMessage sender error leaves no pending request entry
- **WHEN** `RPCWallet.SignMessage` fails because the sender returns an error
- **THEN** the pending sign-message request map SHALL be empty after the call returns

#### Scenario: Address context cancellation leaves no pending request entry
- **WHEN** `RPCWallet.Address` exits because the context is cancelled
- **THEN** the pending address request map SHALL be empty after the call returns
