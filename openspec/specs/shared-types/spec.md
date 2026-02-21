## ADDED Requirements

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
