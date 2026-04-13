## 1. Message Types + Bridge Callbacks

- [x] 1.1 Add DelegationMsg, BudgetWarningMsg, RecoveryMsg, TurnTokenUsageMsg to messages.go
- [x] 1.2 Wire OnDelegation and OnBudgetWarning callbacks in bridge.go enrichRequest()
- [x] 1.3 Add bridge_test.go tests for delegation and budget warning callbacks

## 2. RuntimeTracker (EventBus Bridge)

- [x] 2.1 Create runtimebridge.go with RuntimeTracker struct (tokenSnapshot, runtimeStatus types)
- [x] 2.2 Implement NewRuntimeTracker with TokenUsageEvent subscription (session key + turnActive filter)
- [x] 2.3 Implement RecoveryDecisionEvent subscription with session key filter and sender forwarding
- [x] 2.4 Implement FlushTurnTokens, StartTurn, ResetTurn, RecordDelegation, SetActiveAgent, Snapshot methods
- [x] 2.5 Add runtimebridge_test.go tests (token accumulation, session filter, turnActive gating, recovery forwarding, delegation tracking, reset)

## 3. Thinking Renderer Enhancement

- [x] 3.1 Modify render_thinking.go active state to display summary text preview with ansi.Truncate
- [x] 3.2 Add render_thinking_test.go tests for content preview, empty content, truncation

## 4. Transcript Items + Renderers

- [x] 4.1 Add itemDelegation and itemRecovery kinds to chatview.go
- [x] 4.2 Implement appendDelegation, appendRecovery, appendTokenSummary methods in chatview.go
- [x] 4.3 Add render dispatch for itemDelegation and itemRecovery in chatview.go render()
- [x] 4.4 Create render_delegation.go with renderDelegationBlock (ansi.Strip + ansi.Truncate)
- [x] 4.5 Create render_recovery.go with renderRecoveryBlock and recoveryActionDisplayName
- [x] 4.6 Add render_delegation_test.go and render_recovery_test.go

## 5. Chat Update Handlers

- [x] 5.1 Add DelegationMsg handler (appendDelegation + dismissPending) in chat.go
- [x] 5.2 Add BudgetWarningMsg handler (appendStatus with warning tone) in chat.go
- [x] 5.3 Add RecoveryMsg handler (appendRecovery) in chat.go
- [x] 5.4 Add TurnTokenUsageMsg handler (appendTokenSummary) in chat.go

## 6. Cockpit Routing + RuntimeTracker Wiring

- [x] 6.1 Add runtimeTracker field and SetRuntimeTracker method to cockpit Model
- [x] 6.2 Add StartTurn trigger on ToolStartedMsg/ThinkingStartedMsg/ChunkMsg
- [x] 6.3 Add DelegationMsg intercept (RecordDelegation for outward, SetActiveAgent for return hop)
- [x] 6.4 Add BudgetWarningMsg and RecoveryMsg intercept (forward to chat child)
- [x] 6.5 Add DoneMsg intercept (forward first, then flush tokens, then send TurnTokenUsageMsg, then reset)
- [x] 6.6 Add RuntimeTracker snapshot push on contextTickMsg
- [x] 6.7 Wire NewRuntimeTracker and SetRuntimeTracker in cmd/lango/main.go runCockpit()

## 7. Context Panel Runtime Section

- [x] 7.1 Add runtimeStat field and SetRuntimeStatus method to ContextPanel
- [x] 7.2 Implement renderRuntimeStatus (active agent, delegation count, token count, graceful degradation)
- [x] 7.3 Insert runtime section in View() between Tool Stats and Channels
- [x] 7.4 Add contextpanel_test.go tests for runtime section (running, idle, zero delegations, zero tokens, no agent name)
