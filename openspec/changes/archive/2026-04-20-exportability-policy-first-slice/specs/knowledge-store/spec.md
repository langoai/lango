## MODIFIED Requirements

### Requirement: Knowledge CRUD Operations
The system SHALL provide persistent CRUD operations for knowledge entries identified by key. Knowledge entries SHALL be versioned: each save appends a new version instead of updating in place. All read operations SHALL default to the latest version (`is_latest=true`). When the latest version has the same `(category, content, source, source_class, asset_label)` as the new entry, SaveKnowledge SHALL be a no-op. Changes to `source_class` or `asset_label` SHALL be treated as version-significant because they change exportability semantics.

#### Scenario: Save new knowledge entry
- **WHEN** `SaveKnowledge` is called with a key that does not exist
- **THEN** the system SHALL create a new knowledge entry with `version=1`, `is_latest=true`, and the given key, category, content, tags, source, source class, and asset label

#### Scenario: Save existing knowledge entry (append version)
- **WHEN** `SaveKnowledge` is called with a key that already exists (latest version N) and the content, category, source, source class, or asset label differs
- **THEN** the system SHALL atomically set the existing latest row's `is_latest` to `false` and create a new entry with `version=N+1`, `is_latest=true`
- **AND** the new version SHALL carry forward `use_count` and `relevance_score` from the previous version

#### Scenario: Content and exportability metadata dedup no-op
- **WHEN** `SaveKnowledge` is called with a key whose latest version has the same `(category, content, source, source_class, asset_label)`
- **THEN** the system SHALL return nil without creating a new version

#### Scenario: Get knowledge by key
- **WHEN** `GetKnowledge` is called with an existing key
- **THEN** the system SHALL return the latest version (`is_latest=true`) of the knowledge entry with `Version`, `CreatedAt`, `SourceClass`, and `AssetLabel` populated
- **AND** if no latest entry exists for the key, SHALL return an error
