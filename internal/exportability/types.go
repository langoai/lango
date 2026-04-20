package exportability

import (
	"fmt"
	"strings"
)

type SourceClass string

const (
	ClassPublic              SourceClass = "public"
	ClassUserExportable      SourceClass = "user-exportable"
	ClassPrivateConfidential SourceClass = "private-confidential"
)

// DefaultSourceClass is used when source class metadata is omitted.
const DefaultSourceClass = ClassPrivateConfidential

// Valid reports whether c is one of the supported source classes.
func (c SourceClass) Valid() bool {
	switch c {
	case ClassPublic, ClassUserExportable, ClassPrivateConfidential:
		return true
	default:
		return false
	}
}

// ParseSourceClass normalizes and validates a source class value.
// Empty values default to private-confidential for backwards compatibility.
func ParseSourceClass(value string) (SourceClass, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return DefaultSourceClass, nil
	}
	c := SourceClass(trimmed)
	if c.Valid() {
		return c, nil
	}
	return "", fmt.Errorf("invalid source class %q", value)
}

type DecisionStage string

const (
	StageDraft DecisionStage = "draft"
	StageFinal DecisionStage = "final"
)

type DecisionState string

const (
	StateExportable       DecisionState = "exportable"
	StateBlocked          DecisionState = "blocked"
	StateNeedsHumanReview DecisionState = "needs-human-review"
)

type Policy struct {
	Enabled bool
}

type SourceRef struct {
	AssetID    string
	AssetLabel string
	Class      SourceClass
}

type LineageSummary struct {
	AssetID    string      `json:"asset_id"`
	AssetLabel string      `json:"asset_label"`
	Class      SourceClass `json:"class"`
	Rule       string      `json:"rule"`
}

type Receipt struct {
	Stage       DecisionStage    `json:"stage"`
	State       DecisionState    `json:"state"`
	PolicyCode  string           `json:"policy_code"`
	Explanation string           `json:"explanation"`
	Lineage     []LineageSummary `json:"lineage"`
}
