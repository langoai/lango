## MODIFIED Requirements

### Requirement: Terminal response detection state machine
The chat package SHALL implement a terminal-response input guard as a standalone `cprFilter` struct with `Filter`, `Flush`, and `HandleTimeout` methods. The guard SHALL detect CPR sequences in the form `ESC[<digits>;<digits>R` and OSC responses. ChatModel SHALL delegate to `cprFilter` and handle key replay independently.

#### Scenario: Full CPR sequence discarded
- **WHEN** the terminal emits a CPR response `ESC[43;84R` as individual `tea.KeyMsg` events while the composer is active
- **THEN** the cprFilter SHALL consume the entire sequence via `Filter()` and no characters SHALL reach the composer textarea

#### Scenario: Flush returns buffered keys for replay
- **WHEN** buffered input does not complete a recognized CPR or OSC response
- **THEN** `cprFilter.Flush()` SHALL return the buffered `[]tea.KeyMsg` and reset to idle state
- **AND** ChatModel SHALL replay the returned keys through `handleKey`/`input.Update`

#### Scenario: HandleTimeout flushes on mid-sequence timeout
- **WHEN** `cprTimeoutMsg` fires while the cprFilter is in a non-idle state
- **THEN** `HandleTimeout()` SHALL return the buffered keys for ChatModel to replay
