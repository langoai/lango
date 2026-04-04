package approval

// DisplayTier determines how an approval request is rendered in the TUI.
type DisplayTier int

const (
	// TierInline renders a compact single-line strip.
	TierInline DisplayTier = 1
	// TierFullscreen renders a fullscreen overlay with diff preview.
	TierFullscreen DisplayTier = 2
)

// RiskIndicator provides a human-readable risk assessment for an approval request.
type RiskIndicator struct {
	Level string // "low", "moderate", "high", "critical"
	Label string // human-readable description
}

// ApprovalViewModel bridges ApprovalRequest data to TUI rendering.
type ApprovalViewModel struct {
	Request     ApprovalRequest
	Tier        DisplayTier
	Risk        RiskIndicator
	DiffContent string // unified diff for fs_edit/fs_write, empty otherwise
}

// fullscreenCategories are tool categories that trigger fullscreen approval.
var fullscreenCategories = map[string]bool{
	"filesystem": true,
	"automation": true,
}

// fullscreenActivities are tool activities that trigger fullscreen approval.
var fullscreenActivities = map[string]bool{
	"execute": true,
	"write":   true,
}

// ClassifyTier determines the display tier for an approval request based on
// safety level, category, and activity. Fullscreen is used when the tool is
// dangerous AND targets filesystem/automation or performs execute/write.
func ClassifyTier(safetyLevel, category, activity string) DisplayTier {
	if safetyLevel != "dangerous" {
		return TierInline
	}
	if fullscreenCategories[category] || fullscreenActivities[activity] {
		return TierFullscreen
	}
	return TierInline
}

// ComputeRisk returns a risk indicator based on safety level and category.
func ComputeRisk(safetyLevel, category string) RiskIndicator {
	switch safetyLevel {
	case "dangerous":
		switch category {
		case "filesystem":
			return RiskIndicator{Level: "critical", Label: "Modifies filesystem"}
		case "automation":
			return RiskIndicator{Level: "critical", Label: "Executes arbitrary code"}
		default:
			return RiskIndicator{Level: "high", Label: "Dangerous operation"}
		}
	case "moderate":
		return RiskIndicator{Level: "moderate", Label: "Creates or modifies resources"}
	default:
		return RiskIndicator{Level: "low", Label: "Read-only operation"}
	}
}

// NewViewModel creates an ApprovalViewModel from a request.
func NewViewModel(req ApprovalRequest) ApprovalViewModel {
	return ApprovalViewModel{
		Request: req,
		Tier:    ClassifyTier(req.SafetyLevel, req.Category, req.Activity),
		Risk:    ComputeRisk(req.SafetyLevel, req.Category),
	}
}
