## Purpose

Index-based tool call accumulator for assembling streaming LLM deltas into complete FunctionCall parts. Supports both OpenAI (Index-based) and Anthropic (ID/Name fallback) streaming patterns.

## Requirements

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

### Requirement: Empty-name entries are dropped from done output
The accumulator SHALL exclude entries where Name is still empty after all chunks are processed.

#### Scenario: Accumulated entry with no Name
- **WHEN** chunks with Index=0 but no Name field are accumulated
- **THEN** `done()` returns zero Parts for that entry

### Requirement: ID is preserved in assembled FunctionCall
The accumulator SHALL preserve the original ID from the first chunk that carries it. If no ID is provided, it SHALL generate one as "call_" + Name.

#### Scenario: Explicit ID preserved
- **WHEN** a chunk with ID="call_custom_id" and Name="my_tool" is added
- **THEN** the resulting Part has FunctionCall.ID="call_custom_id"

### Requirement: Streaming partial yield only for named chunks
During streaming mode, `GenerateContent` SHALL yield partial tool call responses only when the delta carries a non-empty Name. Arg-only deltas SHALL be accumulated silently without yielding.

#### Scenario: OpenAI three-chunk streaming regression
- **WHEN** three streaming chunks arrive: chunk1(Index=0, ID, Name, partial args), chunk2(Index=0, args), chunk3(Index=0, args)
- **THEN** exactly one partial response is yielded (for chunk1 with Name)
- **AND** the final done response contains one fully assembled FunctionCall with no empty-name parts

### Requirement: convertMessages skips empty-name FunctionCalls
The `convertMessages` function SHALL skip FunctionCall parts where Name is an empty string.

#### Scenario: Mixed valid and empty-name FunctionCalls
- **WHEN** a Content has two FunctionCall parts, one with Name="valid" and one with Name=""
- **THEN** only the valid FunctionCall is included in the output ToolCalls

### Requirement: Shared convertMessages does not perform orphan repair
The `convertMessages()` function SHALL NOT inject synthetic tool responses for orphaned FunctionCalls. Orphan repair is provider-specific and SHALL be handled by each provider's own conversion logic.

#### Scenario: Orphaned FunctionCall passes through without repair
- **WHEN** a genai.Content sequence contains a model FunctionCall followed by a user message with no intervening FunctionResponse
- **THEN** `convertMessages()` SHALL return the messages as-is without injecting synthetic tool messages

### Requirement: convertTools skips empty-name FunctionDeclarations
The `convertTools` function SHALL skip FunctionDeclarations where Name is an empty string.

#### Scenario: Mixed valid and empty-name FunctionDeclarations
- **WHEN** a GenerateContentConfig has two FunctionDeclarations, one with Name="valid_tool" and one with Name=""
- **THEN** only the valid declaration is included in the output Tools
