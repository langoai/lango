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

### ConfigState
The config state SHALL:
- Hold current `*config.Config` and dirty field tracking
- Provide `UpdateConfigFromForm`, `UpdateProviderFromForm`, `UpdateAuthProviderFromForm` methods
- Map all field keys to their corresponding config paths
