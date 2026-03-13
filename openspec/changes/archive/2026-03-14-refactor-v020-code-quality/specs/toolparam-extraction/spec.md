## ADDED Requirements

### Requirement: Type-safe required parameter extraction
The `toolparam.RequireString` function SHALL extract a string parameter from `map[string]interface{}` and return `ErrMissingParam` when the key is absent or the value is empty.

#### Scenario: Key exists with non-empty value
- **WHEN** `RequireString(params, "workspaceId")` is called and params contains `"workspaceId": "ws-1"`
- **THEN** it returns `("ws-1", nil)`

#### Scenario: Key is missing
- **WHEN** `RequireString(params, "workspaceId")` is called and params does not contain `"workspaceId"`
- **THEN** it returns `("", ErrMissingParam{Name: "workspaceId"})`

#### Scenario: Key exists with empty string
- **WHEN** `RequireString(params, "workspaceId")` is called and params contains `"workspaceId": ""`
- **THEN** it returns `("", ErrMissingParam{Name: "workspaceId"})`

### Requirement: Optional parameter extraction with fallback
The `toolparam` package SHALL provide `OptionalString`, `OptionalInt`, and `OptionalBool` functions that return a fallback value when the key is absent or has wrong type.

#### Scenario: OptionalString with missing key
- **WHEN** `OptionalString(params, "format", "json")` is called and params does not contain `"format"`
- **THEN** it returns `"json"`

#### Scenario: OptionalInt with present key
- **WHEN** `OptionalInt(params, "limit", 10)` is called and params contains `"limit": float64(50)`
- **THEN** it returns `50`

#### Scenario: OptionalBool with wrong type
- **WHEN** `OptionalBool(params, "verbose", false)` is called and params contains `"verbose": "yes"`
- **THEN** it returns `false` (the fallback)

### Requirement: String slice extraction
The `toolparam.StringSlice` function SHALL extract a `[]string` from `map[string]interface{}` supporting both `[]interface{}` and `[]string` underlying types.

#### Scenario: Value is []interface{} with strings
- **WHEN** `StringSlice(params, "tags")` is called and params contains `"tags": []interface{}{"a", "b"}`
- **THEN** it returns `[]string{"a", "b"}`

#### Scenario: Value is missing
- **WHEN** `StringSlice(params, "tags")` is called and params does not contain `"tags"`
- **THEN** it returns `nil`

### Requirement: ErrMissingParam supports errors.As
The `ErrMissingParam` type SHALL implement the `error` interface and be matchable via `errors.As()`.

#### Scenario: errors.As matching
- **WHEN** `RequireString` returns an error for missing key `"id"`
- **THEN** `errors.As(err, &ErrMissingParam{})` returns true and the matched error has `Name == "id"`

### Requirement: Standardized response builders
The `toolparam` package SHALL provide `StatusResponse` and `ListResponse` builders returning `map[string]interface{}`.

#### Scenario: StatusResponse with extras
- **WHEN** `StatusResponse("ok", func(r Response) { r["count"] = 5 })` is called
- **THEN** it returns `Response{"status": "ok", "count": 5}`

#### Scenario: ListResponse
- **WHEN** `ListResponse("items", []string{"a"}, 1)` is called
- **THEN** it returns `Response{"items": []string{"a"}, "count": 1}`

### Requirement: Tool handler migration to toolparam
All 12 `internal/app/tools_*.go` files SHALL use `toolparam.RequireString` for required parameters instead of inline type assertion and empty-check patterns.

#### Scenario: Consistent error format
- **WHEN** any tool handler receives a missing required parameter
- **THEN** the error message follows the format `"missing <paramName> parameter"`
