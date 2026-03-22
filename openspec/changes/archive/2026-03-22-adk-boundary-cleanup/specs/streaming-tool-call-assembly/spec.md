## MODIFIED Requirements

### Requirement: Accumulator assembles OpenAI streaming deltas by Index
The `toolCallAccumulator` SHALL use a provider-agnostic state machine with states `Idle`, `Receiving`, and `Complete`. For OpenAI patterns, the Index field SHALL transition the state from Idle to Receiving and correlate subsequent chunks.

#### Scenario: Single complete tool call
- **WHEN** a single ToolCall with Index=0, ID, Name, and Arguments is added
- **THEN** `done()` returns exactly one Part with the correct Name, ID, and parsed Args

#### Scenario: OpenAI multi-chunk streaming
- **WHEN** a first chunk with Index=0, ID="call_abc", Name="exec", and partial Arguments is added
- **AND** a second chunk with Index=0 and remaining Arguments is added
- **THEN** `done()` returns one Part with Name="exec", ID="call_abc", and fully concatenated Args

#### Scenario: OpenAI interleaved multiple tool calls
- **WHEN** chunks for Index=0 and Index=1 arrive interleaved
- **THEN** `done()` returns two Parts sorted by Index, each with independently assembled Args

### Requirement: Accumulator assembles Anthropic streaming deltas by fallback chain
The accumulator SHALL support providers that do not supply an Index field using the same state machine. When Index is nil but ID or Name is present, it SHALL transition to a new `Receiving` state. When Index is nil and both ID and Name are absent, it SHALL append to the current `Receiving` state's entry.

#### Scenario: Anthropic start + delta
- **WHEN** a start chunk with ID="tool_1", Name="exec" (no Index) is added
- **AND** a delta chunk with only Arguments (no Index, ID, or Name) is added
- **THEN** `done()` returns one Part with Name="exec", ID="tool_1", and the delta's Args

#### Scenario: Anthropic multiple sequential tool calls
- **WHEN** start1(ID+Name) → delta1(Args) → start2(ID+Name) → delta2(Args) are added
- **THEN** `done()` returns two independent Parts in order

### Requirement: Orphan deltas are dropped
The accumulator SHALL drop delta chunks that arrive when the state machine is in `Idle` state (no preceding start chunk) and log a warning.

#### Scenario: Delta with no preceding start
- **WHEN** a delta chunk with only Arguments is added as the first chunk (state machine is Idle)
- **THEN** `done()` returns zero Parts
- **AND** a warning SHALL be logged indicating the orphaned delta was dropped
