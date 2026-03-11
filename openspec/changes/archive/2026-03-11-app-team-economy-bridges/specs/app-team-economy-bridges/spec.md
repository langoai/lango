# Spec: App Team-Economy Bridges

## Overview

Event-driven bridge layer connecting P2P team lifecycle events to escrow engine state transitions, reputation score adjustments, and budget-triggered graceful shutdown. All bridges live in `internal/app/` and use `eventbus.SubscribeTyped` to react to cross-subsystem events.

## ADDED Requirements

### Requirement: On-chain escrow event reconciliation
The system SHALL synchronize the local escrow engine state with on-chain escrow contract events by subscribing to deposit, release, refund, dispute, and resolved events and triggering the corresponding escrow engine transitions.

#### Scenario: Deposit event triggers fund and activate
- **WHEN** an `EscrowOnChainDepositEvent` is published with a non-empty escrow ID
- **THEN** the system SHALL call `engine.Fund` followed by `engine.Activate` on the escrow

#### Scenario: Release event triggers release
- **WHEN** an `EscrowOnChainReleaseEvent` is published with a non-empty escrow ID
- **THEN** the system SHALL call `engine.Release` on the escrow

#### Scenario: Refund event triggers refund
- **WHEN** an `EscrowOnChainRefundEvent` is published with a non-empty escrow ID
- **THEN** the system SHALL call `engine.Refund` on the escrow

#### Scenario: Dispute event triggers dispute
- **WHEN** an `EscrowOnChainDisputeEvent` is published with a non-empty escrow ID
- **THEN** the system SHALL call `engine.Dispute` on the escrow with reason "on-chain dispute"

#### Scenario: Resolved event triggers release or refund based on outcome
- **WHEN** an `EscrowOnChainResolvedEvent` is published with `SellerFavor=true`
- **THEN** the system SHALL call `engine.Release` on the escrow
- **WHEN** an `EscrowOnChainResolvedEvent` is published with `SellerFavor=false`
- **THEN** the system SHALL call `engine.Refund` on the escrow

#### Scenario: Idempotent transition handling
- **WHEN** an on-chain event triggers a transition that has already been applied (ErrInvalidTransition)
- **THEN** the system SHALL log at debug level and NOT treat it as an error

#### Scenario: Empty escrow ID is ignored
- **WHEN** an on-chain event is published with an empty escrow ID
- **THEN** the system SHALL skip the event without calling any engine transition

### Requirement: Team reputation adjustment on task outcomes
The system SHALL adjust peer reputation scores based on team task results and health events, and SHALL evict members whose reputation drops below a configurable minimum threshold.

#### Scenario: Unhealthy member gets timeout recorded and potentially kicked
- **WHEN** a `TeamMemberUnhealthyEvent` is published
- **THEN** the system SHALL call `repStore.RecordTimeout` for the unhealthy member
- **AND** if the member's score drops below `minScore`, the system SHALL call `coordinator.KickMember`

#### Scenario: Successful task boosts worker reputation
- **WHEN** a `TeamTaskCompletedEvent` is published with `Successful > 0`
- **THEN** the system SHALL call `repStore.RecordSuccess` for each active worker in the team who has not failed

#### Scenario: Reputation drop triggers eviction from all teams
- **WHEN** a `ReputationChangedEvent` is published with `NewScore < minScore`
- **THEN** the system SHALL call `coordinator.KickMember` for the peer in every team they belong to

#### Scenario: Reputation above threshold is ignored
- **WHEN** a `ReputationChangedEvent` is published with `NewScore >= minScore`
- **THEN** the system SHALL take no eviction action

### Requirement: Budget-triggered team shutdown
The system SHALL trigger graceful team shutdown when the team's budget is exhausted, and SHALL publish a warning event when spending crosses the 80% threshold.

#### Scenario: Budget alert at 80% threshold publishes warning
- **WHEN** a `BudgetAlertEvent` is published with `Threshold >= 0.8`
- **THEN** the system SHALL publish a `TeamBudgetWarningEvent` with the team's current spent and budget amounts

#### Scenario: Budget alert below 80% is ignored
- **WHEN** a `BudgetAlertEvent` is published with `Threshold < 0.8`
- **THEN** the system SHALL NOT publish any warning event

#### Scenario: Budget exhausted triggers graceful shutdown
- **WHEN** a `BudgetExhaustedEvent` is published
- **THEN** the system SHALL call `coordinator.GracefulShutdown` with reason "budget exhausted"

### Requirement: On-chain escrow event types
The eventbus SHALL define typed events for all on-chain escrow lifecycle transitions.

#### Scenario: Deposit event carries transaction details
- **WHEN** an on-chain deposit occurs
- **THEN** the system SHALL publish an `EscrowOnChainDepositEvent` with EscrowID, DealID, Buyer, Amount, and TxHash fields

#### Scenario: Release event carries payout details
- **WHEN** an on-chain release occurs
- **THEN** the system SHALL publish an `EscrowOnChainReleaseEvent` with EscrowID, DealID, Seller, Amount, and TxHash fields

#### Scenario: Dispute event carries initiator info
- **WHEN** an on-chain dispute is raised
- **THEN** the system SHALL publish an `EscrowOnChainDisputeEvent` with EscrowID, DealID, Initiator, and TxHash fields

#### Scenario: Resolved event carries verdict
- **WHEN** an on-chain dispute is resolved
- **THEN** the system SHALL publish an `EscrowOnChainResolvedEvent` with EscrowID, DealID, SellerFavor, Amount, and TxHash fields

#### Scenario: Dangling event for stuck escrows
- **WHEN** an escrow has been stuck in Pending state beyond the configured timeout
- **THEN** the system SHALL publish an `EscrowDanglingEvent` with EscrowID, BuyerDID, SellerDID, Amount, PendingSince, and Action fields

### Requirement: Team lifecycle event types
The eventbus SHALL define typed events for team health monitoring, budget warnings, and graceful shutdown.

#### Scenario: Member unhealthy event
- **WHEN** a team member misses too many health pings
- **THEN** the system SHALL publish a `TeamMemberUnhealthyEvent` with TeamID, MemberDID, MemberName, MissedPings, and LastSeenAt fields

#### Scenario: Budget warning event
- **WHEN** a team's spending crosses a warning threshold
- **THEN** the system SHALL publish a `TeamBudgetWarningEvent` with TeamID, Threshold, Spent, and Budget fields

#### Scenario: Graceful shutdown event
- **WHEN** a team undergoes graceful shutdown
- **THEN** the system SHALL publish a `TeamGracefulShutdownEvent` with TeamID, Reason, BundlesCreated, and MembersSettled fields
