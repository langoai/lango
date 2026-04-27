package postadjudicationreplay

import (
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/receipts"
)

type ExecutionMode string

const (
	ExecutionModeManualRecovery ExecutionMode = "manual_recovery"
	ExecutionModeInline         ExecutionMode = "inline"
	ExecutionModeBackground     ExecutionMode = "background"
)

// ExecutionPolicy decides which follow-up substrate should run after
// canonical adjudication when explicit surface flags are absent.
type ExecutionPolicy struct {
	DefaultMode ExecutionMode
}

func DefaultExecutionPolicy() ExecutionPolicy {
	return ExecutionPolicy{DefaultMode: ExecutionModeManualRecovery}
}

func (p ExecutionPolicy) Resolve(autoExecute, backgroundExecute bool) (ExecutionMode, error) {
	if autoExecute && backgroundExecute {
		return "", fmt.Errorf("auto_execute and background_execute are mutually exclusive")
	}
	if autoExecute {
		return ExecutionModeInline, nil
	}
	if backgroundExecute {
		return ExecutionModeBackground, nil
	}

	switch p.DefaultMode {
	case "", ExecutionModeManualRecovery:
		return ExecutionModeManualRecovery, nil
	case ExecutionModeInline, ExecutionModeBackground:
		return p.DefaultMode, nil
	default:
		return "", fmt.Errorf("unsupported execution mode %q", p.DefaultMode)
	}
}

// RecoveryPolicy captures the canonical evidence semantics for replay and
// retry/dead-letter recovery.
type RecoveryPolicy struct {
	EvidenceSource              string
	RetryScheduledSubtype       string
	DeadLetteredSubtype         string
	ManualRetryRequestedSubtype string
}

func DefaultRecoveryPolicy() RecoveryPolicy {
	return RecoveryPolicy{
		EvidenceSource:              receipts.PostAdjudicationRecoveryEventSource,
		RetryScheduledSubtype:       receipts.PostAdjudicationRetryScheduledSubtype,
		DeadLetteredSubtype:         receipts.PostAdjudicationDeadLetteredSubtype,
		ManualRetryRequestedSubtype: receipts.PostAdjudicationManualRetryRequestedSubtype,
	}
}

func (p RecoveryPolicy) HasDeadLetterEvidence(events []receipts.ReceiptEvent) bool {
	for _, event := range events {
		if event.Source == p.EvidenceSource && event.Subtype == p.DeadLetteredSubtype {
			return true
		}
	}
	return false
}

func BuildBackgroundDispatchPrompt(
	outcome receipts.EscrowAdjudicationDecision,
	transactionReceiptID string,
	submissionReceiptID string,
	escrowReference string,
) string {
	toolName := "release_escrow_settlement"
	switch outcome {
	case receipts.EscrowAdjudicationRefund:
		toolName = "refund_escrow_settlement"
	}

	return fmt.Sprintf(
		"Execute the adjudicated escrow %s branch for transaction_receipt_id=%s.\nUse %s to perform the branch as a background follow-up.\nThe canonical adjudication is already recorded for submission_receipt_id=%s and escrow_reference=%s.\nDo not re-adjudicate.",
		outcome,
		strings.TrimSpace(transactionReceiptID),
		toolName,
		strings.TrimSpace(submissionReceiptID),
		strings.TrimSpace(escrowReference),
	)
}

func CanonicalRetryKey(transactionReceiptID string, outcome receipts.EscrowAdjudicationDecision) string {
	transactionReceiptID = strings.TrimSpace(transactionReceiptID)
	if transactionReceiptID == "" {
		return ""
	}
	if outcome != receipts.EscrowAdjudicationRelease && outcome != receipts.EscrowAdjudicationRefund {
		return ""
	}
	return transactionReceiptID + ":" + string(outcome)
}

func RetryKeyFromPrompt(prompt string) string {
	transactionReceiptID := ""
	outcome := receipts.EscrowAdjudicationDecision("")

	switch {
	case strings.Contains(prompt, "release_escrow_settlement"):
		outcome = receipts.EscrowAdjudicationRelease
	case strings.Contains(prompt, "refund_escrow_settlement"):
		outcome = receipts.EscrowAdjudicationRefund
	default:
		return ""
	}

	transactionReceiptID = extractPromptField(prompt, "transaction_receipt_id")

	return CanonicalRetryKey(transactionReceiptID, outcome)
}

func ParseRetryKey(retryKey string) (string, receipts.EscrowAdjudicationDecision, bool) {
	parts := strings.SplitN(strings.TrimSpace(retryKey), ":", 2)
	if len(parts) != 2 || parts[0] == "" {
		return "", "", false
	}

	outcome := receipts.EscrowAdjudicationDecision(parts[1])
	switch outcome {
	case receipts.EscrowAdjudicationRelease, receipts.EscrowAdjudicationRefund:
		return parts[0], outcome, true
	default:
		return "", "", false
	}
}

func extractPromptField(prompt, field string) string {
	idx := strings.Index(prompt, field+"=")
	if idx < 0 {
		return ""
	}

	fields := strings.Fields(strings.TrimSpace(prompt[idx+len(field)+1:]))
	if len(fields) == 0 {
		return ""
	}

	return strings.TrimRight(fields[0], ".,;)")
}
