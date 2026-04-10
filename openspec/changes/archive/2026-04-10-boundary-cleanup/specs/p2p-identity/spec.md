## MODIFIED Requirements

### Requirement: DID Derivation from Wallet Public Key

The `WalletDIDProvider` SHALL accept a `KeyProvider` interface (single method `PublicKey(ctx) ([]byte, error)`) instead of `wallet.WalletProvider`. The `internal/p2p/identity` package SHALL NOT import `internal/wallet`. The `wallet.WalletProvider` implicitly satisfies `KeyProvider` via Go structural typing.

#### Scenario: KeyProvider interface replaces wallet dependency
- **WHEN** `NewProvider` is called with any type satisfying `KeyProvider`
- **THEN** the provider SHALL use only `PublicKey(ctx)` to derive the DID
- **AND** `internal/p2p/identity` SHALL NOT have an import path to `internal/wallet`

### Requirement: Signature verification uses bytes.Equal

The `VerifyMessageSignature` function SHALL compare recovered public key bytes with DID public key bytes using `bytes.Equal`, not `string()` conversion.

#### Scenario: Byte comparison for signature verification
- **WHEN** `VerifyMessageSignature` compares the recovered public key with the DID's public key
- **THEN** it SHALL use `bytes.Equal(recovered, did.PublicKey)` for the comparison
