## Purpose

Capability spec for shared-types. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Enum interface
The system SHALL provide a generic `Enum[T any]` interface in `internal/types/enum.go` with methods `Valid() bool` and `Values() []T`.

#### Scenario: Enum interface definition
- **WHEN** a developer creates a new typed enum
- **THEN** they can implement the `Enum[T]` interface to get validation and enumeration capabilities

### Requirement: Shared callback types
The system SHALL define `EmbedCallback`, `ContentCallback`, `Triple`, and `TripleCallback` types in `internal/types/callback.go`, consolidating duplicates from `knowledge/store.go` and `memory/store.go`.

#### Scenario: EmbedCallback consolidation
- **WHEN** `knowledge/store.go` and `memory/store.go` reference `EmbedCallback`
- **THEN** both SHALL import from `internal/types` instead of defining their own

#### Scenario: TripleCallback consolidation
- **WHEN** `learning/graph_engine.go`, `memory/graph_hooks.go`, or `librarian/types.go` reference `TripleCallback`
- **THEN** all SHALL import from `internal/types`

### Requirement: Token estimation consolidation
The system SHALL provide `EstimateTokens(text string) int` and `isCJK(r rune) bool` in `internal/types/token.go`, replacing duplicate implementations in `memory/token.go` and `learning/token.go`.

#### Scenario: Token estimation single source
- **WHEN** any package needs token estimation
- **THEN** it SHALL use `types.EstimateTokens()` from `internal/types`

### Requirement: ChannelType enum
The system SHALL define `ChannelType string` in `internal/types/channel.go` with constants `ChannelTelegram`, `ChannelDiscord`, `ChannelSlack` and `Valid()`/`Values()` methods.

#### Scenario: Channel routing uses typed enum
- **WHEN** `app/sender.go` routes messages by channel
- **THEN** it SHALL use `types.ChannelType` constants instead of raw strings

### Requirement: ProviderType enum
The system SHALL define `ProviderType string` in `internal/types/provider.go` with constants for openai, anthropic, gemini, google, ollama, github and `Valid()`/`Values()` methods.

#### Scenario: Provider initialization uses typed enum
- **WHEN** `supervisor/supervisor.go` initializes providers
- **THEN** it SHALL switch on `types.ProviderType` constants instead of raw strings

### Requirement: MessageRole enum
The system SHALL define `MessageRole string` in `internal/types/role.go` with constants for user, assistant, tool, function, model and `Valid()`/`Values()`/`Normalize()` methods.

#### Scenario: Role normalization
- **WHEN** a message has role "model" or "function"
- **THEN** `Normalize()` SHALL convert "model" to "assistant" and "function" to "tool"

### Requirement: Confidence enum
The system SHALL define `Confidence string` in `internal/types/confidence.go` with constants for high, medium, low and `Valid()`/`Values()` methods.

#### Scenario: Confidence used in librarian and learning
- **WHEN** `librarian/types.go` or `learning/parse.go` reference confidence levels
- **THEN** they SHALL use `types.Confidence` constants

### Requirement: CLI package rename
The system SHALL rename `internal/cli/common/` to `internal/cli/clitypes/` and update all importers.

#### Scenario: Package rename with importer update
- **WHEN** the rename is applied
- **THEN** all 3 importing packages SHALL compile without errors

### Requirement: RPCSenderFunc consolidation
The system SHALL define `RPCSenderFunc func(event string, payload interface{}) error` in `internal/types/sender.go`, replacing duplicates in `security/rpc_provider.go` and `wallet/rpc_wallet.go`.

#### Scenario: Sender function type consolidation
- **WHEN** `security/rpc_provider.go` or `wallet/rpc_wallet.go` define sender functions
- **THEN** both SHALL use `types.RPCSenderFunc` from `internal/types`

### Requirement: DIDPrefix constant in types package
The `internal/types` package SHALL export `DIDPrefix = "did:lango:"` in `types/identity.go`. Both `p2p/identity` and `economy/escrow` MUST reference `types.DIDPrefix` instead of defining local copies.

#### Scenario: p2p/identity uses types.DIDPrefix
- **WHEN** `internal/p2p/identity/identity.go` is compiled
- **THEN** it references `types.DIDPrefix` and does NOT define a local `DIDPrefix` constant

#### Scenario: economy/escrow uses types.DIDPrefix
- **WHEN** `internal/economy/escrow/address_resolver.go` is compiled
- **THEN** it imports `internal/types` (not `internal/p2p/identity`) and references `types.DIDPrefix`

### Requirement: ReputationQuerier type in types package
The `internal/types` package SHALL export `ReputationQuerier func(ctx context.Context, peerDID string) (float64, error)` in `types/reputation.go`. The three former definitions (`economy/pricing.ReputationQuerier`, `economy/risk.ReputationQuerier`, `p2p/paygate.ReputationFunc`) MUST be removed or aliased to this canonical type.

#### Scenario: economy/pricing uses types.ReputationQuerier
- **WHEN** `internal/economy/pricing/engine.go` is compiled
- **THEN** it uses `types.ReputationQuerier` and does NOT define a local `ReputationQuerier` type

#### Scenario: p2p/paygate aliases to types.ReputationQuerier
- **WHEN** `internal/p2p/paygate/trust.go` is compiled
- **THEN** `ReputationFunc` is a type alias for `types.ReputationQuerier`

#### Scenario: wiring adapter code is simplified
- **WHEN** `internal/app/wiring_economy.go` passes a reputation function to both pricing and risk engines
- **THEN** no type-casting closure is needed because both accept `types.ReputationQuerier`
