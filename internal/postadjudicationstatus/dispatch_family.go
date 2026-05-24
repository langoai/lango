package postadjudicationstatus

import "strings"

const (
	DispatchReferenceFamilyDispatch = "dispatch"
	DispatchReferenceFamilyQueue    = "queue"
	DispatchReferenceFamilyWorker   = "worker"
	DispatchReferenceFamilyBridge   = "bridge"
	DispatchReferenceFamilyWebhook  = "webhook"
	DispatchReferenceFamilyUnknown  = "unknown"
)

func ClassifyDispatchReferenceFamily(reference string) string {
	token := firstDispatchReferenceToken(reference)
	switch token {
	case "":
		return DispatchReferenceFamilyUnknown
	case DispatchReferenceFamilyDispatch,
		DispatchReferenceFamilyQueue,
		DispatchReferenceFamilyWorker,
		DispatchReferenceFamilyBridge,
		DispatchReferenceFamilyWebhook:
		return token
	case "job", "runner", "task":
		return DispatchReferenceFamilyWorker
	default:
		return token
	}
}

func DispatchReferenceFamilyOrder() []string {
	return []string{
		DispatchReferenceFamilyDispatch,
		DispatchReferenceFamilyQueue,
		DispatchReferenceFamilyWorker,
		DispatchReferenceFamilyBridge,
		DispatchReferenceFamilyWebhook,
		DispatchReferenceFamilyUnknown,
	}
}

func firstDispatchReferenceToken(reference string) string {
	normalized := strings.ToLower(strings.TrimSpace(reference))
	if normalized == "" {
		return ""
	}

	parts := strings.FieldsFunc(normalized, func(r rune) bool {
		switch r {
		case '-', '_', ':', '/', ' ', '\t', '\n':
			return true
		default:
			return false
		}
	})
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}
