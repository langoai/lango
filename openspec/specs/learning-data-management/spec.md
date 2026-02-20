## ADDED Requirements

### Requirement: Learning Statistics
The system SHALL provide aggregate statistics about stored learning entries via `Store.GetLearningStats()`.

#### Scenario: Return complete statistics
- **WHEN** `GetLearningStats` is called
- **THEN** the system SHALL return a `LearningStats` struct containing: total count, category distribution (map[string]int), average confidence, oldest entry time, newest entry time, total occurrences, and total successes

#### Scenario: Empty store returns zero stats
- **WHEN** `GetLearningStats` is called and no learning entries exist
- **THEN** the system SHALL return a `LearningStats` with TotalCount=0, empty ByCategory map, and zero-value times

### Requirement: Learning Listing with Filters
The system SHALL provide filtered, paginated listing of learning entries via `Store.ListLearnings()`.

#### Scenario: Filter by category
- **WHEN** `ListLearnings` is called with a non-empty category
- **THEN** the system SHALL return only learning entries matching that category

#### Scenario: Filter by minimum confidence
- **WHEN** `ListLearnings` is called with minConfidence > 0
- **THEN** the system SHALL return only entries with confidence >= minConfidence

#### Scenario: Filter by age
- **WHEN** `ListLearnings` is called with a non-zero olderThan time
- **THEN** the system SHALL return only entries created before that time

#### Scenario: Pagination
- **WHEN** `ListLearnings` is called with limit and offset
- **THEN** the system SHALL return at most `limit` entries starting from `offset`, along with the total count matching the filters

### Requirement: Single Learning Deletion
The system SHALL support deleting a single learning entry by UUID via `Store.DeleteLearning()`.

#### Scenario: Delete existing entry
- **WHEN** `DeleteLearning` is called with a valid UUID
- **THEN** the system SHALL remove that learning entry from the store

#### Scenario: Delete non-existent entry
- **WHEN** `DeleteLearning` is called with a UUID that does not exist
- **THEN** the system SHALL return an error

### Requirement: Bulk Learning Deletion by Criteria
The system SHALL support bulk deletion of learning entries by criteria via `Store.DeleteLearningsWhere()`.

#### Scenario: Delete by category
- **WHEN** `DeleteLearningsWhere` is called with a non-empty category
- **THEN** the system SHALL delete only entries matching that category and return the count of deleted entries

#### Scenario: Delete by maximum confidence
- **WHEN** `DeleteLearningsWhere` is called with maxConfidence > 0
- **THEN** the system SHALL delete only entries with confidence <= maxConfidence

#### Scenario: Delete by age
- **WHEN** `DeleteLearningsWhere` is called with a non-zero olderThan time
- **THEN** the system SHALL delete only entries created before that time

#### Scenario: Combined criteria
- **WHEN** `DeleteLearningsWhere` is called with multiple non-zero criteria
- **THEN** the system SHALL apply all criteria with AND logic

### Requirement: Learning Stats Agent Tool
The system SHALL provide a `learning_stats` agent tool that returns learning statistics as JSON.

#### Scenario: Invoke learning_stats tool
- **WHEN** the agent invokes the `learning_stats` tool with no parameters
- **THEN** the system SHALL call `Store.GetLearningStats()` and return the result as formatted JSON
- **AND** the tool SHALL have safety level "safe"

### Requirement: Learning Cleanup Agent Tool
The system SHALL provide a `learning_cleanup` agent tool that deletes learning entries by criteria.

#### Scenario: Dry run (default)
- **WHEN** the agent invokes `learning_cleanup` with `dry_run=true` (or omitted, as default is true)
- **THEN** the system SHALL return the count of entries matching the criteria without actually deleting them

#### Scenario: Execute cleanup
- **WHEN** the agent invokes `learning_cleanup` with `dry_run=false`
- **THEN** the system SHALL delete entries matching the criteria and return the count of deleted entries

#### Scenario: Delete by ID
- **WHEN** the agent invokes `learning_cleanup` with an `id` parameter (UUID string)
- **THEN** the system SHALL delete only the single entry with that UUID

#### Scenario: Delete by criteria
- **WHEN** the agent invokes `learning_cleanup` with category, max_confidence, and/or older_than_days
- **THEN** the system SHALL apply all provided criteria with AND logic
- **AND** the tool SHALL have safety level "moderate"
