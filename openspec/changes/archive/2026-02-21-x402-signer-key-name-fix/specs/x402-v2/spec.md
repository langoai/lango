## MODIFIED Requirements

### Requirement: SignerProvider (`internal/x402/signer.go`)
- Interface: `SignerProvider` with `EvmSigner(ctx) (ClientEvmSigner, error)`
- Implementation: `LocalSignerProvider` loads key from SecretsStore under key `"wallet.privatekey"`
- Key material zeroed after signer creation

#### Scenario: Signer loads key using wallet convention
- **WHEN** `LocalSignerProvider.EvmSigner(ctx)` is called
- **THEN** the provider SHALL read the private key from SecretsStore using key name `"wallet.privatekey"`

#### Scenario: Key name matches wallet package
- **WHEN** a wallet is created via `CreateWallet()` and then x402 payment is attempted
- **THEN** the x402 signer SHALL successfully load the same key stored by the wallet package
