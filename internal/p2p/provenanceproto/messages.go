package provenanceproto

import "encoding/json"

// ProtocolID is the libp2p protocol identifier for provenance bundles.
const ProtocolID = "/lango/provenance/1.0.0"

// RequestType identifies the provenance protocol request.
type RequestType string

const (
	RequestPushBundle  RequestType = "push_bundle"
	RequestFetchBundle RequestType = "fetch_bundle"
)

// Request is the envelope for provenance protocol messages.
type Request struct {
	Type    RequestType     `json:"type"`
	Token   string          `json:"token,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// PushBundlePayload carries a signed provenance bundle.
type PushBundlePayload struct {
	Bundle []byte `json:"bundle"`
}

// PushBundleResponse acknowledges provenance bundle storage.
type PushBundleResponse struct {
	Stored  bool   `json:"stored"`
	Message string `json:"message,omitempty"`
}

// FetchBundlePayload requests a signed provenance bundle for a session.
type FetchBundlePayload struct {
	SessionKey     string `json:"sessionKey"`
	RedactionLevel string `json:"redactionLevel"`
}

// FetchBundleResponse carries a signed provenance bundle from a peer.
type FetchBundleResponse struct {
	Bundle  []byte `json:"bundle,omitempty"`
	Message string `json:"message,omitempty"`
}
