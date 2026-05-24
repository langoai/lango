package postadjudicationstatus

import "strings"

const (
	ManualReplayActorFamilyOperator = "operator"
	ManualReplayActorFamilySystem   = "system"
	ManualReplayActorFamilyService  = "service"
	ManualReplayActorFamilyUnknown  = "unknown"
)

func ClassifyManualReplayActorFamily(actor string) string {
	normalized := strings.ToLower(strings.TrimSpace(actor))
	if normalized == "" {
		return ManualReplayActorFamilyUnknown
	}
	if hasAnyPrefix(normalized, "operator:", "user:", "human:") {
		return ManualReplayActorFamilyOperator
	}
	if hasAnyPrefix(normalized, "system:", "runtime:", "auto:", "worker:") {
		return ManualReplayActorFamilySystem
	}
	if hasAnyPrefix(normalized, "service:", "bridge:", "integration:", "bot:") {
		return ManualReplayActorFamilyService
	}
	return ManualReplayActorFamilyUnknown
}

func hasAnyPrefix(value string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}
