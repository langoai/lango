package protocol

import (
	"context"
	"encoding/json"
)

// SchemaQueryRequest requests a peer's ontology schema bundle.
type SchemaQueryRequest struct {
	// RequestedTypes filters the response to specific entity type names.
	// Empty means return all types.
	RequestedTypes []string `json:"requestedTypes,omitempty"`

	// IncludePredicates controls whether predicate definitions are included.
	IncludePredicates bool `json:"includePredicates"`
}

// SchemaQueryResponse returns the peer's ontology schema bundle.
type SchemaQueryResponse struct {
	// Bundle is the JSON-encoded SchemaBundle from the ontology package.
	// Uses json.RawMessage to avoid importing internal/ontology.
	Bundle json.RawMessage `json:"bundle"`
}

// SchemaProposeRequest proposes ontology schema elements for import.
type SchemaProposeRequest struct {
	// Bundle is the JSON-encoded SchemaBundle to propose for import.
	// Uses json.RawMessage to avoid importing internal/ontology.
	Bundle json.RawMessage `json:"bundle"`

	// Reason describes why the schema elements are being proposed.
	Reason string `json:"reason,omitempty"`
}

// Ontology proposal action outcomes.
const (
	OntologyActionAccepted = "accepted"
	OntologyActionPartial  = "partial"
	OntologyActionRejected = "rejected"
)

// SchemaProposeResponse reports the result of a schema proposal.
type SchemaProposeResponse struct {
	// Action is the outcome: OntologyActionAccepted, OntologyActionPartial, or OntologyActionRejected.
	Action string `json:"action"`

	// Accepted lists the names of schema elements that were accepted.
	Accepted []string `json:"accepted,omitempty"`

	// Rejected lists the names of schema elements that were rejected.
	Rejected []string `json:"rejected,omitempty"`

	// Result is the JSON-encoded ImportResult from the ontology package.
	// Uses json.RawMessage to avoid importing internal/ontology.
	Result json.RawMessage `json:"result,omitempty"`
}

// OntologyHandler processes ontology exchange protocol messages.
// Implementations live in the bridge package to avoid import cycles.
type OntologyHandler interface {
	HandleSchemaQuery(ctx context.Context, peerDID string, req SchemaQueryRequest) (*SchemaQueryResponse, error)
	HandleSchemaPropose(ctx context.Context, peerDID string, req SchemaProposeRequest) (*SchemaProposeResponse, error)
}
