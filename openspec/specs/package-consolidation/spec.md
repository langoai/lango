# Spec: Package Consolidation

## Overview
Merge three underused packages into their logical parent packages to improve codebase clarity.

## Purpose

Capability spec for package-consolidation. See requirements below for scope and behavior contracts.

## Requirements

### R1: ctxutil → types
The system SHALL keep `DetachContext()` and its supporting implementation in `internal/types/context.go` with tests in `internal/types/context_test.go`, and current importers SHALL use that path.

#### Scenario: Background task context detach behavior is preserved
- **WHEN** `types.DetachContext(ctx)` is used after the consolidation
- **THEN** it SHALL preserve `Value()` lookups
- **AND** it SHALL detach from parent cancellation

#### Scenario: types package remains cycle-free
- **WHEN** the consolidated `types` package is compiled
- **THEN** it SHALL not gain upstream import cycles from the moved context helpers

### R2: passphrase → security/passphrase
The system SHALL keep passphrase acquisition, keyfile, stdin, and interactive helpers in `internal/security/passphrase/` under the `passphrase` package name.

#### Scenario: Passphrase acquisition order is preserved
- **WHEN** passphrase acquisition runs after the package move
- **THEN** the priority order SHALL remain keyring → keyfile → interactive → stdin

#### Scenario: Keyfile helpers remain behaviorally unchanged
- **WHEN** read, write, shred, or permission validation helpers are called after consolidation
- **THEN** their observable behavior SHALL remain unchanged

### R3: zkp → p2p/zkp
The system SHALL keep ZKP implementation files under `internal/p2p/zkp/` (including `circuits/`) with package names `zkp` and `circuits`, and current importers SHALL use that path.

#### Scenario: ZKP proving and verifying behavior is preserved
- **WHEN** `ProverService` is used after the move to `internal/p2p/zkp/`
- **THEN** its proving and verification behavior SHALL remain unchanged

#### Scenario: All moved circuits still compile and run
- **WHEN** the ownership, attestation, capability, and balance circuits are compiled after consolidation
- **THEN** they SHALL behave identically to the pre-move versions

## Constraints
- Zero functional changes — only import paths change
- No import cycles introduced
- All existing tests must pass without modification
