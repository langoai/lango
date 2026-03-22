// Package provenance implements session provenance tracking including
// checkpoints, session trees, and attribution for the Lango P2P layer.
package provenance

import "time"

// CheckpointTrigger identifies what caused a checkpoint to be created.
type CheckpointTrigger string

const (
	TriggerManual       CheckpointTrigger = "manual"
	TriggerStepComplete CheckpointTrigger = "step_complete"
	TriggerPolicy       CheckpointTrigger = "policy_applied"
)

// Checkpoint is a thin metadata record marking a point in a RunLedger journal.
// It does NOT contain snapshot data — restoration replays the journal up to JournalSeq.
type Checkpoint struct {
	ID         string            `json:"id"`
	SessionKey string            `json:"session_key"`
	RunID      string            `json:"run_id,omitempty"`
	Label      string            `json:"label"`
	Trigger    CheckpointTrigger `json:"trigger"`
	JournalSeq int64             `json:"journal_seq"`
	GitRef     string            `json:"git_ref,omitempty"`
	TokensUsed *TokenSummary     `json:"tokens_used,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

// TokenSummary aggregates token usage across a session or checkpoint.
type TokenSummary struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
	TotalTokens  int64 `json:"total_tokens"`
}

// SessionStatus is the lifecycle status of a session node in the session tree.
type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusMerged    SessionStatus = "merged"
	SessionStatusDiscarded SessionStatus = "discarded"
	SessionStatusCompleted SessionStatus = "completed"
)

// SessionNode represents a node in the session tree hierarchy.
type SessionNode struct {
	SessionKey  string        `json:"session_key"`
	ParentKey   string        `json:"parent_key,omitempty"`
	AgentName   string        `json:"agent_name"`
	Goal        string        `json:"goal,omitempty"`
	RunID       string        `json:"run_id,omitempty"`
	WorkspaceID string        `json:"workspace_id,omitempty"`
	Depth       int           `json:"depth"`
	Status      SessionStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	ClosedAt    *time.Time    `json:"closed_at,omitempty"`
}

// AttributionSource identifies how an attribution record was captured.
type AttributionSource string

const (
	AttributionSourceWorkspaceMerge       AttributionSource = "workspace_merge"
	AttributionSourceWorkspaceBundlePush  AttributionSource = "workspace_bundle_push"
	AttributionSourceWorkspaceBundleApply AttributionSource = "workspace_bundle_apply"
	AttributionSourceSessionFork          AttributionSource = "session_fork"
	AttributionSourceSessionMerge         AttributionSource = "session_merge"
	AttributionSourceSessionDiscard       AttributionSource = "session_discard"
	AttributionSourceBundleImport         AttributionSource = "bundle_import"
)

// AuthorType identifies the kind of contributor.
type AuthorType string

const (
	AuthorHuman      AuthorType = "human"
	AuthorAgent      AuthorType = "agent"
	AuthorRemotePeer AuthorType = "remote_peer"
)

// Attribution records a coarse contribution by an author within a session.
type Attribution struct {
	ID           string            `json:"id"`
	SessionKey   string            `json:"session_key"`
	RunID        string            `json:"run_id,omitempty"`
	WorkspaceID  string            `json:"workspace_id,omitempty"`
	AuthorType   AuthorType        `json:"author_type"`
	AuthorID     string            `json:"author_id"`
	FilePath     string            `json:"file_path,omitempty"`
	CommitHash   string            `json:"commit_hash,omitempty"`
	StepID       string            `json:"step_id,omitempty"`
	Source       AttributionSource `json:"source,omitempty"`
	LinesAdded   int               `json:"lines_added"`
	LinesRemoved int               `json:"lines_removed"`
	TokensUsed   TokenSummary      `json:"tokens_used"`
	CreatedAt    time.Time         `json:"created_at"`
}

// AuthorStats summarizes an author's contributions.
type AuthorStats struct {
	AuthorType   AuthorType   `json:"author_type"`
	LinesAdded   int          `json:"lines_added"`
	LinesRemoved int          `json:"lines_removed"`
	TokensUsed   TokenSummary `json:"tokens_used"`
	FileCount    int          `json:"file_count"`
}

// FileStats summarizes contributions to a file.
type FileStats struct {
	LinesAdded   int `json:"lines_added"`
	LinesRemoved int `json:"lines_removed"`
	AuthorCount  int `json:"author_count"`
}

// AttributionReport aggregates attribution data for a session.
type AttributionReport struct {
	SessionKey  string                 `json:"session_key"`
	ByAuthor    map[string]AuthorStats `json:"by_author"`
	ByFile      map[string]FileStats   `json:"by_file"`
	TotalTokens TokenSummary           `json:"total_tokens"`
	Checkpoints int                    `json:"checkpoints"`
	GeneratedAt time.Time              `json:"generated_at"`
}

// RedactionLevel controls how much detail is exposed in provenance bundles.
type RedactionLevel string

const (
	RedactionNone    RedactionLevel = "none"
	RedactionContent RedactionLevel = "content"
	RedactionFull    RedactionLevel = "full"
)

// Valid reports whether r is a recognised redaction level.
func (r RedactionLevel) Valid() bool {
	switch r {
	case RedactionNone, RedactionContent, RedactionFull:
		return true
	}
	return false
}

// ProvenanceBundle is the portable container for provenance data.
type ProvenanceBundle struct {
	Version            string             `json:"version"`
	Checkpoints        []Checkpoint       `json:"checkpoints"`
	SessionTree        []SessionNode      `json:"session_tree,omitempty"`
	Attributions       []Attribution      `json:"attributions,omitempty"`
	Report             *AttributionReport `json:"report,omitempty"`
	SignerDID          string             `json:"signer_did,omitempty"`
	SignatureAlgorithm string             `json:"signature_algorithm,omitempty"`
	Signature          []byte             `json:"signature,omitempty"`
	RedactionLevel     RedactionLevel     `json:"redaction_level"`
}
