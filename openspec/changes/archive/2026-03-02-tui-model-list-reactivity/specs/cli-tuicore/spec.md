## MODIFIED Requirements

### Requirement: Field struct
The `tuicore.Field` struct SHALL include `OnChange func(newValue string) tea.Cmd`, `Loading bool`, and `LoadError error` fields in addition to all existing fields. The `OnChange` callback SHALL be invoked by the form when an InputSelect field value changes via user interaction.

#### Scenario: Field with OnChange on InputSelect
- **WHEN** an InputSelect field with an `OnChange` callback changes value via left/right keys
- **THEN** the form SHALL invoke `OnChange(newValue)` and execute the returned `tea.Cmd`

#### Scenario: Field with nil OnChange
- **WHEN** an InputSelect field has a nil `OnChange` and changes value
- **THEN** the form SHALL proceed normally without invoking any callback

### Requirement: FormModel handles FieldOptionsLoadedMsg
The `FormModel.Update()` method SHALL handle `FieldOptionsLoadedMsg` by finding the target field by `FieldKey` and updating its options, type, loading state, and description accordingly.

#### Scenario: Successful async model load
- **WHEN** `FormModel.Update()` receives a `FieldOptionsLoadedMsg` with options
- **THEN** the matching field SHALL be updated to `InputSearchSelect` with the new options and `Loading` set to false

#### Scenario: Failed async model load
- **WHEN** `FormModel.Update()` receives a `FieldOptionsLoadedMsg` with an error
- **THEN** the matching field SHALL fall back to `InputText`, set `LoadError`, and show the error in its description
