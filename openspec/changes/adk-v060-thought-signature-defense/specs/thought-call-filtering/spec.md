## ADDED Requirements

### Requirement: OpenAI provider filters Gemini thought tool calls
The OpenAI provider's `convertParams()` SHALL identify tool calls with `Thought=true` and exclude them from the converted message sequence. Both the FunctionCall entry and its corresponding tool response message (matched by tool_call_id) SHALL be removed.

#### Scenario: Thought tool call and response filtered from mixed history
- **WHEN** the message history contains an assistant message with both a thought tool call (`Thought=true`) and a normal tool call, followed by tool responses for each
- **THEN** the converted OpenAI request SHALL contain only the normal tool call and its response; the thought call and its response SHALL be absent

#### Scenario: All tool calls are thought calls
- **WHEN** an assistant message contains only thought tool calls
- **THEN** the converted OpenAI request SHALL contain the assistant message with zero tool calls, and all corresponding tool responses SHALL be removed

#### Scenario: No thought calls leaves messages unchanged
- **WHEN** no tool calls in the history have `Thought=true`
- **THEN** the converted OpenAI request SHALL contain all original messages and tool calls unmodified

### Requirement: Thought call filtering uses droppedThoughtIDs set
The filtering mechanism SHALL build a set of dropped tool call IDs (`droppedThoughtIDs`) in a single pre-scan pass, then use that set to filter both FunctionCall entries and tool response messages in the conversion loop.

#### Scenario: Paired deletion by ID
- **WHEN** a thought tool call has ID "call_thought_1"
- **THEN** any tool response message with `tool_call_id: "call_thought_1"` SHALL also be removed
