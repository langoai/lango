package postadjudicationstatus

import "strings"

const (
	DeadLetterReasonFamilyRetryExhausted   = "retry-exhausted"
	DeadLetterReasonFamilyPolicyBlocked    = "policy-blocked"
	DeadLetterReasonFamilyReceiptInvalid   = "receipt-invalid"
	DeadLetterReasonFamilyBackgroundFailed = "background-failed"
	DeadLetterReasonFamilyUnknown          = "unknown"
)

func ClassifyDeadLetterReasonFamily(reason string) string {
	normalized := strings.ToLower(strings.TrimSpace(reason))
	if normalized == "" {
		return DeadLetterReasonFamilyUnknown
	}
	if containsAny(normalized, "retry exhausted", "attempts exhausted", "max retry", "retry limit") {
		return DeadLetterReasonFamilyRetryExhausted
	}
	if containsAny(normalized, "policy", "gate denied", "blocked", "not allowed", "forbidden") {
		return DeadLetterReasonFamilyPolicyBlocked
	}
	if containsAny(
		normalized,
		"invalid receipt",
		"invalid adjudication",
		"invalid transaction",
		"receipt",
		"adjudication invalid",
		"adjudication missing",
		"transaction invalid",
		"transaction receipt",
		"evidence invalid",
	) {
		return DeadLetterReasonFamilyReceiptInvalid
	}
	if containsAny(normalized, "background", "dispatch", "worker", "task failed", "queue") {
		return DeadLetterReasonFamilyBackgroundFailed
	}
	return DeadLetterReasonFamilyUnknown
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
