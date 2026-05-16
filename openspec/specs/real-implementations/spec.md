# Spec: Stub to Real Implementations

## Purpose

Capability spec for real-implementations. See requirements below for scope and behavior contracts.

## Requirements

### REQ-1: CLI commands must call real services
All smart account CLI commands (deploy, info, session create/list/revoke, policy show/set, module list/install, paymaster status/approve) SHALL initialize dependencies from bootstrap and call actual service methods.

#### Scenario: Account deploy uses the real manager
- **WHEN** `lango account deploy` is executed
- **THEN** `manager.GetOrDeploy()` SHALL be called
- **AND** the real account address SHALL be displayed

#### Scenario: Session create returns a real key ID
- **WHEN** `lango account session create` is executed with valid flags
- **THEN** a real session key SHALL be created
- **AND** the resulting key ID SHALL be returned

### REQ-2: PolicySyncer bridges Go and on-chain policies
A `PolicySyncer` SHALL support:
- `PushToChain`: Write Go-side policy limits to the SpendingHook contract
- `PullFromChain`: Read on-chain config and update the Go-side policy
- `DetectDrift`: Compare and report differences between Go and on-chain policies

#### Scenario: PolicySyncer exposes push, pull, and drift detection
- **WHEN** the policy syncer is used to reconcile Go-side and on-chain policy state
- **THEN** it SHALL support push-to-chain, pull-from-chain, and drift-detection operations

### REQ-3: Paymaster recovery with retry and fallback
A `RecoverableProvider` SHALL wrap any `PaymasterProvider` with:
- Exponential-backoff retry for transient errors (`ErrPaymasterTimeout`)
- Immediate failure for permanent errors (`ErrPaymasterRejected`, `ErrInsufficientToken`)
- Configurable fallback: abort or switch to direct gas
- `IsTransient()`/`IsPermanent()` error classification functions

#### Scenario: RecoverableProvider differentiates transient and permanent failures
- **WHEN** a wrapped paymaster request fails
- **THEN** transient failures SHALL use retry and fallback policy
- **AND** permanent failures SHALL fail immediately without retry
