## ADDED Requirements

### Requirement: Experimental badge on settings menu categories
The settings menu SHALL display an `[EXP]` badge next to categories that correspond to experimental features. The badge SHALL use `BadgeExperimentalStyle` and appear after the `[ADV]` badge and before the dependency warning badge.

#### Scenario: Experimental category shows EXP badge
- **WHEN** a category ID is present in the `ExperimentalCategories` map
- **THEN** the menu item SHALL render with an `[EXP]` badge

#### Scenario: Non-experimental category has no EXP badge
- **WHEN** a category ID is not in the `ExperimentalCategories` map
- **THEN** no `[EXP]` badge SHALL appear

### Requirement: Experimental search filter
The settings menu search SHALL support an `@experimental` filter that shows only categories marked as experimental.

#### Scenario: Filter by experimental
- **WHEN** user types `@experimental` in the search bar
- **THEN** only categories present in `ExperimentalCategories` SHALL be shown

### Requirement: ExperimentalCategories drift prevention
A test SHALL verify that the `ExperimentalCategories` map contains exactly the expected set of category IDs. This prevents silent drift when categories are added or removed.

#### Scenario: Category added without map update
- **WHEN** a new experimental category is added to the menu but not to `ExperimentalCategories`
- **THEN** the drift test SHALL fail
