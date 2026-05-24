## MODIFIED Requirements

### Requirement: TaskBudget allocation and lifecycle
The system SHALL manage task budgets through a `Store` with `Allocate`, `Get`, `List`, `Update`, and `Delete` operations. Each task has exactly one TaskBudget identified by TaskID. Store methods that return or accept `TaskBudget` values SHALL isolate stored state from caller-owned mutable pointers and slices.

#### Scenario: Allocate a new task budget
- **WHEN** `Store.Allocate(taskID, total)` is called for a new task
- **THEN** a TaskBudget is created with TotalBudget=total, Spent=0, Reserved=0, Status="active"

#### Scenario: Allocate fails for existing task
- **WHEN** `Store.Allocate(taskID, total)` is called for an existing task
- **THEN** `ErrBudgetExists` is returned

#### Scenario: Get budget for unknown task
- **WHEN** `Store.Get(taskID)` is called for a non-existent task
- **THEN** `ErrBudgetNotFound` is returned

#### Scenario: Returned budgets are detached snapshots
- **WHEN** a caller mutates a TaskBudget returned by `Allocate`, `Get`, or `List`
- **THEN** the stored TaskBudget SHALL remain unchanged until `Store.Update` is called
- **AND** nested mutable fields such as `TotalBudget`, `Spent`, `Reserved`, and spend entry amounts SHALL NOT alias stored values

#### Scenario: Update stores a detached snapshot
- **WHEN** `Store.Update(budget)` succeeds
- **AND** the caller later mutates the same `budget` pointer or its nested mutable fields
- **THEN** the stored TaskBudget SHALL retain the state captured at update time
