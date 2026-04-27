## ADDED Requirements

### Requirement: Active DID exposure reflects the currently available identity mode
Operator-facing identity surfaces SHALL expose the active DID when one is available, regardless of whether the runtime is using a legacy wallet-derived `did:lango:<hex>` identity or a bundle-backed `did:lango:v2:<hash>` identity.

#### Scenario: Bundle-backed identity is active
- **WHEN** the runtime has an active bundle-backed identity
- **THEN** operator-facing identity surfaces SHALL expose the `did:lango:v2:<hash>` DID

#### Scenario: Legacy identity fallback is active
- **WHEN** no bundle-backed identity is active and a legacy wallet-derived identity is available
- **THEN** operator-facing identity surfaces SHALL expose the legacy `did:lango:<hex>` DID

### Requirement: Identity lookup remains read-only
Any operator-facing identity lookup path SHALL remain read-only. Querying local identity SHALL NOT create, persist, or rotate identity material just to render a DID.

#### Scenario: Bundle-backed DID lookup
- **WHEN** an operator-facing surface reads the local DID from persisted identity state
- **THEN** the lookup SHALL NOT create or overwrite an identity bundle

#### Scenario: Legacy DID fallback
- **WHEN** an operator-facing surface falls back to wallet public-key DID derivation
- **THEN** the lookup SHALL derive the DID from existing key material without mutating identity state
