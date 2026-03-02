## MODIFIED Requirements

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
