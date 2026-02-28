## MODIFIED Requirements

### Requirement: InputSearchSelect field type in form model
The FormModel MUST support InputSearchSelect as a field type with dedicated state management.

#### Scenario: Field initialization
- **WHEN** AddField is called with InputSearchSelect type
- **THEN** TextInput is initialized with search placeholder, FilteredOptions copies Options

#### Scenario: HasOpenDropdown query
- **WHEN** any field has SelectOpen == true
- **THEN** HasOpenDropdown() returns true

#### Scenario: Context-dependent help bar
- **WHEN** a dropdown is open
- **THEN** help bar shows dropdown-specific keys (↑↓ Navigate, Enter Select, Esc Close, Type Filter)
- **WHEN** no dropdown is open
- **THEN** help bar shows form-level keys including Enter Search
