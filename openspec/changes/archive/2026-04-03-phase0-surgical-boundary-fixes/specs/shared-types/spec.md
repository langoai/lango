## ADDED Requirements

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
