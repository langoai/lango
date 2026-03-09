package budget

import (
	"math/big"
	"time"
)

// BudgetStatus represents the current state of a task budget.
type BudgetStatus string

const (
	StatusActive    BudgetStatus = "active"
	StatusExhausted BudgetStatus = "exhausted"
	StatusClosed    BudgetStatus = "closed"
)

// TaskBudget tracks budget allocation and spending for a single task.
type TaskBudget struct {
	TaskID      string       `json:"taskId"`
	TotalBudget *big.Int     `json:"totalBudget"`
	Spent       *big.Int     `json:"spent"`
	Reserved    *big.Int     `json:"reserved"`
	Status      BudgetStatus `json:"status"`
	Progress    float64      `json:"progress"`
	Entries     []SpendEntry `json:"entries"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

// Remaining returns totalBudget - spent - reserved.
func (tb *TaskBudget) Remaining() *big.Int {
	r := new(big.Int).Sub(tb.TotalBudget, tb.Spent)
	return r.Sub(r, tb.Reserved)
}

// SpendEntry records a single spend event.
type SpendEntry struct {
	ID        string    `json:"id"`
	Amount    *big.Int  `json:"amount"`
	PeerDID   string    `json:"peerDid"`
	ToolName  string    `json:"toolName"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// BudgetReport is returned when a budget is closed.
type BudgetReport struct {
	TaskID      string        `json:"taskId"`
	TotalBudget *big.Int      `json:"totalBudget"`
	TotalSpent  *big.Int      `json:"totalSpent"`
	EntryCount  int           `json:"entryCount"`
	Duration    time.Duration `json:"duration"`
	Status      BudgetStatus  `json:"status"`
}
