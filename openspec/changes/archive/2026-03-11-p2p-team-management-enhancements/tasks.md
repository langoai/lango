# Tasks: P2P Team Management Enhancements

## 1. Team Model Enhancements

- [x] 1.1 Add Budget, Spent, StatusShuttingDown, StatusCompleted fields to Team type
- [x] 1.2 Add Members() method returning deep copies of all members
- [x] 1.3 Add MemberCount() method returning the number of members
- [x] 1.4 Add ActiveMembers() method filtering out MemberLeft/MemberFailed members
- [x] 1.5 Add AddSpend() method with ErrBudgetExceeded guard
- [x] 1.6 Add RoleReviewer constant alongside existing role types
- [x] 1.7 Implement custom MarshalJSON/UnmarshalJSON for members map-to-slice serialization
- [x] 1.8 Add sentinel errors: ErrTeamFull, ErrBudgetExceeded, ErrAlreadyMember, ErrNotMember, ErrTeamDisbanded, ErrTeamShuttingDown

## 2. BoltDB Persistent Store

- [x] 2.1 Define TeamStore interface (Save, Load, LoadAll, Delete)
- [x] 2.2 Implement BoltStore with NewBoltStore creating teams bucket
- [x] 2.3 Implement BoltStore.Save with JSON marshal into BoltDB
- [x] 2.4 Implement BoltStore.Load with ErrTeamNotFound for missing keys
- [x] 2.5 Implement BoltStore.LoadAll with corrupt entry skip-and-warn
- [x] 2.6 Implement BoltStore.Delete
- [x] 2.7 Write comprehensive BoltStore tests (save, load, loadAll, delete, not-found, corrupt)

## 3. Coordinator Persistence Integration

- [x] 3.1 Add TeamStore field to CoordinatorConfig and Coordinator
- [x] 3.2 Persist team in FormTeam() after store.Save()
- [x] 3.3 Persist team in DelegateTask() after task completion
- [x] 3.4 Delete team in DisbandTeam() via store.Delete()
- [x] 3.5 Implement LoadPersistedTeams() to restore Active/Forming teams from store
- [x] 3.6 Persist team in KickMember() after member removal

## 4. Coordinator Membership Management

- [x] 4.1 Implement KickMember(ctx, teamID, memberDID, reason) with event publishing
- [x] 4.2 Implement TeamsForMember(did) returning active team IDs for a member
- [x] 4.3 Add GetTeam() method to Coordinator
- [x] 4.4 Add ListTeams() and ActiveTeams() methods

## 5. Graceful Shutdown

- [x] 5.1 Implement GracefulShutdown(ctx, teamID, reason) in coordinator_shutdown.go
- [x] 5.2 Transition team to StatusShuttingDown blocking new DelegateTask calls
- [x] 5.3 Calculate proportional milestone settlement based on active members and spend
- [x] 5.4 Publish TeamGracefulShutdownEvent with settlement details
- [x] 5.5 Call DisbandTeam() to complete the shutdown

## 6. Health Monitor

- [x] 6.1 Define HealthMonitorConfig struct with Coordinator, Bus, Logger, Interval, MaxMissed, InvokeFn
- [x] 6.2 Implement NewHealthMonitor with default interval (30s) and maxMissed (3)
- [x] 6.3 Implement lifecycle.Component interface (Name, Start, Stop)
- [x] 6.4 Implement periodic health check loop with ticker and stop channel
- [x] 6.5 Implement checkAll() iterating active teams
- [x] 6.6 Implement checkTeam() pinging non-leader active members concurrently
- [x] 6.7 Implement pingMember() with 10s timeout, miss counter tracking, and unhealthy detection
- [x] 6.8 Publish TeamHealthCheckEvent after each team sweep with healthy/total counts
- [x] 6.9 Subscribe to TeamTaskCompletedEvent to reset miss counters
- [x] 6.10 Subscribe to TeamDisbandedEvent to cleanup tracking maps
- [x] 6.11 Write health monitor tests (ping success/fail, threshold detection, counter reset)

## 7. Event Types

- [x] 7.1 Add TeamHealthCheckEvent (TeamID, Healthy, Total)
- [x] 7.2 Add TeamMemberUnhealthyEvent (TeamID, MemberDID, MemberName, MissedPings, LastSeenAt)
- [x] 7.3 Add TeamBudgetWarningEvent (TeamID, Threshold, Spent, Budget)
- [x] 7.4 Add TeamGracefulShutdownEvent (TeamID, Reason, BundlesCreated, MembersSettled)
- [x] 7.5 Add TeamLeaderChangedEvent (TeamID, OldLeaderDID, NewLeaderDID)

## 8. Configuration

- [x] 8.1 Add TeamConfig struct to config package (HealthCheckInterval, MaxMissedHeartbeats, MinReputationScore)
- [x] 8.2 Add Team field to P2PConfig

## 9. Integration Tests

- [x] 9.1 Create integration_test.go with end-to-end team formation test
- [x] 9.2 Write test for task delegation with event verification
- [x] 9.3 Write test for team disbanding lifecycle
- [x] 9.4 Write test for BoltDB persistence round-trip across coordinator restart
- [x] 9.5 Write test for health monitor integration with coordinator
