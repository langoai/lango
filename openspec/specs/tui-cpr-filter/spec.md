### Requirement: CPR sequence detection state machine
The ChatModel SHALL implement a 4-state finite state machine (`cprIdle`, `cprGotEsc`, `cprGotBracket`, `cprInParams`) that detects ANSI CPR (Cursor Position Report) sequences in the form `ESC[<digits>;<digits>R` arriving as individual `tea.KeyMsg` events before they reach the idle composer input path.

#### Scenario: Full CPR sequence discarded
- **WHEN** the terminal emits a CPR response ESC[43;84R as individual KeyMsg events (KeyEscape, '[', '4', '3', ';', '8', '4', 'R')
- **THEN** the entire sequence SHALL be consumed by the filter and no characters SHALL reach the textarea input

#### Scenario: State machine transitions
- **WHEN** `KeyEscape` is received in `cprIdle`
- **THEN** the state SHALL transition to `cprGotEsc` and the key SHALL be buffered

#### Scenario: CPR filter scoped to idle input
- **WHEN** the TUI is in approval state or another non-composer interaction mode
- **THEN** the CPR filter SHALL NOT intercept unrelated key handling for that state

### Requirement: CPR filter non-CPR flush
When the state machine detects that a buffered sequence is NOT a CPR response, it SHALL replay buffered keys in order through the normal idle input processing path and reset to `cprIdle`.

#### Scenario: ESC followed by non-bracket flushes
- **WHEN** `KeyEscape` is followed by a non-`[` key while the composer is active
- **THEN** the buffered `Esc` and the current key SHALL be replayed through the normal idle input path in order

#### Scenario: Partial sequence with non-digit flushes
- **WHEN** `ESC[` is followed by digits and then a non-digit, non-semicolon, non-`R` character
- **THEN** all buffered keys SHALL be replayed through the normal idle input path and the state SHALL reset to `cprIdle`

#### Scenario: Alt sequence preserved
- **WHEN** an `Alt+key` sequence is delivered as `Esc` followed by another key while the composer is active
- **THEN** the sequence SHALL be replayed through the normal idle input path instead of being discarded as CPR

### Requirement: CPR detection timeout
The filter SHALL start a 50ms timeout when transitioning to `cprGotEsc`. If the timeout expires before a CPR sequence completes, buffered keys SHALL be replayed through the normal idle input path.

#### Scenario: Timeout flushes buffered ESC
- **WHEN** `KeyEscape` is received and 50ms passes without a following `[` character
- **THEN** the buffered `Esc` key SHALL be replayed through the normal idle input path and the state SHALL reset to `cprIdle`

#### Scenario: Real Esc key not blocked
- **WHEN** a user presses the physical `Esc` key and no CPR sequence follows within 50ms
- **THEN** the `Esc` key SHALL be delivered to the normal idle input path and the filter SHALL reset to `cprIdle`
