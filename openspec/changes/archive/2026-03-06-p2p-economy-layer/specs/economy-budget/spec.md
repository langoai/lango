## ADDED Requirements

### Requirement: Budget allocation
The system SHALL allow allocating a spending budget for a task identified by taskID, with a total amount in USDC smallest units (6 decimals). If no amount is provided, the system SHALL use the configured default max budget.

#### Scenario: Allocate with explicit amount
- **WHEN** Allocate is called with taskID "task-1" and amount 1000000
- **THEN** a TaskBudget is created with TotalBudget=1000000, Status=active, Spent=0

#### Scenario: Allocate with default max
- **WHEN** Allocate is called with taskID "task-1" and nil amount, and DefaultMax is "10.00"
- **THEN** a TaskBudget is created with TotalBudget=10000000

#### Scenario: Allocate duplicate
- **WHEN** Allocate is called with a taskID that already exists
- **THEN** the system SHALL return ErrBudgetExists

### Requirement: Spend checking with hard limit
The system SHALL verify that a proposed spend amount does not exceed the remaining budget when hard limit is enabled (default). The system SHALL reject spends against closed or exhausted budgets.

#### Scenario: Check within budget
- **WHEN** Check is called with amount 100000 on a budget with 1000000 remaining
- **THEN** no error is returned

#### Scenario: Check exceeds budget
- **WHEN** Check is called with amount exceeding remaining budget
- **THEN** ErrBudgetExceeded is returned

#### Scenario: Check on closed budget
- **WHEN** Check is called on a budget with status "closed"
- **THEN** ErrBudgetClosed is returned

### Requirement: Spend recording
The system SHALL record spend entries with amount, peerDID, toolName, and reason. The system SHALL auto-generate entry IDs and timestamps when not provided. When spending exhausts the budget, status SHALL transition to "exhausted".

#### Scenario: Record valid spend
- **WHEN** Record is called with amount 100000
- **THEN** Spent is updated, entry is appended with auto-generated ID

#### Scenario: Record exhausts budget
- **WHEN** Record is called with amount equal to remaining budget
- **THEN** Status transitions to "exhausted"

### Requirement: Budget reservation
The system SHALL support reserving amounts that temporarily reduce available budget. A release function SHALL be returned that restores the reserved amount. Release SHALL be idempotent.

#### Scenario: Reserve and release
- **WHEN** Reserve is called with 500000, then release is called
- **THEN** Reserved goes from 500000 back to 0

### Requirement: Threshold alerts
The system SHALL fire alert callbacks when the spent/total ratio crosses configured threshold percentages. Each threshold SHALL fire at most once per task.

#### Scenario: Alert at 50% threshold
- **WHEN** Spending reaches 50% with threshold [0.5, 0.8] configured
- **THEN** alertCallback is called with threshold=0.5

### Requirement: Budget close and report
The system SHALL finalize a budget by transitioning to "closed" status and returning a BudgetReport with total spent, entry count, and duration.

#### Scenario: Close active budget
- **WHEN** Close is called on an active budget with 2 entries totaling 500000
- **THEN** BudgetReport is returned with TotalSpent=500000, EntryCount=2, Status=closed
