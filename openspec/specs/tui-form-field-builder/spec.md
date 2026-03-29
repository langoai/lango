### Requirement: Form field factory functions
The `tuicore` package SHALL provide factory functions that create `*Field` values with common configurations, reducing boilerplate in settings form files.

#### Scenario: BoolInput creates toggle field
- **WHEN** `BoolInput(key, label, checked, desc)` is called
- **THEN** it SHALL return a `*Field` with `Type: InputBool`, `Checked` set to the given value, and `Description` set

#### Scenario: IntInput creates validated integer field
- **WHEN** `IntInput(key, label, value, desc)` is called
- **THEN** it SHALL return a `*Field` with `Type: InputInt`, `Value` set to `strconv.Itoa(value)`, and a `Validate` function that rejects non-positive integers

#### Scenario: SelectInput creates dropdown field
- **WHEN** `SelectInput(key, label, value, options, desc)` is called
- **THEN** it SHALL return a `*Field` with `Type: InputSelect`, `Options` set, and `Value` set

#### Scenario: TextInput creates text field
- **WHEN** `TextInput(key, label, value, desc)` is called
- **THEN** it SHALL return a `*Field` with `Type: InputText` and `Value` set

#### Scenario: TextInputWithPlaceholder creates text field with hint
- **WHEN** `TextInputWithPlaceholder(key, label, value, placeholder, desc)` is called
- **THEN** it SHALL return a `*Field` with `Placeholder` set in addition to standard text field properties

### Requirement: Factory functions coexist with manual AddField
Factory functions SHALL return `*Field` compatible with `FormModel.AddField()`. Existing manual `&Field{...}` patterns SHALL continue to work without modification.

#### Scenario: Mixed usage in same form
- **WHEN** a form uses both `tuicore.BoolInput(...)` and `&tuicore.Field{...}` in the same form
- **THEN** both fields SHALL render and function identically
