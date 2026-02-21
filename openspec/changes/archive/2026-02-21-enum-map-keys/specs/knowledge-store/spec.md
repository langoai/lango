## MODIFIED Requirements

### Requirement: LearningStats ByCategory field type
The `LearningStats.ByCategory` field SHALL use `map[entlearning.Category]int` instead of `map[string]int`. The map SHALL be populated using the `Category` field directly without string casting.

#### Scenario: ByCategory uses enum type as key
- **WHEN** `GetLearningStats` is called and learning entries exist
- **THEN** `ByCategory` map keys SHALL be of type `entlearning.Category`, not `string`

#### Scenario: JSON serialization compatibility
- **WHEN** `LearningStats` is serialized to JSON
- **THEN** the `by_category` field SHALL produce identical JSON output as the previous `map[string]int` representation
