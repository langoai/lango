## ADDED Requirements

### Requirement: Specialist completion contract
After a specialist sub-agent receives successful tool output, the runtime SHALL require the specialist turn to end in exactly one of: visible assistant completion, `transfer_to_agent`, or a structured incomplete outcome emitted by the runtime. Tool-only terminal states SHALL NOT be treated as successful completion.

#### Scenario: Successful tool use followed by visible completion
- **WHEN** `vault` successfully retrieves wallet balance information
- **THEN** the specialist turn SHALL produce a visible assistant completion summarizing the result
- **AND** the runtime SHALL classify the turn as successful

#### Scenario: Tool-only terminal state becomes incomplete outcome
- **WHEN** a specialist receives successful tool output but terminates without visible completion or transfer
- **THEN** the runtime SHALL terminate the specialist turn with a structured incomplete outcome
- **AND** the parent turn SHALL NOT treat the specialist turn as a silent success

### Requirement: Repeated identical specialist call containment
Within a single user turn, repeated calls from the same specialist to the same tool with canonically equal params SHALL be detected and stopped even if the model changes call IDs between attempts.

#### Scenario: Same tool same params repeated with different call IDs
- **WHEN** `vault` repeatedly calls `payment_balance` with `{}` and each attempt has a different call ID
- **THEN** the runtime SHALL still count those attempts against the same call-signature loop budget
- **AND** SHALL stop the loop when the threshold is reached

#### Scenario: Different params do not trip the identical-call budget immediately
- **WHEN** the same specialist calls the same tool name with materially different params
- **THEN** the runtime SHALL treat those calls as distinct signatures for loop containment purposes

### Requirement: Evidence-only orchestrator recovery
When the orchestrator has no direct tools and a delegated specialist fails or returns an incomplete outcome, the orchestrator SHALL either re-route to another agent or answer only from evidence already gathered in the turn trace. It SHALL NOT emit direct FunctionCalls to specialist-only tools.

#### Scenario: Tool-less orchestrator cannot call specialist tool directly
- **WHEN** a previous `vault` turn failed and the orchestrator enters recovery
- **THEN** the orchestrator SHALL NOT emit a direct FunctionCall to `payment_balance`
- **AND** SHALL instead re-route, answer from existing evidence, or report an inability to complete

#### Scenario: Recovery answer references gathered evidence only
- **WHEN** the orchestrator answers after a specialist failure
- **THEN** the answer SHALL be derived only from tool results or summaries already recorded in the current turn trace
- **AND** it SHALL NOT claim that a new unavailable tool call was executed
