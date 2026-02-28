## ADDED Requirements

### Requirement: New packages documented in architecture
The README.md Architecture section and docs/architecture/project-structure.md SHALL include dbmigrate, lifecycle, keyring, and sandbox packages.

#### Scenario: README architecture tree includes new packages
- **WHEN** a user reads README.md Architecture section
- **THEN** dbmigrate, lifecycle, keyring, sandbox, and cli/p2p packages SHALL appear in the tree

#### Scenario: project-structure.md Infrastructure table includes new packages
- **WHEN** a user reads docs/architecture/project-structure.md Infrastructure section
- **THEN** lifecycle, keyring, sandbox, and dbmigrate packages SHALL have entries with descriptions

### Requirement: Security package description updated
The docs/architecture/project-structure.md security package description SHALL mention KMS providers.

#### Scenario: security row mentions KMS
- **WHEN** a user reads the security row in project-structure.md
- **THEN** the description SHALL include KMS providers (AWS, GCP, Azure, PKCS#11)

### Requirement: Skills description corrected
The README.md and docs/architecture/project-structure.md SHALL NOT reference "30" or "38" embedded default skills, and SHALL explain that built-in skills were removed due to the passphrase security model.

#### Scenario: README skills line is accurate
- **WHEN** a user reads the README.md Architecture section skills line
- **THEN** it SHALL describe the skill system as a scaffold with an explanation of why built-in skills were removed

#### Scenario: project-structure.md skills section is accurate
- **WHEN** a user reads the skills section of project-structure.md
- **THEN** it SHALL explain that ~30 built-in skills were removed and the infrastructure remains functional for user-defined skills

### Requirement: Security feature card updated in docs landing page
The docs/index.md Security card SHALL mention hardware keyring, SQLCipher, and Cloud KMS.

#### Scenario: docs/index.md Security card is complete
- **WHEN** a user reads the Security card on docs/index.md
- **THEN** it SHALL mention hardware keyring (Touch ID / TPM), SQLCipher database encryption, and Cloud KMS integration

### Requirement: README Features security line updated
The README.md Features section security line SHALL mention hardware keyring, SQLCipher, and Cloud KMS.

#### Scenario: README security feature is complete
- **WHEN** a user reads the Features section of README.md
- **THEN** the Secure line SHALL include hardware keyring, SQLCipher DB encryption, and Cloud KMS
