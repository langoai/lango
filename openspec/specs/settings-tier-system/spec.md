## Purpose

Settings TUI tier system that reduces initial complexity by showing only essential (~14) categories by default, with Tab toggle to reveal all (~47) advanced categories.

## Requirements

### Requirement: Category tier filtering
Each settings Category SHALL have a Tier field (TierBasic=0, TierAdvanced=1). By default, only TierBasic categories are shown.

#### Scenario: Basic mode shows essential categories
- **WHEN** settings menu opens in default mode
- **THEN** only categories with Tier=TierBasic are displayed (~14 items)

#### Scenario: Advanced mode shows all categories
- **WHEN** user presses Tab to toggle advanced mode
- **THEN** all categories (Basic + Advanced) are displayed (~47 items)

### Requirement: Tab key toggles tier mode
The menu SHALL toggle between Basic and Advanced views when Tab is pressed in normal (non-search) mode. The help bar SHALL display "Tab: Show Advanced" or "Tab: Show Basic" accordingly.

#### Scenario: Toggle to advanced
- **WHEN** user presses Tab in basic mode
- **THEN** showAdvanced becomes true, all categories are visible, help bar shows "Show Basic"

#### Scenario: Toggle back to basic
- **WHEN** user presses Tab in advanced mode
- **THEN** showAdvanced becomes false, only basic categories visible, help bar shows "Show Advanced"

### Requirement: Search ignores tier filter
Search mode SHALL always search across ALL categories regardless of the current tier setting.

#### Scenario: Search finds advanced category in basic mode
- **WHEN** user is in basic mode and searches for "zkp"
- **THEN** P2P ZKP category appears in search results despite being TierAdvanced

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
