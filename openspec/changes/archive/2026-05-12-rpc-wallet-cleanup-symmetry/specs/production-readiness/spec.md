## ADDED Requirements

### Requirement: RPC wallet pending-state cleanup is symmetric across request kinds

RPC wallet coverage SHALL verify that pending request bookkeeping is cleaned up consistently for both transaction-signing and message-signing request lifecycles.

#### Scenario: SignTransaction cleanup holds across response and error paths
- **WHEN** `RPCWallet.SignTransaction` exits through a response, companion error, sender error, timeout, or cancelled context
- **THEN** the pending sign-transaction request map SHALL be empty after the call returns

#### Scenario: SignMessage cleanup holds across response and error paths
- **WHEN** `RPCWallet.SignMessage` exits through a response, companion error, sender error, timeout, or cancelled context
- **THEN** the pending sign-message request map SHALL be empty after the call returns
