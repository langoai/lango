## ADDED Requirements

### Requirement: Knowledge runtime dependency guards stay actionable

The knowledge-runtime service SHALL preserve actionable fail-closed behavior when its backing receipt store is unavailable.

#### Scenario: Missing knowledgeruntime receipt store fails closed when opening a transaction
- **WHEN** `knowledgeruntime.Service.OpenTransaction` runs without a configured receipt store
- **THEN** the call SHALL return an error identifying the `open transaction` action
- **AND** SHALL preserve the actionable `knowledge runtime receipt store is required` cause instead of panicking

#### Scenario: Missing knowledgeruntime receipt store fails closed when selecting an execution path
- **WHEN** `knowledgeruntime.Service.SelectExecutionPath` runs without a configured receipt store
- **THEN** the call SHALL return an error identifying the `select execution path` action
- **AND** SHALL preserve the actionable `knowledge runtime receipt store is required` cause instead of panicking
