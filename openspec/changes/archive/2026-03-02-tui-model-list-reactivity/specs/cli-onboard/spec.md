## MODIFIED Requirements

### Requirement: Agent step reactive model list
The Onboard Agent step form SHALL wire `OnChange` on the provider field to asynchronously fetch and update the model field when the provider changes. The model field SHALL use `InputSearchSelect` type.

#### Scenario: Provider change in onboard triggers model refresh
- **WHEN** a user changes the provider in the Agent step of the onboard wizard
- **THEN** the model field SHALL show loading state, fetch models from the new provider, and update the placeholder with `suggestModel(newProvider)`

#### Scenario: Model fetch error shows feedback
- **WHEN** model fetching fails during onboard Agent step
- **THEN** the model field SHALL fall back to `InputText` with an error message in the description

### Requirement: Wizard forwards async messages
The onboard Wizard's `Update()` method SHALL forward non-key, non-window messages to the active form so that `FieldOptionsLoadedMsg` and other async results reach the form's update handler.

#### Scenario: FieldOptionsLoadedMsg reaches active form
- **WHEN** the Wizard receives a `FieldOptionsLoadedMsg` while on a form step
- **THEN** the message SHALL be forwarded to `activeForm.Update()` for processing
