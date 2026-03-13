package sentinel

import "time"

// AlertSeverity represents the severity of a security alert.
type AlertSeverity string

const (
	SeverityCritical AlertSeverity = "critical"
	SeverityHigh     AlertSeverity = "high"
	SeverityMedium   AlertSeverity = "medium"
	SeverityLow      AlertSeverity = "low"
)

// AlertMetadata holds structured metadata for alerts.
type AlertMetadata struct {
	Count           int    `json:"count,omitempty"`
	Window          string `json:"window,omitempty"`
	Amount          string `json:"amount,omitempty"`
	Threshold       string `json:"threshold,omitempty"`
	Elapsed         string `json:"elapsed,omitempty"`
	PreviousBalance string `json:"previousBalance,omitempty"`
	NewBalance      string `json:"newBalance,omitempty"`
}

// Alert represents a detected anomaly.
type Alert struct {
	ID           string        `json:"id"`
	Severity     AlertSeverity `json:"severity"`
	Type         string        `json:"type"`
	Message      string        `json:"message"`
	DealID       string        `json:"dealId,omitempty"`
	PeerDID      string        `json:"peerDid,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
	Acknowledged bool          `json:"acknowledged"`
	Metadata     AlertMetadata `json:"metadata,omitempty"`
}

// SentinelConfig holds detection thresholds.
type SentinelConfig struct {
	RapidCreationWindow   time.Duration `json:"rapidCreationWindow"`
	RapidCreationMax      int           `json:"rapidCreationMax"`
	LargeWithdrawalAmount string        `json:"largeWithdrawalAmount"`
	DisputeWindow         time.Duration `json:"disputeWindow"`
	DisputeMax            int           `json:"disputeMax"`
	WashTradeWindow       time.Duration `json:"washTradeWindow"`
}

// DefaultSentinelConfig returns sensible defaults.
func DefaultSentinelConfig() SentinelConfig {
	return SentinelConfig{
		RapidCreationWindow:   1 * time.Minute,
		RapidCreationMax:      5,
		LargeWithdrawalAmount: "10000000000", // 10,000 USDC (6 decimals)
		DisputeWindow:         1 * time.Hour,
		DisputeMax:            3,
		WashTradeWindow:       1 * time.Minute,
	}
}

// Detector interface for pattern detection.
type Detector interface {
	Name() string
	Analyze(event interface{}) *Alert
}
