# Tasks: P2P Team-Escrow-Workspace Connectivity

## Wave 1: Independent Units

- [x] **1.1** Add TeamFormedEvent publishing to FormTeam() after t.Activate()
- [x] **1.2** Add TeamTaskDelegatedEvent publishing to DelegateTask() before wg.Wait()
- [x] **1.3** Add TeamTaskCompletedEvent publishing to DelegateTask() after wg.Wait()
- [x] **1.4** Add TeamConflictDetectedEvent publishing to CollectResults() on resolver error
- [x] **1.5** Add TeamDisbandedEvent publishing to DisbandTeam() after t.Disband()
- [x] **1.6** Add event publishing tests (setupCoordinatorWithBus helper + 3 test functions)
- [x] **2.1** Create tools_team.go with buildTeamTools() returning 5 tools
- [x] **2.2** Implement team_form tool (FormTeam + UUID generation)
- [x] **2.3** Implement team_delegate tool (DelegateTask + CollectResults)
- [x] **2.4** Implement team_status, team_list, team_disband tools
- [x] **3.1** Add TeamHandler type and field to Handler struct
- [x] **3.2** Add SetTeamHandler() setter method
- [x] **3.3** Add team request type cases to handleRequest switch
- [x] **3.4** Implement handleTeamMessage() method
- [x] **3.5** Create team_handler.go with TeamRouter and typed handlers
- [x] **8.1** Add SendTeamInvite() to P2PRemoteAgent
- [x] **8.2** Add SendTeamTask() to P2PRemoteAgent
- [x] **8.3** Add SendTeamDisband() to P2PRemoteAgent

## Wave 2: Event-Driven Bridges

- [x] **4.1** Create bridge_team_escrow.go with wireTeamEscrowBridge()
- [x] **4.2** Implement TeamFormedEvent → escrow creation with per-worker milestones
- [x] **4.3** Implement TeamTaskCompletedEvent → complete next pending milestone
- [x] **4.4** Implement TeamDisbandedEvent → release or dispute+refund
- [x] **5.1** Create bridge_team_budget.go with wireTeamBudgetBridge()
- [x] **5.2** Implement TeamFormedEvent → budget allocation + TeamPaymentAgreedEvent
- [x] **5.3** Implement TeamTaskDelegatedEvent → budget reserve
- [x] **5.4** Implement TeamTaskCompletedEvent → budget record
- [x] **6.1** Create bridge_workspace_team.go with wireWorkspaceTeamBridge()
- [x] **6.2** Implement TeamFormedEvent → auto-create workspace
- [x] **6.3** Implement TeamTaskCompletedEvent → record contribution
- [x] **6.4** Implement TeamDisbandedEvent → cleanup

## Wave 3: Integration

- [x] **7.1** Register team tools in app.go after P2P tools
- [x] **7.2** Wire team-escrow bridge in app.go after economy init
- [x] **7.3** Wire team-budget bridge in app.go after economy init
- [x] **7.4** Wire workspace-team bridge in app.go inside workspace section
- [x] **7.5** Wire team protocol handler in wiring_p2p.go with TeamRouter
- [x] **7.6** Register convenience tools in app.go
- [x] **9.1** Create tools_team_escrow.go with buildTeamEscrowTools()
- [x] **9.2** Implement team_form_with_budget tool
- [x] **9.3** Implement team_complete_milestone tool
- [x] **10.1** Create bridge_integration_test.go with setupBridgeTestEnv()
- [x] **10.2** TestBridge_TeamFormed_CreatesEscrowAndBudget
- [x] **10.3** TestBridge_TeamTaskCompleted_CompletesMilestoneAndRecordsBudget
- [x] **10.4** TestBridge_TeamDisbanded_RefundsIncompleteEscrow
- [x] **10.5** TestBridge_FullLifecycle_ReleasesOnAllMilestonesCompleted
- [x] **10.6** TestBridge_NoBudget_SkipsBridges
