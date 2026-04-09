## Purpose

Capability spec for escrow-test-coverage. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Solidity forge tests for LangoEscrowHub
The system SHALL have Solidity forge tests covering all LangoEscrowHub contract functions including createDeal, deposit, submitWork, release, refund, dispute, resolveDispute, and getDeal with both success and revert scenarios.

#### Scenario: All Hub contract functions tested
- **WHEN** `forge test` is run in the contracts directory
- **THEN** all Hub test cases pass covering constructor, createDeal (success + 4 reverts), deposit (success + 2 reverts), submitWork (success + 3 reverts), release (success + 1 revert), refund (success + 1 revert), dispute (buyer/seller/2 reverts), resolveDispute (success + 3 reverts), getDeal, and full lifecycle

### Requirement: Solidity forge tests for LangoVault
The system SHALL have Solidity forge tests covering all LangoVault contract functions including initialize, deposit, submitWork, release, refund, dispute, and resolve.

#### Scenario: All Vault contract functions tested
- **WHEN** `forge test` is run in the contracts directory
- **THEN** all Vault test cases pass covering initialize (success + double-init + 6 zero-param reverts), deposit, submitWork, release, refund, dispute, resolve, and full lifecycle

### Requirement: Solidity forge tests for LangoVaultFactory
The system SHALL have Solidity forge tests covering LangoVaultFactory constructor, createVault, getVault, and vaultCount.

#### Scenario: All Factory contract functions tested
- **WHEN** `forge test` is run in the contracts directory
- **THEN** all Factory test cases pass covering constructor, createVault (success + clone usability + multiple), getVault, and vaultCount

### Requirement: Go unit tests for HubClient
The system SHALL have Go unit tests for all HubClient methods using a mock ContractCaller.

#### Scenario: HubClient methods tested with mock
- **WHEN** `go test ./internal/economy/escrow/hub/...` is run
- **THEN** all HubClient tests pass covering CreateDeal, Deposit, SubmitWork, Release, Refund, Dispute, ResolveDispute, GetDeal, and NextDealID with both success and error cases

### Requirement: Go unit tests for VaultClient
The system SHALL have Go unit tests for all VaultClient methods using a mock ContractCaller.

#### Scenario: VaultClient methods tested with mock
- **WHEN** `go test ./internal/economy/escrow/hub/...` is run
- **THEN** all VaultClient tests pass covering Deposit, SubmitWork, Release, Refund, Dispute, Resolve, Status, and Amount

### Requirement: Go unit tests for FactoryClient
The system SHALL have Go unit tests for all FactoryClient methods using a mock ContractCaller.

#### Scenario: FactoryClient methods tested with mock
- **WHEN** `go test ./internal/economy/escrow/hub/...` is run
- **THEN** all FactoryClient tests pass covering CreateVault, GetVault, and VaultCount

### Requirement: Go unit tests for HubSettler and VaultSettler
The system SHALL have Go unit tests for HubSettler and VaultSettler covering interface compliance, mapping operations, no-op methods, accessors, and concurrent safety.

#### Scenario: Settler tests pass
- **WHEN** `go test ./internal/economy/escrow/hub/...` is run
- **THEN** all settler tests pass including interface compliance, mapping roundtrip, concurrent mapping safety, and accessor methods

### Requirement: Go unit tests for EventMonitor helpers
The system SHALL have Go unit tests for EventMonitor helper functions (topicToBigInt, topicToAddress, decodeAmount, resolveEscrowID) and handleEvent for all 6 event types.

#### Scenario: Monitor helper and event tests pass
- **WHEN** `go test ./internal/economy/escrow/hub/...` is run
- **THEN** all monitor tests pass including helper functions, resolveEscrowID with various store states, handleEvent for each event type, and processLog edge cases

### Requirement: Go unit tests for ABI parsing and types
The system SHALL have Go unit tests verifying ABI parsing functions return expected methods/events and OnChainDealStatus.String() returns correct values.

#### Scenario: ABI and type tests pass
- **WHEN** `go test ./internal/economy/escrow/hub/...` is run
- **THEN** ParseHubABI/ParseVaultABI/ParseFactoryABI return expected methods and events, and all 7 deal statuses + unknown map to correct strings

### Requirement: Anvil integration tests for full E2E flows
The system SHALL have integration tests (build tag `integration`) that deploy contracts to Anvil and test complete escrow lifecycles.

#### Scenario: Integration tests pass with running Anvil
- **WHEN** Anvil is running on localhost:8545 and `go test -tags integration ./internal/economy/escrow/hub/...` is run
- **THEN** all 7 integration tests pass: Hub full lifecycle, Hub dispute+resolve, Hub refund after deadline, Vault full lifecycle, Vault dispute+resolve, Factory multiple vaults, and Monitor event detection
