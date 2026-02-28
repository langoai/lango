# CLI TUI Core Spec

## Goal
The `tuicore` package provides shared TUI form components used by both the onboard wizard and the settings editor.

## Requirements

### Field Types
The package SHALL define the following input types:
- `InputText` — Free-text input
- `InputInt` — Integer input
- `InputBool` — Boolean toggle (spacebar)
- `InputSelect` — Cycle through options (left/right arrows)
- `InputPassword` — Masked text input

### Field Struct
Each field SHALL have:
- Key, Label, Type, Value, Placeholder, Options, Checked, Width, Validate
- Exported `TextInput` field (bubbletea textinput.Model) for cross-package access

### FormModel
The form model SHALL:
- Manage a list of fields with cursor navigation (up/down, tab/shift-tab)
- Support text input, boolean toggle, and select cycling
- Render with title, field labels, and help footer
- Call OnCancel on Esc

### FormModel cursor navigation
The form cursor SHALL index into `VisibleFields()` instead of the full `Fields` slice. After any input event (including bool toggles that may change visibility), the cursor SHALL be clamped to `[0, len(visible)-1]`.

#### Scenario: Cursor clamp after visibility change
- **WHEN** the user is on the last visible field and toggles a bool that hides fields below
- **THEN** the cursor SHALL be clamped so it does not exceed the new visible field count

#### Scenario: Cursor re-evaluated after toggle
- **WHEN** the user toggles a bool field (space key)
- **THEN** the form SHALL re-evaluate `VisibleFields()` and clamp the cursor before processing further input

### FormModel View renders description
The form View SHALL render the `Description` of the currently focused field below that field's input widget, styled with `tui.FieldDescStyle`.

#### Scenario: Focused field description displayed
- **WHEN** the form View is rendered and field at cursor has a non-empty Description
- **THEN** the view SHALL include a line with the description text below that field

#### Scenario: No description for unfocused fields
- **WHEN** a field is not focused
- **THEN** its Description SHALL not be rendered in the View output

### ConfigState
The config state SHALL:
- Hold current `*config.Config` and dirty field tracking
- Provide `UpdateConfigFromForm`, `UpdateProviderFromForm`, `UpdateAuthProviderFromForm` methods
- Map all field keys to their corresponding config paths

### Skill field mappings in UpdateConfigFromForm
The `UpdateConfigFromForm` method SHALL map the following field keys to config paths:
- `skill_enabled` → `config.Skill.Enabled` (boolean)
- `skill_dir` → `config.Skill.SkillsDir` (string)

#### Scenario: Apply skill form values
- **WHEN** a form containing `skill_enabled` and `skill_dir` fields is processed by `UpdateConfigFromForm`
- **THEN** the values SHALL be written to `config.Skill.Enabled` and `config.Skill.SkillsDir` respectively

### Cron field mappings in UpdateConfigFromForm
The `UpdateConfigFromForm` method SHALL map the following field keys to config paths:
- `cron_enabled` → `config.Cron.Enabled` (boolean)
- `cron_timezone` → `config.Cron.Timezone` (string)
- `cron_max_jobs` → `config.Cron.MaxConcurrentJobs` (integer)
- `cron_session_mode` → `config.Cron.DefaultSessionMode` (string)
- `cron_history_retention` → `config.Cron.HistoryRetention` (string)

#### Scenario: Apply cron form values
- **WHEN** a form containing cron fields is processed by `UpdateConfigFromForm`
- **THEN** the values SHALL be written to the corresponding `config.Cron` fields

### Background field mappings in UpdateConfigFromForm
The `UpdateConfigFromForm` method SHALL map the following field keys to config paths:
- `bg_enabled` → `config.Background.Enabled` (boolean)
- `bg_yield_ms` → `config.Background.YieldMs` (integer)
- `bg_max_tasks` → `config.Background.MaxConcurrentTasks` (integer)

#### Scenario: Apply background form values
- **WHEN** a form containing background fields is processed by `UpdateConfigFromForm`
- **THEN** the values SHALL be written to the corresponding `config.Background` fields

### Workflow field mappings in UpdateConfigFromForm
The `UpdateConfigFromForm` method SHALL map the following field keys to config paths:
- `wf_enabled` → `config.Workflow.Enabled` (boolean)
- `wf_max_steps` → `config.Workflow.MaxConcurrentSteps` (integer)
- `wf_timeout` → `config.Workflow.DefaultTimeout` (duration parsed from string)
- `wf_state_dir` → `config.Workflow.StateDir` (string)

#### Scenario: Apply workflow form values
- **WHEN** a form containing workflow fields is processed by `UpdateConfigFromForm`
- **THEN** the values SHALL be written to the corresponding `config.Workflow` fields

### Field Description property
The `Field` struct SHALL include a `Description string` property for inline help text.

#### Scenario: Description stored on field
- **WHEN** a Field is created with a Description value
- **THEN** the Description SHALL be accessible on the field instance

### VisibleWhen conditional visibility
The `Field` struct SHALL include a `VisibleWhen func() bool` property. When non-nil, the field is shown only when the function returns true. When nil, the field is always visible.

#### Scenario: VisibleWhen nil means always visible
- **WHEN** a Field has `VisibleWhen` set to nil
- **THEN** `IsVisible()` SHALL return true

#### Scenario: VisibleWhen returns false hides field
- **WHEN** a Field has `VisibleWhen` returning false
- **THEN** `IsVisible()` SHALL return false and the field SHALL not appear in `VisibleFields()`

#### Scenario: VisibleWhen dynamically responds to state
- **WHEN** a VisibleWhen closure captures a pointer to a parent field's Checked state
- **THEN** toggling the parent field SHALL immediately affect the child field's visibility on next `VisibleFields()` call

### IsVisible method on Field
The `Field` struct SHALL expose an `IsVisible() bool` method that returns true when `VisibleWhen` is nil, and the result of `VisibleWhen()` otherwise.

### VisibleFields on FormModel
`FormModel` SHALL expose a `VisibleFields() []*Field` method that returns only fields where `IsVisible()` returns true.

#### Scenario: VisibleFields filters hidden fields
- **WHEN** a form has 5 fields and 2 have VisibleWhen returning false
- **THEN** VisibleFields() SHALL return 3 fields

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

### Embedding ProviderID deprecation in state update
The `UpdateConfigFromForm` case for `emb_provider_id` SHALL set `cfg.Embedding.Provider` to the value AND clear `cfg.Embedding.ProviderID` to empty string.

#### Scenario: emb_provider_id clears deprecated field
- **WHEN** UpdateConfigFromForm processes key "emb_provider_id" with value "openai"
- **THEN** `cfg.Embedding.Provider` SHALL be "openai" AND `cfg.Embedding.ProviderID` SHALL be ""
