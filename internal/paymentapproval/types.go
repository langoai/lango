package paymentapproval

type Decision string

const (
	DecisionApprove  Decision = "approve"
	DecisionReject   Decision = "reject"
	DecisionEscalate Decision = "escalate"
)

type SuggestedMode string

const (
	ModePrepay   SuggestedMode = "prepay"
	ModeEscrow   SuggestedMode = "escrow"
	ModeEscalate SuggestedMode = "escalate"
	ModeReject   SuggestedMode = "reject"
)

type AmountClass string

const (
	AmountLow      AmountClass = "low"
	AmountMedium   AmountClass = "medium"
	AmountHigh     AmountClass = "high"
	AmountCritical AmountClass = "critical"
)

type RiskClass string

const (
	RiskLow      RiskClass = "low"
	RiskMedium   RiskClass = "medium"
	RiskHigh     RiskClass = "high"
	RiskCritical RiskClass = "critical"
)

type TrustInput struct {
	Score           float64
	ScoreSource     string
	RecentRiskFlags []string
}

type BudgetPolicyContext struct {
	BudgetCap             string
	RemainingBudget       string
	UserMaxPrepay         string
	CounterpartyException string
	TransactionMode       string
}

type Input struct {
	Amount         string
	Counterparty   string
	RequestedScope string
	Trust          TrustInput
	Budget         BudgetPolicyContext
}

type Outcome struct {
	Decision      Decision      `json:"decision"`
	Reason        string        `json:"reason"`
	PolicyCode    string        `json:"policy_code,omitempty"`
	SuggestedMode SuggestedMode `json:"suggested_mode"`
	AmountClass   AmountClass   `json:"amount_class,omitempty"`
	RiskClass     RiskClass     `json:"risk_class,omitempty"`
	FailureDetail string        `json:"failure_detail,omitempty"`
}
