## Purpose

Phase-based bootstrap with sequential execution and reverse-order cleanup on failure.

## Requirements

### Requirement: Phase-based pipeline
The bootstrap system SHALL execute phases sequentially using a Pipeline with Phase structs containing Name, Run, and optional Cleanup functions.

#### Scenario: All phases succeed
- **WHEN** all phases complete without error
- **THEN** Pipeline.Execute SHALL return the Result from State

### Requirement: Reverse cleanup on failure
If a phase fails, the Pipeline SHALL call Cleanup functions of all previously completed phases in reverse order.

#### Scenario: Phase 4 fails after phases 1-3 complete
- **WHEN** phase 4 returns an error
- **THEN** cleanup SHALL run for phases 3, 2, 1 in that order (if they have Cleanup functions)

### Requirement: State passes data between phases
The Pipeline SHALL use a State struct to carry data between phases, including Options, Result, and intermediate values.

#### Scenario: Database handle passes from open to security phase
- **WHEN** phaseOpenDatabase sets Client on State
- **THEN** phaseLoadSecurityState SHALL read Client from State

### Requirement: Default bootstrap phases
The system SHALL provide DefaultPhases() returning the 7-phase bootstrap sequence: ensureDataDir, detectEncryption, acquirePassphrase, openDatabase, loadSecurityState, initCrypto, loadProfile.

#### Scenario: Run uses default phases
- **WHEN** bootstrap.Run(opts) is called
- **THEN** it SHALL create a Pipeline with DefaultPhases and execute it

#### Scenario: Data directory writability verified
- **WHEN** phaseEnsureDataDir creates `~/.lango/`
- **THEN** it SHALL write a probe file (`.write-test`) to verify writability
- **AND** it SHALL remove the probe file immediately after verification
- **AND** if the directory is not writable, it SHALL return an error including the current UID

#### Scenario: Skills directory pre-created
- **WHEN** phaseEnsureDataDir completes successfully
- **THEN** `~/.lango/skills/` SHALL exist with the same permission mode as the parent data directory

#### Scenario: Consistent permission mode
- **WHEN** phaseEnsureDataDir or openDatabase create directories
- **THEN** they SHALL use the `dataDirPerm` constant (0700)
