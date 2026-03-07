## ADDED Requirements

### Requirement: Economy commands in CLI reference
The docs/cli/index.md SHALL include an Economy section with a table listing all 5 economy commands: `lango economy budget status`, `lango economy risk status`, `lango economy pricing status`, `lango economy negotiate status`, and `lango economy escrow status`.

#### Scenario: Economy table exists in CLI index
- **WHEN** a user reads docs/cli/index.md
- **THEN** an "Economy" section SHALL appear with 5 command entries after the P2P Network section

### Requirement: Contract commands in CLI reference
The docs/cli/index.md SHALL include a Contract section with a table listing all 3 contract commands: `lango contract read`, `lango contract call`, and `lango contract abi load`.

#### Scenario: Contract table exists in CLI index
- **WHEN** a user reads docs/cli/index.md
- **THEN** a "Contract" section SHALL appear with 3 command entries after the Economy section

### Requirement: Metrics commands in CLI reference
The docs/cli/index.md SHALL include a Metrics section with a table listing all 5 metrics commands: `lango metrics`, `lango metrics sessions`, `lango metrics tools`, `lango metrics agents`, and `lango metrics history`.

#### Scenario: Metrics table exists in CLI index
- **WHEN** a user reads docs/cli/index.md
- **THEN** a "Metrics" section SHALL appear with 5 command entries after the Contract section
