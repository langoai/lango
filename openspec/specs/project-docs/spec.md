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

### Requirement: Architecture landing page links the escrow refund slice
The `docs/architecture/index.md` page SHALL include a quick link to `escrow-refund.md` and a short summary that frames it as the first funded-escrow refund slice for `knowledge exchange v1`.

#### Scenario: Escrow refund appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Escrow Refund entry linking to `escrow-refund.md`
- **AND** the entry SHALL describe the slice as funded-escrow refund bounded by current implementation limits

### Requirement: Architecture landing page links the dispute hold slice
The `docs/architecture/index.md` page SHALL include a quick link to `dispute-hold.md` and a short summary that frames it as the first dispute-linked escrow hold slice for `knowledge exchange v1`.

#### Scenario: Dispute hold appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Dispute Hold entry linking to `dispute-hold.md`
- **AND** the entry SHALL describe the slice as funded dispute-ready escrow hold bounded by current implementation limits

### Requirement: Architecture landing page links the release-vs-refund adjudication slice
The `docs/architecture/index.md` page SHALL include a quick link to `release-vs-refund-adjudication.md` and a short summary that frames it as the first post-hold release-vs-refund adjudication slice for `knowledge exchange v1`.

#### Scenario: Release-vs-refund adjudication appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Release vs Refund Adjudication entry linking to `release-vs-refund-adjudication.md`
- **AND** the entry SHALL describe the slice as canonical post-hold branching bounded by current implementation limits

### Requirement: Architecture landing page links adjudication-aware release/refund execution gating
The `docs/architecture/index.md` page SHALL include a quick link to `adjudication-aware-release-refund-execution-gating.md` and a short summary that frames it as the first slice that connects canonical escrow adjudication to release/refund execution gating.

#### Scenario: Adjudication-aware execution gating appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Adjudication-Aware Release/Refund Execution Gating entry linking to `adjudication-aware-release-refund-execution-gating.md`
- **AND** the entry SHALL describe the slice as release/refund execution gating bounded by current implementation limits

### Requirement: Architecture landing page links automatic post-adjudication execution
The `docs/architecture/index.md` page SHALL include a quick link to `automatic-post-adjudication-execution.md` and a short summary that frames it as the first inline convenience slice after escrow adjudication.

#### Scenario: Automatic post-adjudication execution appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Automatic Post-Adjudication Execution entry linking to `automatic-post-adjudication-execution.md`
- **AND** the entry SHALL describe the slice as inline post-adjudication orchestration bounded by current implementation limits

### Requirement: Architecture landing page links background post-adjudication execution
The `docs/architecture/index.md` page SHALL include a quick link to `background-post-adjudication-execution.md` and a short summary that frames it as the first async convenience slice after escrow adjudication.

#### Scenario: Background post-adjudication execution appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Background Post-Adjudication Execution entry linking to `background-post-adjudication-execution.md`
- **AND** the entry SHALL describe the slice as async post-adjudication orchestration bounded by current implementation limits

### Requirement: Architecture landing page links retry / dead-letter handling
The `docs/architecture/index.md` page SHALL include a quick link to `retry-dead-letter-handling.md` and a short summary that frames it as the first retry / dead-letter slice for background post-adjudication execution.

#### Scenario: Retry / dead-letter handling appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Retry / Dead-Letter Handling entry linking to `retry-dead-letter-handling.md`
- **AND** the entry SHALL describe the slice as bounded retry and terminal dead-letter handling bounded by current implementation limits

### Requirement: Architecture landing page links operator replay / manual retry
The `docs/architecture/index.md` page SHALL include a quick link to `operator-replay-manual-retry.md` and a short summary that frames it as the first operator-facing replay slice for dead-lettered post-adjudication execution.

#### Scenario: Operator replay / manual retry appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Operator Replay / Manual Retry entry linking to `operator-replay-manual-retry.md`
- **AND** the entry SHALL describe the slice as operator replay bounded by current implementation limits

### Requirement: Architecture landing page links policy-driven replay controls
The `docs/architecture/index.md` page SHALL include a quick link to `policy-driven-replay-controls.md` and a short summary that frames it as the first authorization slice for replay.

#### Scenario: Policy-driven replay controls appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Policy-Driven Replay Controls entry linking to `policy-driven-replay-controls.md`
- **AND** the entry SHALL describe the slice as replay authorization bounded by current implementation limits

### Requirement: Architecture landing page links dead-letter browsing / status observation
The `docs/architecture/index.md` page SHALL include a quick link to `dead-letter-browsing-status-observation.md` and a short summary that frames it as the first read-only operator visibility slice for actor/time-aware post-adjudication dead-letter browsing.

#### Scenario: Dead-letter browsing / status observation appears in architecture landing page
- **WHEN** a user reads `docs/architecture/index.md`
- **THEN** they SHALL find the Dead-Letter Browsing / Status Observation entry linking to `dead-letter-browsing-status-observation.md`
- **AND** the entry SHALL describe the slice as read-only actor/time-aware dead-letter visibility bounded by current implementation limits
