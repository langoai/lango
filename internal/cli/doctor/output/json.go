package output

import (
	"encoding/json"

	"github.com/langoai/lango/internal/cli/doctor/checks"
)

// JSONOutput represents the JSON output structure.
type JSONOutput struct {
	Results []JSONResult `json:"results"`
	Summary JSONSummary  `json:"summary"`
}

// JSONResult represents a single check result in JSON format.
type JSONResult struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	Fixable   bool   `json:"fixable,omitempty"`
	FixAction string `json:"fixAction,omitempty"`
}

// JSONSummary represents the summary in JSON format.
type JSONSummary struct {
	Passed   int `json:"passed"`
	Warnings int `json:"warnings"`
	Failed   int `json:"failed"`
	Skipped  int `json:"skipped"`
}

// JSONRenderer renders check results as JSON.
type JSONRenderer struct{}

// Render renders all results as JSON.
func (r *JSONRenderer) Render(summary checks.Summary) (string, error) {
	output := JSONOutput{
		Results: make([]JSONResult, len(summary.Results)),
		Summary: JSONSummary{
			Passed:   summary.Passed,
			Warnings: summary.Warnings,
			Failed:   summary.Failed,
			Skipped:  summary.Skipped,
		},
	}

	for i, result := range summary.Results {
		output.Results[i] = JSONResult{
			Name:      result.Name,
			Status:    result.Status.String(),
			Message:   result.Message,
			Details:   result.Details,
			Fixable:   result.Fixable,
			FixAction: result.FixAction,
		}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
