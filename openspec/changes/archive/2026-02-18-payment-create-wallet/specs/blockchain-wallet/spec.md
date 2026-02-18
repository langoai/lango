## ADDED Requirements

### Requirement: Wallet creation with encrypted key storage
The system SHALL provide a `CreateWallet` function in the wallet package that generates a new ECDSA private key using `crypto.GenerateKey()`, stores it encrypted in SecretsStore under key `wallet.privatekey`, and returns the derived public address. The private key bytes MUST be zeroed immediately after storage.

#### Scenario: Create new wallet when none exists
- **WHEN** `CreateWallet(ctx, secrets)` is called and no `wallet.privatekey` exists in SecretsStore
- **THEN** a new ECDSA key is generated, stored encrypted under `wallet.privatekey`, and the public address is returned as a hex string

#### Scenario: Reject creation when wallet already exists
- **WHEN** `CreateWallet(ctx, secrets)` is called and `wallet.privatekey` already exists in SecretsStore
- **THEN** an error is returned indicating a wallet already exists, along with the existing wallet's address

#### Scenario: Private key is zeroed after storage
- **WHEN** a new wallet is created successfully
- **THEN** the raw private key bytes are zeroed in memory immediately after being written to SecretsStore
