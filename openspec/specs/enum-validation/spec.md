## Purpose

Capability spec for enum-validation. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Valid and Values methods on existing enums
The system SHALL add `Valid() bool` and `Values() []T` methods to all existing typed enum types.

#### Scenario: ApprovalPolicy enum methods
- **WHEN** `config/types.go` defines `ApprovalPolicy`
- **THEN** it SHALL have `Valid()` and `Values()` methods

#### Scenario: PIICategory enum methods
- **WHEN** `agent/pii_pattern.go` defines `PIICategory`
- **THEN** it SHALL have `Valid()` and `Values()` methods

#### Scenario: KeyType enum methods
- **WHEN** `security/key_registry.go` defines `KeyType`
- **THEN** it SHALL have `Valid()` and `Values()` methods

#### Scenario: SectionID enum methods
- **WHEN** `prompt/section.go` defines `SectionID`
- **THEN** it SHALL have `Valid()` and `Values()` methods

#### Scenario: StreamEventType enum methods
- **WHEN** `provider/provider.go` defines `StreamEventType`
- **THEN** it SHALL have `Valid()` and `Values()` methods

#### Scenario: SafetyLevel enum methods
- **WHEN** `agent/runtime.go` defines `SafetyLevel`
- **THEN** it SHALL have `Valid()` and `Values()` methods

#### Scenario: Background Status enum methods
- **WHEN** `background/task.go` defines `Status`
- **THEN** it SHALL have `Valid()` and `Values()` methods

#### Scenario: ContextLayer enum methods
- **WHEN** `knowledge/types.go` defines `ContextLayer`
- **THEN** it SHALL have `Valid()` and `Values()` methods

### Requirement: Package-local enum types
The system SHALL convert untyped string constants to typed enums with `Valid()`/`Values()` in their respective packages.

#### Scenario: graph.Predicate typed enum
- **WHEN** `graph/store.go` defines predicate constants
- **THEN** they SHALL be typed as `Predicate string` with `Valid()` and `Values()`

#### Scenario: knowledge.KnowledgeCategory typed enum
- **WHEN** `knowledge/types.go` defines category constants
- **THEN** they SHALL be typed as `KnowledgeCategory string` with `Valid()` and `Values()`

#### Scenario: skill.SkillStatus and SkillType typed enums
- **WHEN** `skill/types.go` defines status and type constants
- **THEN** they SHALL be typed enums with `Valid()` and `Values()`

### Requirement: ResponseStatus enum type
The system SHALL define `ResponseStatus` as a typed string enum in `protocol/messages.go` with constants `ResponseStatusOK`, `ResponseStatusError`, `ResponseStatusDenied`, `ResponseStatusPaymentRequired` and a `Valid()` method.

#### Scenario: Response.Status uses typed enum
- **WHEN** the protocol handler constructs a `Response`
- **THEN** it SHALL set `Status` using `ResponseStatus` constants, never raw strings

#### Scenario: JSON wire format preserved
- **WHEN** a `Response` with `ResponseStatus` is serialized to JSON
- **THEN** the `status` field SHALL contain the plain string value (e.g., `"ok"`)

### Requirement: ACLAction enum type
The system SHALL define `ACLAction` as a typed string enum in `firewall/firewall.go` with constants `ACLActionAllow`, `ACLActionDeny` and a `Valid()` method.

#### Scenario: ACLRule.Action uses typed enum
- **WHEN** an `ACLRule` is constructed
- **THEN** the `Action` field SHALL be `ACLAction` type, not raw string

### Requirement: WildcardAll constant
The system SHALL define `WildcardAll = "*"` in `firewall/firewall.go`.

#### Scenario: Wildcard comparisons use constant
- **WHEN** firewall code checks for wildcard peer or tool patterns
- **THEN** it SHALL compare against `WildcardAll`, not the literal `"*"`

### Requirement: ProofScheme enum type
The system SHALL define `ProofScheme` as a typed string enum in `zkp/zkp.go` with constants `SchemePlonk`, `SchemeGroth16` and a `Valid()` method.

#### Scenario: ZKP config and proof use typed scheme
- **WHEN** `Config.Scheme`, `ProverService.scheme`, or `Proof.Scheme` stores a proving scheme
- **THEN** it SHALL use the `ProofScheme` type

### Requirement: SRSMode enum type
The system SHALL define `SRSMode` as a typed string enum in `zkp/zkp.go` with constants `SRSModeUnsafe`, `SRSModeFile` and a `Valid()` method.

#### Scenario: ZKP config uses typed SRS mode
- **WHEN** `Config.SRSMode` or `ProverService.srsMode` stores the SRS mode
- **THEN** it SHALL use the `SRSMode` type

### Requirement: KMSProviderName enum type
The system SHALL define `KMSProviderName` as a typed string enum in `security/kms_factory.go` with constants `KMSProviderAWS`, `KMSProviderGCP`, `KMSProviderAzure`, `KMSProviderPKCS11` and a `Valid()` method.

#### Scenario: NewKMSProvider accepts typed name
- **WHEN** `NewKMSProvider` is called
- **THEN** the `providerName` parameter SHALL be `KMSProviderName` type

### Requirement: ChainID type and constants
The system SHALL define `ChainID` as a typed `int64` in `wallet/wallet.go` with constants `ChainEthereumMainnet` (1), `ChainBase` (8453), `ChainBaseSepolia` (84532), `ChainSepolia` (11155111).

#### Scenario: NetworkName uses typed constants
- **WHEN** `NetworkName()` switches on a chain ID
- **THEN** it SHALL compare against `ChainID` constants

### Requirement: CurrencyUSDC constant
The system SHALL define `CurrencyUSDC = "USDC"` in `wallet/wallet.go`.

#### Scenario: All USDC references use constant
- **WHEN** any package references the USDC currency ticker
- **THEN** it SHALL use `wallet.CurrencyUSDC` instead of the string literal `"USDC"`
