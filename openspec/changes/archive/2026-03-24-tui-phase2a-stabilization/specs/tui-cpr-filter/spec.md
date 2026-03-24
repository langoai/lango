## ADDED Requirements

### Requirement: CPR sequence detection state machine
The ChatModel SHALL implement a 4-state finite state machine (cprIdle, cprGotEsc, cprGotBracket, cprInParams) that detects ANSI CPR (Cursor Position Report) sequences in the form `ESC[<digits>;<digits>R` arriving as individual tea.KeyMsg events.

#### Scenario: Full CPR sequence discarded
- **WHEN** the terminal emits a CPR response ESC[43;84R as individual KeyMsg events (KeyEscape, '[', '4', '3', ';', '8', '4', 'R')
- **THEN** the entire sequence SHALL be consumed by the filter and no characters SHALL reach the textarea input

#### Scenario: State machine transitions
- **WHEN** KeyEscape is received in cprIdle state
- **THEN** the state SHALL transition to cprGotEsc and the key SHALL be buffered

### Requirement: CPR filter non-CPR flush
When the state machine detects that a buffered sequence is NOT a CPR response (e.g., ESC followed by a non-'[' character), it SHALL flush all buffered keys through the normal input processing path and reset to cprIdle.

#### Scenario: ESC followed by non-bracket flushes
- **WHEN** KeyEscape is followed by a non-'[' character (e.g., 'a')
- **THEN** the buffered ESC and the current character SHALL be replayed through handleKey() and the input component

#### Scenario: Partial sequence with non-digit flushes
- **WHEN** ESC[ is followed by digits then a non-digit/non-semicolon/non-R character
- **THEN** all buffered keys SHALL be flushed through normal input and the state SHALL reset to cprIdle

#### Scenario: R without preceding digits is not CPR
- **WHEN** ESC[R is received (R immediately after bracket, no digits)
- **THEN** the sequence SHALL NOT be treated as CPR and all buffered keys SHALL be flushed

### Requirement: CPR detection timeout
The filter SHALL start a 50ms timeout when transitioning to cprGotEsc. If the timeout expires before the sequence completes, all buffered keys SHALL be flushed as normal input.

#### Scenario: Timeout flushes buffered ESC
- **WHEN** KeyEscape is received and 50ms passes without a following '[' character
- **THEN** the buffered ESC key SHALL be flushed through normal input and the state SHALL reset to cprIdle

#### Scenario: Real Esc key not blocked
- **WHEN** a user presses the physical Esc key (no CPR sequence follows)
- **THEN** the Esc key SHALL be delivered to the normal key handler within 50ms
