## MODIFIED Requirements

### Field Types
The package SHALL define the following input types:
- `InputText` — Free-text input
- `InputInt` — Integer input
- `InputBool` — Boolean toggle (spacebar)
- `InputSelect` — Cycle through options (left/right arrows)
- `InputPassword` — Masked text input
- `InputReadOnly` — Non-editable informational field

#### Scenario: Standard field types are available
- **WHEN** callers build forms with shared input components
- **THEN** the package SHALL expose `InputText`, `InputInt`, `InputBool`, `InputSelect`, `InputPassword`, and `InputReadOnly`

### Requirement: InputSearchSelect field type in form model
The FormModel MUST support InputSearchSelect as a field type with dedicated state management.

#### Scenario: Context-dependent help bar
- **WHEN** a dropdown is open
- **THEN** help bar shows dropdown-specific keys (↑↓ Navigate, Enter Select, Esc Close, Type Filter)
- **WHEN** no dropdown is open and the focused field is editable
- **THEN** help bar shows form-level keys including Enter Search
- **WHEN** no dropdown is open and the focused field is `InputReadOnly`
- **THEN** help bar shows read-only informational guidance instead of edit/search controls

## ADDED Requirements

### Requirement: InputReadOnly field type in form model
The FormModel MUST support InputReadOnly as a non-editable field type for status and compatibility notices.

#### Scenario: Read-only field does not mutate
- **WHEN** the focused field is `InputReadOnly`
- **THEN** text input, space, arrow, and enter keys SHALL NOT change the field value
- **AND** the field SHALL NOT be marked edited

#### Scenario: Read-only field renders as informational text
- **WHEN** the form View renders an `InputReadOnly` field
- **THEN** the field value SHALL be shown as informational text
- **AND** focused read-only fields SHALL render a read-only help footer
