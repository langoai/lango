## MODIFIED Requirements

### Requirement: PublicKey() Added to WalletProvider Interface

The `WalletProvider` interface SHALL be extended with a `PublicKey(ctx context.Context) ([]byte, error)` method. This method SHALL return the compressed secp256k1 public key bytes (33 bytes) corresponding to the wallet's private key. The private key MUST NOT be exposed by this or any other method. The returned public key SHALL be deterministic: repeated calls with the same wallet MUST return the same bytes.

`PublicKey` is required for P2P identity derivation: the DID system calls `WalletProvider.PublicKey()` to derive `did:lango:<hex-pubkey>` and the corresponding libp2p peer ID. P2P is gated on `payment.enabled`; therefore `PublicKey` is only called when a wallet is present.

#### Scenario: PublicKey returns 33-byte compressed public key
- **WHEN** `WalletProvider.PublicKey(ctx)` is called on a wallet initialized with an ECDSA keypair
- **THEN** the method SHALL return a 33-byte slice (compressed secp256k1 format, prefix `0x02` or `0x03`)

#### Scenario: PublicKey is deterministic
- **WHEN** `WalletProvider.PublicKey(ctx)` is called multiple times on the same wallet
- **THEN** all calls SHALL return identical byte slices

#### Scenario: PublicKey never exposes private key
- **WHEN** `WalletProvider.PublicKey(ctx)` is called
- **THEN** the returned bytes SHALL contain only the public key; the private key bytes SHALL NOT appear in any return value or log output

#### Scenario: All WalletProvider implementations satisfy interface
- **WHEN** the codebase is compiled
- **THEN** all types that implemented the previous `WalletProvider` interface (`LocalWallet`, `RPCWallet`, `CompositeWallet`) SHALL implement `PublicKey` and satisfy the updated interface at compile time

#### Scenario: PublicKey error propagates to DID derivation
- **WHEN** `WalletProvider.PublicKey(ctx)` returns an error (e.g., wallet not initialized, RPC failure)
- **THEN** `WalletDIDProvider.DID(ctx)` SHALL return a wrapped error containing "get wallet public key" and SHALL NOT cache a nil result

---

### Requirement: Existing WalletProvider Methods Unchanged

The pre-existing `WalletProvider` methods SHALL remain unchanged in signature and semantics:

- `Address(ctx context.Context) (string, error)` — returns the wallet's checksummed Ethereum address
- `Balance(ctx context.Context) (*big.Int, error)` — returns native token balance in wei
- `SignTransaction(ctx context.Context, rawTx []byte) ([]byte, error)` — signs a raw transaction
- `SignMessage(ctx context.Context, message []byte) ([]byte, error)` — signs an arbitrary message

Adding `PublicKey` to the interface constitutes a breaking change for any external implementations. Internal implementations (`LocalWallet`, `RPCWallet`, `CompositeWallet`) MUST all implement `PublicKey` before the interface change is merged.

#### Scenario: Existing wallet methods continue to function
- **WHEN** `WalletProvider.Address(ctx)` is called on any implementation after the interface extension
- **THEN** the method SHALL return the same result as before the change (no regression)

#### Scenario: Compile-time interface compliance check
- **WHEN** the package containing `LocalWallet` is compiled
- **THEN** the compile-time assertion `var _ WalletProvider = (*LocalWallet)(nil)` SHALL succeed without error
