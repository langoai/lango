## ADDED Requirements

### Requirement: Field OnChange callback
The `tuicore.Field` struct SHALL support an `OnChange` callback of type `func(string) tea.Cmd` that is invoked when the field's value changes via user interaction (e.g., InputSelect left/right navigation).

#### Scenario: OnChange fires on InputSelect value change
- **WHEN** a user navigates an InputSelect field with left/right keys and the value changes
- **THEN** the `OnChange` callback is invoked with the new value and the returned `tea.Cmd` is executed

#### Scenario: OnChange not fired when value unchanged
- **WHEN** a user presses left/right on an InputSelect with a single option
- **THEN** the `OnChange` callback SHALL NOT be invoked

### Requirement: Field loading state
The `tuicore.Field` struct SHALL have a `Loading` boolean field that indicates an async operation is in progress, and a `LoadError` error field that holds the last fetch error.

#### Scenario: Loading indicator displayed
- **WHEN** a field has `Loading == true`
- **THEN** the form view SHALL display "Loading models..." instead of the field's normal input widget

#### Scenario: Loading cleared on result
- **WHEN** a `FieldOptionsLoadedMsg` is received for a field
- **THEN** the field's `Loading` SHALL be set to `false`

### Requirement: FieldOptionsLoadedMsg async message
The system SHALL define a `FieldOptionsLoadedMsg` message type with `FieldKey`, `ProviderID`, `Options`, and `Err` fields for communicating async model fetch results back to the form.

#### Scenario: Successful options load
- **WHEN** a `FieldOptionsLoadedMsg` with non-empty `Options` and nil `Err` is received
- **THEN** the target field's `Options` SHALL be updated, type set to `InputSearchSelect`, and `FilteredOptions` initialized

#### Scenario: Error options load
- **WHEN** a `FieldOptionsLoadedMsg` with non-nil `Err` is received
- **THEN** the target field SHALL fall back to `InputText` type and display the error in its description

#### Scenario: Message for unknown field key
- **WHEN** a `FieldOptionsLoadedMsg` with a `FieldKey` that matches no field is received
- **THEN** the message SHALL be silently ignored
