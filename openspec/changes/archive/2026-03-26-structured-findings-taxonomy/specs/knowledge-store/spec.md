## MODIFIED Requirements

### Requirement: Knowledge CRUD Operations
The system SHALL provide persistent CRUD operations for knowledge entries identified by key. Knowledge entries SHALL be versioned: each save appends a new version instead of updating in place. All read operations SHALL default to the latest version (`is_latest=true`). When the latest version has the same `(category, content)` as the new entry, SaveKnowledge SHALL be a no-op (no new version created). Changes to `source`, `tags`, or temporal hints alone do NOT justify a new version.

#### Scenario: Save new knowledge entry
- **WHEN** `SaveKnowledge` is called with a key that does not exist
- **THEN** the system SHALL create a new knowledge entry with `version=1`, `is_latest=true`, and the given key, category, content, tags, and source

#### Scenario: Save existing knowledge entry (append version)
- **WHEN** `SaveKnowledge` is called with a key that already exists (latest version N) and the content or category differs
- **THEN** the system SHALL atomically set the existing latest row's `is_latest` to `false` and create a new entry with `version=N+1`, `is_latest=true`
- **AND** the new version SHALL carry forward `use_count` and `relevance_score` from the previous version

#### Scenario: Content-dedup no-op
- **WHEN** `SaveKnowledge` is called with a key whose latest version has the same `(category, content)`
- **THEN** the system SHALL return nil without creating a new version

#### Scenario: Get knowledge by key
- **WHEN** `GetKnowledge` is called with an existing key
- **THEN** the system SHALL return the latest version (`is_latest=true`) of the knowledge entry with `Version` and `CreatedAt` populated
- **AND** if no latest entry exists for the key, SHALL return an error

#### Scenario: Delete knowledge by key
- **WHEN** `DeleteKnowledge` is called with an existing key
- **THEN** the system SHALL remove ALL versions of the entry from the store

#### Scenario: Increment knowledge use count
- **WHEN** `IncrementKnowledgeUseCount` is called with a valid key
- **THEN** the system SHALL increment the use count by 1 on the latest version only (`is_latest=true`)
