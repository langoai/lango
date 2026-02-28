## ADDED Requirements

### Requirement: Security extension commands documented in CLI reference
The docs/cli/index.md SHALL include keyring (store/clear/status), db-migrate, db-decrypt, and kms (status/test/keys) commands in the Security table.

#### Scenario: Security table contains all 13 commands
- **WHEN** a user reads docs/cli/index.md Security section
- **THEN** the table SHALL list 13 security commands including the 8 new extension commands

### Requirement: P2P Network section in CLI reference
The docs/cli/index.md SHALL include a P2P Network table with all 17 P2P commands (status, peers, connect, disconnect, firewall, discover, identity, reputation, pricing, session, sandbox).

#### Scenario: P2P table exists between Payment and Automation
- **WHEN** a user reads docs/cli/index.md
- **THEN** a "P2P Network" section SHALL appear with 17 command entries

### Requirement: Background task commands in CLI reference
The docs/cli/index.md Automation section SHALL include bg list, bg status, bg cancel, and bg result commands.

#### Scenario: bg commands appear in Automation table
- **WHEN** a user reads the Automation section of docs/cli/index.md
- **THEN** 4 bg commands SHALL be listed after the workflow commands

### Requirement: README CLI section includes all commands
The README.md CLI Commands section SHALL include security keyring/db/kms commands, p2p session/sandbox commands, and bg commands.

#### Scenario: README CLI section is complete
- **WHEN** a user reads README.md CLI Commands section
- **THEN** all security extension, p2p session/sandbox, and bg commands SHALL be listed
