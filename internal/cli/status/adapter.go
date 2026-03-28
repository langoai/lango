package status

import "github.com/langoai/lango/internal/types"

// FeatureStatusToFeatureInfo converts a types.FeatureStatus to a FeatureInfo.
func FeatureStatusToFeatureInfo(s types.FeatureStatus) FeatureInfo {
	detail := s.Reason
	if detail == "" && s.Suggestion != "" {
		detail = s.Suggestion
	}
	return FeatureInfo{
		Name:    s.Name,
		Enabled: s.Enabled,
		Detail:  detail,
	}
}
