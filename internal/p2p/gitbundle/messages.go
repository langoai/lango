package gitbundle

import (
	"encoding/json"
	"time"
)

// ProtocolID is the libp2p protocol identifier for git bundle exchange.
const ProtocolID = "/lango/p2p-git/1.0.0"

// RequestType identifies git protocol request types.
type RequestType string

const (
	RequestPushBundle  RequestType = "push_bundle"
	RequestFetchByHash RequestType = "fetch_by_hash"
	RequestListCommits RequestType = "list_commits"
	RequestFindLeaves  RequestType = "find_leaves"
	RequestDiff        RequestType = "diff"
)

// Request is a git protocol request.
type Request struct {
	Type        RequestType     `json:"type"`
	WorkspaceID string          `json:"workspaceId"`
	Token       string          `json:"token"`
	Payload     json.RawMessage `json:"payload,omitempty"`
	Timestamp   time.Time       `json:"timestamp"`
}

// Response is a git protocol response.
type Response struct {
	Status string          `json:"status"` // "ok" or "error"
	Error  string          `json:"error,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}

// PushBundlePayload contains a git bundle for pushing.
type PushBundlePayload struct {
	Bundle    []byte `json:"bundle"`    // base64-encoded in JSON
	CommitMsg string `json:"commitMsg"` // description of the push
	SenderDID string `json:"senderDid"`
}

// FetchByHashPayload requests a bundle containing a specific commit.
type FetchByHashPayload struct {
	CommitHash string `json:"commitHash"`
}

// ListCommitsPayload requests a commit listing.
type ListCommitsPayload struct {
	Limit int `json:"limit,omitempty"`
}

// DiffPayload requests a diff between two commits.
type DiffPayload struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// PushBundleResponse is returned after a successful push.
type PushBundleResponse struct {
	Applied bool   `json:"applied"`
	Message string `json:"message,omitempty"`
}

// ListCommitsResponse contains commit information.
type ListCommitsResponse struct {
	Commits []CommitInfo `json:"commits"`
}

// FindLeavesResponse contains DAG leaf commit hashes.
type FindLeavesResponse struct {
	Leaves []string `json:"leaves"`
}

// DiffResponse contains a diff output.
type DiffResponse struct {
	Diff string `json:"diff"`
}
