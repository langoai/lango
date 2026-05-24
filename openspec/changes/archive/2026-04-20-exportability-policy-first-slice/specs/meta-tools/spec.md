## ADDED Requirements

### Requirement: Exportability evaluation tool
The meta tools surface SHALL provide an `evaluate_exportability` tool that evaluates an artifact label and a list of knowledge source keys, then returns the exportability decision receipt.

#### Scenario: Tool returns receipt payload
- **WHEN** `evaluate_exportability` is invoked with `artifact_label`, `source_keys`, and `stage`
- **THEN** it SHALL load the latest knowledge entries for the provided keys
- **AND** it SHALL evaluate them through the exportability policy engine
- **AND** it SHALL return the decision stage, decision state, policy code, explanation, and lineage summary

#### Scenario: Tool emits audit-backed exportability receipt
- **WHEN** `evaluate_exportability` completes
- **THEN** it SHALL append an `exportability_decision` audit entry recording the receipt payload

## MODIFIED Requirements

### Requirement: Save knowledge meta tool
The `save_knowledge` tool SHALL accept exportability-related source tagging metadata in addition to key, category, content, tags, and source.

#### Scenario: Save knowledge with source tagging
- **WHEN** `save_knowledge` is called with `source_class` and `asset_label`
- **THEN** the stored knowledge entry SHALL persist those fields for later exportability evaluation

#### Scenario: Save knowledge default source class
- **WHEN** `save_knowledge` is called without `source_class`
- **THEN** the tool SHALL default the stored source class to `private-confidential`
