package checks

import "github.com/langoai/lango/internal/types"

// FeatureStatusToDoctorResult converts a types.FeatureStatus to a doctor Result.
func FeatureStatusToDoctorResult(s types.FeatureStatus) Result {
	status := StatusPass
	if !s.Enabled && s.Reason != "" {
		status = StatusWarn
	} else if !s.Healthy {
		status = StatusFail
	} else if !s.Enabled {
		status = StatusSkip
	}

	message := s.Reason
	if message == "" && s.Suggestion != "" {
		message = s.Suggestion
	}
	if message == "" {
		message = s.Name + " is healthy"
	}

	return Result{
		Name:    s.Name,
		Status:  status,
		Message: message,
	}
}
