package exportability

type SourceClass string

const (
	ClassPublic              SourceClass = "public"
	ClassUserExportable      SourceClass = "user-exportable"
	ClassPrivateConfidential SourceClass = "private-confidential"
)

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
