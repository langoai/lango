## Purpose

Capability spec for store-util-helpers. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: MarshalField helper
The `storeutil` package SHALL provide a `MarshalField(v interface{}) (json.RawMessage, error)` function that marshals a value to JSON. It SHALL return an error on marshal failure rather than swallowing it, ensuring store persistence callers can abort writes.

#### Scenario: Successful marshal
- **WHEN** `MarshalField` is called with a valid struct
- **THEN** it SHALL return the JSON-encoded bytes and nil error

#### Scenario: Marshal failure propagates error
- **WHEN** `MarshalField` is called with an unmarshalable value (e.g., `math.NaN()`)
- **THEN** it SHALL return nil data and a non-nil error wrapping the underlying marshal error

### Requirement: UnmarshalField helper
The `storeutil` package SHALL provide an `UnmarshalField(data []byte, target interface{}, context string) error` function that unmarshals JSON with context-enriched error messages.

#### Scenario: Successful unmarshal
- **WHEN** `UnmarshalField` is called with valid JSON and a matching target
- **THEN** it SHALL populate the target and return nil

#### Scenario: Unmarshal failure includes context
- **WHEN** `UnmarshalField` fails to unmarshal
- **THEN** the error message SHALL include the context string for debugging

### Requirement: CopySlice generic helper
The `storeutil` package SHALL provide a `CopySlice[T any](src []T) []T` function that returns an independent copy. It SHALL return nil when src is nil.

#### Scenario: Independent copy
- **WHEN** `CopySlice` is called and the copy is mutated
- **THEN** the original slice SHALL remain unchanged

### Requirement: CopyMap generic helper
The `storeutil` package SHALL provide a `CopyMap[K comparable, V any](src map[K]V) map[K]V` function that returns a shallow independent copy. It SHALL return nil when src is nil.

#### Scenario: Independent copy
- **WHEN** `CopyMap` is called and the copy is mutated
- **THEN** the original map SHALL remain unchanged
