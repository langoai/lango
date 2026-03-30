## MODIFIED Requirements

### Requirement: Orchestrator direct response assessment
The orchestrator's Decision Protocol SHALL include a Step 0 (ASSESS) that evaluates whether a request is a simple conversational message (greeting, opinion, math, small talk). If yes, the orchestrator SHALL respond directly without delegation. The ASSESS step SHALL explicitly state that the orchestrator MUST NOT emit any function calls even when responding directly, and that requests requiring real-time data (weather, news, prices, search) SHALL be delegated to the appropriate sub-agent.

#### Scenario: Simple greeting handled directly
- **WHEN** the user sends a greeting like "hello"
- **THEN** the orchestrator SHALL respond directly without delegating to any sub-agent

#### Scenario: Math question handled directly
- **WHEN** the user asks a math question (e.g., "what is 2 + 2?")
- **THEN** the orchestrator SHALL respond directly without delegation

#### Scenario: Weather request delegated
- **WHEN** the user asks about weather (e.g., "tell me today's weather")
- **THEN** the orchestrator SHALL delegate to the appropriate sub-agent (e.g., navigator or librarian)
- **AND** the orchestrator SHALL NOT attempt to respond directly or emit any function calls

#### Scenario: General knowledge requiring search delegated
- **WHEN** the user asks a factual question that may require search (e.g., "what is the latest news?")
- **THEN** the orchestrator SHALL delegate to the appropriate sub-agent

#### Scenario: Tool-requiring request delegated normally
- **WHEN** the user requests an action requiring tools (e.g., "create a wallet")
- **THEN** the orchestrator SHALL delegate to the appropriate sub-agent per the routing table

#### Scenario: No function calls in direct response
- **WHEN** the orchestrator responds directly (ASSESS step determines direct-answer)
- **THEN** the orchestrator SHALL NOT emit any function calls
- **AND** the ASSESS instruction SHALL contain "MUST NOT emit any function calls"

### Requirement: Orchestrator Short-Circuit for Simple Queries
The orchestrator instruction SHALL direct the LLM to respond directly to simple conversational queries (greetings, opinions, math, small talk) without delegating to sub-agents.

#### Scenario: Simple greeting
- **WHEN** user sends a greeting like "hello"
- **THEN** the orchestrator SHALL respond directly without delegation

#### Scenario: Task requiring tools
- **WHEN** user requests an action requiring tool execution
- **THEN** the orchestrator SHALL delegate to the appropriate sub-agent

#### Scenario: Direct-answer list excludes real-time topics
- **WHEN** the orchestrator instruction is built
- **THEN** the direct-answer categories in ASSESS and Delegation Rules SHALL NOT include "weather" or "general knowledge"
