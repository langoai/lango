## MODIFIED Requirements

### Requirement: Repository test-harness guards stay executable

Repository-level test-harness regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review. Security-sensitive public wrappers that expose stdio-backed behavior SHALL have focused tests proving they can be exercised through injected seams instead of process-global stdio replacement.

#### Scenario: Passphrase public wrapper stdio seams stay covered

- **WHEN** passphrase wrapper tests run
- **THEN** they SHALL fail if `Acquire` stops using injected stdin, stderr, or terminal-detection seams
- **AND** they SHALL fail if `AcquireNonInteractive` stops using the injected stderr seam
