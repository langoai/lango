## Purpose

Capability spec for repository-facing docs references and layout. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: Repository docs references describe the Zensical docs toolchain
The README.md, docs/architecture/project-structure.md, and docs/development/build-test.md SHALL describe Zensical as the canonical docs toolchain and reference `zensical.toml` and `.venv/bin/zensical build` instead of MkDocs as the default docs path.

#### Scenario: README and architecture docs reference Zensical
- **WHEN** a user reads README.md and docs/architecture/project-structure.md
- **THEN** they SHALL see Zensical-native docs tooling references instead of `mkdocs.yml` as the canonical site definition

#### Scenario: Build-test docs reference the Zensical build path
- **WHEN** a user reads docs/development/build-test.md
- **THEN** the docs build instructions SHALL use `.venv/bin/zensical build`

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

### Requirement: Build and installation docs describe FTS5 as always on
Project documentation MUST describe FTS5 as included in the default runtime and MUST NOT require `-tags "fts5"` for normal builds or installs.

#### Scenario: Install docs use default build commands
- **WHEN** a user reads installation or development build instructions
- **THEN** normal build and install examples omit `-tags "fts5"`
- **AND** optional `vec` examples remain explicitly tagged

### Requirement: Architecture landing page links the identity trust reputation audit
The `docs/architecture/index.md` page SHALL include a quick link to `identity-trust-reputation-audit.md` and a short summary that frames it as the audit ledger for identity continuity, trust entry, reputation, and revocation in `knowledge exchange v1`.

#### Scenario: Identity audit appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Identity Trust Reputation Audit entry linking to `identity-trust-reputation-audit.md`
- **AND** the entry SHALL describe the audit ledger in terms of identity continuity, trust entry, reputation, and revocation in `knowledge exchange v1`

### Requirement: Architecture landing page links the pricing negotiation settlement audit
The `docs/architecture/index.md` page SHALL include a quick link to `pricing-negotiation-settlement-audit.md` and a short summary that frames it as the audit ledger for pricing, negotiation, settlement, and escrow in `knowledge exchange v1`.

#### Scenario: Pricing audit appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Pricing Negotiation Settlement Audit entry linking to `pricing-negotiation-settlement-audit.md`
- **AND** the entry SHALL describe the audit ledger in terms of pricing, negotiation, settlement, and escrow in `knowledge exchange v1`

### Requirement: Architecture landing page links the knowledge exchange runtime design slice
The `docs/architecture/index.md` page SHALL include a quick link to `knowledge-exchange-runtime.md` and a short summary that frames it as the first transaction-oriented runtime control-plane design slice for `knowledge exchange v1`, centered on transaction receipt and submission receipt with explicit current limits.

#### Scenario: Knowledge exchange runtime appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Knowledge Exchange Runtime entry linking to `knowledge-exchange-runtime.md`
- **AND** the entry SHALL describe the design slice as transaction-oriented and bounded by current limits

### Requirement: Architecture landing page links the settlement progression slice
The `docs/architecture/index.md` page SHALL include a quick link to `settlement-progression.md` and a short summary that frames it as the first transaction-level settlement progression slice for `knowledge exchange v1`.

#### Scenario: Settlement progression appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Settlement Progression entry linking to `settlement-progression.md`
- **AND** the entry SHALL describe the slice as transaction-level settlement progression bounded by current implementation limits

### Requirement: Architecture landing page links the actual settlement execution slice
The `docs/architecture/index.md` page SHALL include a quick link to `actual-settlement-execution.md` and a short summary that frames it as the first direct settlement execution slice for `knowledge exchange v1`.

#### Scenario: Actual settlement execution appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Actual Settlement Execution entry linking to `actual-settlement-execution.md`
- **AND** the entry SHALL describe the slice as direct settlement execution bounded by current implementation limits

### Requirement: Architecture landing page links the partial settlement execution slice
The `docs/architecture/index.md` page SHALL include a quick link to `partial-settlement-execution.md` and a short summary that frames it as the first direct partial settlement execution slice for `knowledge exchange v1`.

#### Scenario: Partial settlement execution appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Partial Settlement Execution entry linking to `partial-settlement-execution.md`
- **AND** the entry SHALL describe the slice as direct partial settlement execution bounded by current implementation limits

### Requirement: Architecture landing page links the escrow release slice
The `docs/architecture/index.md` page SHALL include a quick link to `escrow-release.md` and a short summary that frames it as the first funded-escrow release slice for `knowledge exchange v1`.

#### Scenario: Escrow release appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Escrow Release entry linking to `escrow-release.md`
- **AND** the entry SHALL describe the slice as funded-escrow release bounded by current implementation limits
