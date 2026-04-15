# interruptible-streaming Specification

## Purpose
TBD - created by archiving change ux-elastic-turns. Update Purpose after archive.
## Requirements
### Requirement: Pending redirect input queue
The `ChatModel` SHALL maintain a `pendingRedirectInput` string field. When the user submits input during `stateStreaming`, the input SHALL be stored in this field (not submitted immediately), the current turn SHALL be cancelled via `cancelFn()`, and the partial stream SHALL be finalized with an `[interrupted]` marker.

#### Scenario: User types and submits during streaming
- **WHEN** the user presses Enter with non-empty input while `state == stateStreaming`
- **THEN** the input value SHALL be stored in `pendingRedirectInput`
- **AND** `cancelFn()` SHALL be called to cancel the current turn
- **AND** the input field SHALL be reset
- **AND** the partial stream SHALL be finalized with an `[interrupted]` marker

#### Scenario: Empty input during streaming is ignored
- **WHEN** the user presses Enter with empty input while `state == stateStreaming`
- **THEN** no redirect SHALL be queued and no cancellation SHALL occur

### Requirement: Redirect consumption in DoneMsg handler
The `DoneMsg` handler SHALL check `pendingRedirectInput` before any other processing (including the `stateFailed` + error status path). If the field is non-empty, the handler SHALL skip error/cancelled message display, transition to `stateIdle`, submit the pending input via `submitCmd()`, and clear the field.

#### Scenario: DoneMsg with pending redirect
- **WHEN** a `DoneMsg` arrives and `pendingRedirectInput != ""`
- **THEN** the handler SHALL NOT display any error or "Generation cancelled" message
- **AND** SHALL transition to `stateIdle`
- **AND** SHALL call `submitCmd(pendingRedirectInput)`
- **AND** SHALL clear `pendingRedirectInput` to `""`

#### Scenario: DoneMsg without pending redirect
- **WHEN** a `DoneMsg` arrives and `pendingRedirectInput == ""`
- **THEN** the existing DoneMsg processing logic SHALL execute unchanged

### Requirement: Composer active during streaming
During `stateStreaming`, the input composer SHALL be focused (not blurred) and its placeholder SHALL indicate that typing will interrupt the current turn.

#### Scenario: Composer focused during streaming
- **WHEN** the chat transitions to `stateStreaming`
- **THEN** the composer SHALL call `Focus()` (not `Blur()`)
- **AND** the placeholder SHALL be set to a message indicating interrupt capability (e.g., "Type to interrupt and redirect...")

#### Scenario: inputAcceptsText includes stateStreaming
- **WHEN** `inputAcceptsText()` is called during `stateStreaming`
- **THEN** it SHALL return `true`

