## MODIFIED Requirements

### Requirement: Cockpit Update dispatches via named handler methods
The cockpit `Update()` SHALL dispatch messages to named handler methods instead of inline intercept blocks. Each message type SHALL have a dedicated method: `handleContextTick`, `handleChannelMessage`, `handleApprovalRequest`, `handleDelegation`, `handleBudgetWarning`, `handleRecovery`, `handleDone`, and `markTurnStarted`.

#### Scenario: Channel message routed via handleChannelMessage
- **WHEN** a `ChannelMessageMsg` arrives while a non-chat page is active
- **THEN** `handleChannelMessage` SHALL forward the message to the chat child model

#### Scenario: Done message routed via handleDone
- **WHEN** a `DoneMsg` arrives
- **THEN** `handleDone` SHALL forward to chat child first, flush turn tokens, send TurnTokenUsageMsg, and reset the runtime tracker

#### Scenario: Turn started marked via markTurnStarted
- **WHEN** a `ToolStartedMsg`, `ThinkingStartedMsg`, or `ChunkMsg` arrives
- **THEN** `markTurnStarted()` SHALL call `runtimeTracker.StartTurn()` if the tracker is available

#### Scenario: Unregistered page click is no-op
- **WHEN** a sidebar PageSelectedMsg selects a PageID that has no registered page
- **THEN** `activePage` SHALL change but no page.Activate() SHALL be called and forwardToActive SHALL silently no-op
