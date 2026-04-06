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

### Requirement: OSC sequence detection
The terminal-response input guard SHALL detect OSC response sequences beginning with `Esc ]` and ending with BEL or ST (`Esc \`) before they reach the idle composer input path.

#### Scenario: OSC 11 response discarded
- **WHEN** the terminal emits an OSC 11 response such as `Esc ] 11 ; rgb:... BEL` while the composer is active
- **THEN** the entire sequence SHALL be consumed by the guard and no characters SHALL reach the composer textarea

#### Scenario: OSC ST termination discarded
- **WHEN** the terminal emits an OSC response terminated by ST (`Esc \`)
- **THEN** the full buffered sequence SHALL be discarded once the terminating `Esc \` is received

### Requirement: Terminal response guard non-match replay
When buffered input does not complete a recognized CPR or OSC response, the guard SHALL replay buffered keys in order through the normal idle input path and reset to idle.

#### Scenario: Esc followed by non-bracket, non-osc key
- **WHEN** `KeyEscape` is followed by a key other than `[` or `]` while the composer is active
- **THEN** the buffered `Esc` and current key SHALL be replayed through the normal idle input path in order

#### Scenario: Alt sequence preserved
- **WHEN** an `Alt+key` sequence is delivered as `Esc` followed by another key while the composer is active
- **THEN** the sequence SHALL be replayed through the normal idle input path instead of being discarded as terminal response noise

### Requirement: Terminal response timeout
The guard SHALL start a 50ms timeout when buffering an initial `Esc`. If no recognized CPR or OSC sequence completes within that window, buffered keys SHALL be replayed through the normal idle input path.

#### Scenario: Real Esc key restored
- **WHEN** a user presses the physical `Esc` key and no CPR or OSC response follows within 50ms
- **THEN** the `Esc` key SHALL be delivered to the normal idle input path and the guard SHALL reset to idle
