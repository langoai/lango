## ADDED Requirements

### Requirement: Wallet provider interface
The system SHALL define a `WalletProvider` interface with methods: `Address`, `Balance`, `SignTransaction`, `SignMessage`. All implementations MUST ensure private keys are never returned to callers.

#### Scenario: Get wallet address
- **WHEN** `Address(ctx)` is called on any WalletProvider implementation
- **THEN** the wallet's public Ethereum address is returned as a hex string

#### Scenario: Sign transaction
- **WHEN** `SignTransaction(ctx, rawTx)` is called with a transaction hash
- **THEN** a secp256k1 signature is returned and the private key bytes are zeroed immediately after

### Requirement: Local wallet with encrypted key storage
The system SHALL implement a `LocalWallet` that loads its private key from `SecretsStore` under key `wallet.privatekey`, signs using go-ethereum crypto, and zeroes key bytes immediately after each operation.

#### Scenario: Local wallet signs transaction
- **WHEN** `SignTransaction` is called on LocalWallet
- **THEN** the key is loaded from SecretsStore, the transaction is signed, and the key bytes are zeroed

#### Scenario: Local wallet derives address
- **WHEN** `Address` is called on LocalWallet
- **THEN** the address is derived from the stored private key via publicKey → keccak256

### Requirement: RPC wallet for companion delegation
The system SHALL implement an `RPCWallet` that delegates signing to a companion app via WebSocket RPC, using correlation IDs and 30-second timeout, mirroring the `security.RPCProvider` pattern.

#### Scenario: RPC wallet signs with companion
- **WHEN** `SignTransaction` is called on RPCWallet with a connected companion
- **THEN** a `wallet.sign_tx.request` message is sent and the response is awaited with 30s timeout

#### Scenario: RPC wallet times out
- **WHEN** the companion does not respond within 30 seconds
- **THEN** a timeout error is returned

### Requirement: Composite wallet with fallback
The system SHALL implement a `CompositeWallet` that uses RPCWallet when companion is connected and falls back to LocalWallet, mirroring the `CompositeCryptoProvider` pattern.

#### Scenario: Composite wallet uses primary
- **WHEN** `ConnectionChecker.IsConnected()` returns true
- **THEN** the RPCWallet (primary) is used for signing

#### Scenario: Composite wallet falls back to local
- **WHEN** `ConnectionChecker.IsConnected()` returns false
- **THEN** the LocalWallet (fallback) is used for signing

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
