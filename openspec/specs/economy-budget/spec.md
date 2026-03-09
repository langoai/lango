## ADDED Requirements

### Requirement: Budget Guard interface for task-level spending control
The system SHALL provide a `Guard` interface in `internal/economy/budget/` with methods `Check`, `Record`, and `Reserve` to enforce per-task spending constraints.

#### Scenario: Check spending against budget
- **WHEN** `Guard.Check(taskID, amount)` is called for a task with remaining budget >= amount
- **THEN** nil is returned (spending is allowed)

#### Scenario: Check spending exceeds remaining budget
- **WHEN** `Guard.Check(taskID, amount)` is called and amount > Remaining()
- **THEN** an error is returned indicating budget would be exceeded

#### Scenario: Record a spend entry
- **WHEN** `Guard.Record(taskID, entry)` is called with a valid SpendEntry
- **THEN** the entry is appended to TaskBudget.Entries, Spent is increased by entry.Amount, and UpdatedAt is refreshed

#### Scenario: Reserve budget for a pending operation
- **WHEN** `Guard.Reserve(taskID, amount)` is called with amount <= Remaining()
- **THEN** Reserved is increased by amount and a releaseFunc is returned that decreases Reserved when called

#### Scenario: Reserve fails when insufficient budget
- **WHEN** `Guard.Reserve(taskID, amount)` is called and amount > Remaining()
- **THEN** an error is returned and no reservation is made

### Requirement: TaskBudget allocation and lifecycle
The system SHALL manage task budgets through a `Store` with `Allocate`, `Get`, `List`, `Update`, and `Delete` operations. Each task has exactly one TaskBudget identified by TaskID.

#### Scenario: Allocate a new task budget
- **WHEN** `Store.Allocate(taskID, total)` is called for a new task
- **THEN** a TaskBudget is created with TotalBudget=total, Spent=0, Reserved=0, Status="active"

#### Scenario: Allocate fails for existing task
- **WHEN** `Store.Allocate(taskID, total)` is called for an existing task
- **THEN** `ErrBudgetExists` is returned

#### Scenario: Get budget for unknown task
- **WHEN** `Store.Get(taskID)` is called for a non-existent task
- **THEN** `ErrBudgetNotFound` is returned

### Requirement: Budget remaining calculation
`TaskBudget.Remaining()` SHALL return `TotalBudget - Spent - Reserved`, representing the truly available budget.

#### Scenario: Remaining with no spending
- **WHEN** TotalBudget=10, Spent=0, Reserved=0
- **THEN** Remaining() returns 10

#### Scenario: Remaining with active reservation
- **WHEN** TotalBudget=10, Spent=3, Reserved=2
- **THEN** Remaining() returns 5

### Requirement: Budget status transitions
The system SHALL track budget status through three states: `active`, `exhausted`, and `closed`.

#### Scenario: Budget becomes exhausted
- **WHEN** Spent + Reserved >= TotalBudget after a Record or Reserve
- **THEN** Status transitions from "active" to "exhausted"

#### Scenario: Budget is closed manually
- **WHEN** the budget is finalized (task completed)
- **THEN** Status transitions to "closed" and a BudgetReport is generated

### Requirement: Threshold-based budget alerts
The system SHALL publish `BudgetAlertEvent` when spending crosses configured alert thresholds (e.g. 50%, 80%, 95% of TotalBudget).

#### Scenario: Spending crosses 80% threshold
- **WHEN** a Record causes Spent/TotalBudget to cross 0.8
- **THEN** a BudgetAlertEvent is published with threshold=0.8 and current progress

#### Scenario: Budget exhausted event
- **WHEN** Spent reaches TotalBudget
- **THEN** a BudgetExhaustedEvent is published with the TaskID and final BudgetReport

### Requirement: Hard limit enforcement
When `BudgetConfig.HardLimit` is true (default), the Guard SHALL reject any spend that would cause `Spent + amount > TotalBudget`. When false, spending is allowed with a warning event.

#### Scenario: Hard limit rejects overspend
- **WHEN** HardLimit=true and amount > Remaining()
- **THEN** Guard.Check returns an error

#### Scenario: Soft limit allows overspend with warning
- **WHEN** HardLimit=false and amount > Remaining()
- **THEN** Guard.Check returns nil but a BudgetAlertEvent is published

### Requirement: SpendEntry tracking
Each spend event SHALL be recorded as a `SpendEntry` with ID, Amount, PeerDID, ToolName, Reason, and Timestamp for audit purposes.

#### Scenario: SpendEntry records tool invocation payment
- **WHEN** a tool invocation is paid for
- **THEN** a SpendEntry is created with ToolName set to the invoked tool, PeerDID set to the provider, and Amount set to the payment

### Requirement: BudgetConfig defaults
The system SHALL use the following defaults from `config.BudgetConfig`:
- `DefaultMax`: "10.00" (USDC)
- `AlertThresholds`: [0.5, 0.8, 0.95]
- `HardLimit`: true

#### Scenario: Budget created with default config
- **WHEN** no explicit budget total is provided
- **THEN** DefaultMax ("10.00") is parsed and used as TotalBudget

### Requirement: Budget integration with wallet SpendingLimiter
The Guard SHALL consult `wallet.SpendingLimiter` before allowing spending, ensuring both task-level and wallet-level limits are respected.

#### Scenario: Task budget available but wallet limit exceeded
- **WHEN** Guard.Check passes for task budget but SpendingLimiter.Check returns an error
- **THEN** the spend is rejected with the wallet limit error

### Requirement: BudgetReport on close
When a budget is closed, the system SHALL produce a `BudgetReport` containing TaskID, TotalBudget, TotalSpent, EntryCount, Duration, and final Status.

#### Scenario: Close generates report
- **WHEN** a task budget is closed after 3 spend entries totaling 7.50 USDC over 2 hours
- **THEN** BudgetReport contains EntryCount=3, TotalSpent=7.50, Duration=2h, Status="closed"
