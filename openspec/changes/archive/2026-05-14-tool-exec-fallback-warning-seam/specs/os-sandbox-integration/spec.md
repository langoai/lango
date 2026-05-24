## MODIFIED Requirements

### Requirement: Exec tool sandbox integration
The exec tool SHALL apply `OSIsolator` to all 3 `exec.Command` call sites (`Run`, `RunWithPTY`, `StartBackground`) after command creation and before process start.

#### Scenario: Fail-open warning path is deterministic under test
- **WHEN** the exec tool's fail-open warning path is exercised in tests
- **THEN** the stderr writer SHALL be injectable so the one-shot warning contract can be asserted without intercepting process-global stderr
